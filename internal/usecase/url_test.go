package usecase_test

import (
	"context"
	"errors"
	"shortener/internal/entities"
	"shortener/internal/usecase"
	"testing"
	"time"

	// "github.com/bwmarrin/snowflake"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestURLUseCase_GetLongURL(t *testing.T) {
	// Базовый контекст для вызова основного метода
	ctx := context.Background()

	// Описываем структуру одного тест-кейса
	tests := []struct {
		name        string
		shortCode   string
		setupMock   func(m *URLRepositoryMock)
		wantLongURL string
		wantErr     bool
		asyncWait   time.Duration // сколько подождать, чтобы фоновая горутина успела отработать
	}{
		{
			name:      "Успешный редирект и асинхронный клик",
			shortCode: "go12",
			setupMock: func(m *URLRepositoryMock) {
				// 1. Ожидаем синхронный вызов поиска в БД
				m.On("GetByShortCode", ctx, "go12").
					Return(&entities.URL{LongURL: "https://go.dev"}, nil)

				// 2. Ожидаем АСИНХРОННЫЙ вызов инкремента кликов.
				// Так как внутри go func используется context.Background(),
				// мы подставляем mock.Anything, потому что точный указатель на контекст мы не поймаем.
				m.On("IncrementClicks", mock.Anything, "go12").
					Return(nil)
			},
			wantLongURL: "https://go.dev",
			wantErr:     false,
			asyncWait:   5 * time.Millisecond, // 5 миллисекунд хватит Go, чтобы провернуть горутину
		},
		{
			name:      "Ошибка: Код не найден в базе данных",
			shortCode: "notfound",
			setupMock: func(m *URLRepositoryMock) {
				// Если запись не найдена, база вернет ошибку, а инкремент кликов вызываться не должен
				m.On("GetByShortCode", ctx, "notfound").
					Return(nil, errors.New("sql: no rows in result set"))
			},
			wantLongURL: "",
			wantErr:     true,
			asyncWait:   0, // горутина не запускается, ждать нечего
		},
	}

	// Запускаем цикл по всем сценариям
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Создаем чистый экземпляр мока для текущего сценария
			mockRepo := new(URLRepositoryMock)
			tt.setupMock(mockRepo)

			// 2. Инициализируем UseCase, передавая структуру зависимостей с моком
			uc := usecase.NewURLUseCase(usecase.Dependencies{
				Repo:       mockRepo,
				ClicksChan: make(chan string),
			})

			// 3. Вызываем тестируемый метод
			res, err := uc.GetLongURL(ctx, tt.shortCode)

			// 4. Даем фоновым горутинам (если они есть) время выполниться в памяти
			if tt.asyncWait > 0 {
				time.Sleep(tt.asyncWait)
			}

			// 5. Проверяем результаты с помощью assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantLongURL, res)
			}

			// 6. КРИТИЧЕСКИ ВАЖНО: Проверяем, что все методы `.On()` у мока
			// действительно были вызваны. Если горутина не запустилась, тест упадет здесь.
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestURLUseCase_Shorten(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx     context.Context
		longURL string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(m *URLRepositoryMock, longURL string)
		wantErr   bool
		err       error
	}{
		{
			name: "Успешное создание короткой ссылки",
			args: args{
				ctx:     ctx,
				longURL: "https://example.com",
			},
			setupMock: func(m *URLRepositoryMock, longURL string) {
				// 1. mock.Anything — для контекста
				// 2. longURL — точное совпадение по ссылке
				// 3. mock.Anything — так как shortCode генерируется случайно
				m.On("InsertURL", mock.Anything, longURL, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Ошибка при сохранении в базу данных",
			args: args{
				ctx:     ctx,
				longURL: "https://example.com",
			},
			setupMock: func(m *URLRepositoryMock, longURL string) {
				// Возвращаем ошибку при попытке вставить запись
				m.On("InsertURL", mock.Anything, longURL, mock.Anything).
					Return(errors.New("database connection failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Инициализируем мок-репозиторий
			mockRepo := new(URLRepositoryMock)

			// 2. Настраиваем поведение мока для текущего тест-кейса
			tt.setupMock(mockRepo, tt.args.longURL)

			node, err := snowflake.NewNode(1)
			require.NoError(t, err)

			// 3. Создаем UseCase и передаем туда наш мок
			// Передаем также инициализированный канал кликов, чтобы избежать зависаний (как в GetLongURL)
			cfg := usecase.Dependencies{
				Repo: mockRepo,
				Node: node,
			}
			uc := usecase.NewURLUseCase(cfg)

			// 4. Вызываем тестируемый метод
			shortCode, err := uc.Shorten(tt.args.ctx, tt.args.longURL)

			// 5. Проверяем результаты с помощью assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, shortCode) // При ошибке короткий код должен быть пустым
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, shortCode) // Код должен успешно сгенерироваться
				assert.Len(t, shortCode, 11)  // Например, если ваша длина кода всегда 6 символов
			}

			// 6. Проверяем, что все ожидаемые методы мока были вызваны
			mockRepo.AssertExpectations(t)
		})
	}
}
