package handler_test

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type URLUseCaseMock struct {
	mock.Mock
}

func (m *URLUseCaseMock) Shorten(ctx context.Context, longURL string) (string, error) {
	args := m.Called(ctx, longURL)
	return args.String(0), args.Error(1)
}

func (m *URLUseCaseMock) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	args := m.Called(ctx, shortCode)
	return args.String(0), args.Error(1)
}
