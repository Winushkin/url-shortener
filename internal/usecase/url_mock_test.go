package usecase_test

import (
	"context"
	"shortener/internal/entities"

	"github.com/stretchr/testify/mock"
)

type URLRepositoryMock struct {
	mock.Mock
}

func (m *URLRepositoryMock) InsertURL(ctx context.Context, url, shortCode string) error {
	args := m.Called(ctx, url, shortCode)
	return args.Error(0)
}

func (m *URLRepositoryMock) GetByShortCode(ctx context.Context, shortCode string) (*entities.URL, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.URL), args.Error(1)
}

func (m *URLRepositoryMock) IncrementClicks(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)
	return args.Error(0)
}
