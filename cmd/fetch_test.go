package cmd

import (
	"context"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
	"github.com/stretchr/testify/require"
)

type mockFetcher struct {
	latestRates []models.ExchangeRate
}

func (m *mockFetcher) GetLatestRates(ctx context.Context) ([]models.ExchangeRate, error) {
	return m.latestRates, nil
}

func (m *mockFetcher) GetLatestCurrencyRate(ctx context.Context, currency string) (models.ExchangeRate, error) {
	// wait for a short time to simulate network delay
	time.Sleep(time.Duration(rand.IntN(500)) * time.Millisecond)
	idx := slices.IndexFunc(m.latestRates, func(rate models.ExchangeRate) bool {
		return rate.Currency == currency
	})

	return m.latestRates[idx], nil
}

func (m *mockFetcher) GetHistoricalRates(ctx context.Context, currency string) ([]models.ExchangeRate, error) {
	return m.latestRates, nil
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
					{Currency: "USD", Rate: 1.15, Date: now},
					{Currency: "EUR", Rate: 1.15, Date: now},
					{Currency: "GBP", Rate: 1.15, Date: now},
					{Currency: "JPY", Rate: 1.15, Date: now},
					{Currency: "AUD", Rate: 1.15, Date: now},
					{Currency: "CAD", Rate: 1.15, Date: now},
					{Currency: "CHF", Rate: 1.15, Date: now},
					{Currency: "CNY", Rate: 1.15, Date: now},
					{Currency: "SEK", Rate: 1.15, Date: now},
					{Currency: "NZD", Rate: 1.15, Date: now},
				},
			},
			expected: []models.ExchangeRate{
				{
					Currency: "USD", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "EUR", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "GBP", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "JPY", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "AUD", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "CAD", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "CHF", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "CNY", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "SEK", Rate: 1.15,
					Date: now,
				},
				{
					Currency: "NZD", Rate: 1.15,
					Date: now,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			rates, err := executeFetch(ctx, tt.mockFetcher, tt.currencies)
			require.NoError(t, err)

			require.ElementsMatch(t, tt.expected, rates)
		})
	}
}
