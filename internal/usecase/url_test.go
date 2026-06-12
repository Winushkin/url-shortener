package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"shortener/internal/entities"
	"shortener/internal/usecase"
	"testing"
	"time"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestURLUseCase_GetLongURL(t *testing.T) {
	ctx, err := logger.NewLoggerContext(context.Background(), false)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	tests := []struct {
		name        string
		shortCode   string
		setupMock   func(rm *URLRepositoryMock, pm *PublisherMock)
		wantLongURL string
		wantErr     bool
		asyncWait   time.Duration
	}{
		{
			name:      "Успешный редирект и асинхронный клик",
			shortCode: "go12",
			setupMock: func(rm *URLRepositoryMock, pm *PublisherMock) {
				rm.On("GetByShortCode", ctx, "go12").
					Return(&entities.URL{LongURL: "https://go.dev"}, nil)

				pm.On("PublishClick", mock.Anything, mock.AnythingOfType("ClickEvent")).
					Return(nil)
			},
			wantLongURL: "https://go.dev",
			wantErr:     false,
			asyncWait:   5 * time.Millisecond,
		},
		{
			name:      "Ошибка: Код не найден в базе данных",
			shortCode: "notfound",
			setupMock: func(m *URLRepositoryMock, pm *PublisherMock) {
				m.On("GetByShortCode", ctx, "notfound").
					Return(nil, errors.New("sql: no rows in result set"))
			},
			wantLongURL: "",
			wantErr:     true,
			asyncWait:   0,
		},
	}

	// Запускаем цикл по всем сценариям
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(URLRepositoryMock)
			mockPublisher := new(PublisherMock)
			tt.setupMock(mockRepo, mockPublisher)

			uc := usecase.NewURLUseCase(
				usecase.Dependencies{
					Repo:      mockRepo,
					Publisher: mockPublisher,
				},
			)

			res, err := uc.GetLongURL(ctx, tt.shortCode)

			if tt.asyncWait > 0 {
				time.Sleep(tt.asyncWait)
			}

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
