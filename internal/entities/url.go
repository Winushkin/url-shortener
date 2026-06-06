// Package entities содержит слой сущностей
package entities

import (
	"time"
)

type URL struct {
	ID           int       `json:"id"`
	LongURL      string    `json:"long_url"`
	ShortCode    string    `json:"short_code"`
	ClicksAmount int       `json:"click_amount"`
	CreatedAt    time.Time `json:"crated_at"`
}
