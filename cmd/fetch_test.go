package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

type mockFetcher struct {
	latestRates []models.ExchangeRate
}

func (m *mockFetcher) GetAllRates(ctx context.Context) ([][]models.ExchangeRate, error) {
	return [][]models.ExchangeRate{m.latestRates}, nil
}

func (m *mockFetcher) GetCurrencyRates(ctx context.Context, currency string) ([]models.ExchangeRate, error) {
	var rates []models.ExchangeRate
	for _, rate := range m.latestRates {
		if rate.Currency == currency {
			rates = append(rates, rate)
		}
	}

	return rates, nil
}

func (m *mockFetcher) GetHistoricalRates(ctx context.Context, currency string) ([]models.ExchangeRate, error) {
	return m.latestRates, nil
}

func (m *mockFetcher) SaveRate(ctx context.Context, rate models.ExchangeRate) error {
	return nil
}

func (m *mockFetcher) SaveRates(ctx context.Context, rates []models.ExchangeRate) error {
	return nil
}

func TestExecuteFetch(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Minute) // Truncate to avoid precision issues

	tests := []struct {
		name        string
		currencies  []string
		mockFetcher *mockFetcher
		expected    []models.ExchangeRate
	}{
		{
			name:       "successfully fetches rates",
			currencies: []string{"USD", "EUR", "GBP", "JPY", "AUD", "CAD", "CHF", "CNY", "SEK", "NZD"},
			mockFetcher: &mockFetcher{
				latestRates: []models.ExchangeRate{
					{Currency: "USD", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "EUR", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "GBP", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "JPY", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "AUD", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "CAD", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "CHF", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "CNY", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "SEK", Rate: decimal.NewFromFloat(1.15), Date: now},
					{Currency: "NZD", Rate: decimal.NewFromFloat(1.15), Date: now},
				},
			},
			expected: []models.ExchangeRate{
				{
					Currency: "USD", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "EUR", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "GBP", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "JPY", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "AUD", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "CAD", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "CHF", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "CNY", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "SEK", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
				{
					Currency: "NZD", Rate: decimal.NewFromFloat(1.15),
					Date: now,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			rates, err := executeFetch(ctx, tt.mockFetcher, tt.mockFetcher, tt.currencies)
			require.NoError(t, err)

			require.ElementsMatch(t, tt.expected, rates)
		})
	}
}
