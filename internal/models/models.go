package models

import "time"

type ExchangeRate struct {
	ID       int64     `json:"id,omitempty"`
	Currency string    `json:"currency"`
	Rate     float64   `json:"rate"`
	Date     time.Time `json:"date"`
}
