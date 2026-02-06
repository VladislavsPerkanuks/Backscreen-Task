package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type ExchangeRate struct {
	ID       int64           `json:"id,omitempty"`
	Currency string          `json:"currency"`
	Rate     decimal.Decimal `json:"rate"`
	Date     time.Time       `json:"date"`
}
