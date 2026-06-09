package handler_test

import (
	"shortener/internal/http/handler"
	"testing"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"valid URL", "https://example.com", true},
		{"valid URL", "ftp://192.168.1.1", true},
		{"invalid URL", "invalid-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handler.IsValidURL(tt.url); got != tt.want {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}
