package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
	"github.com/stretchr/testify/require"
)

// based on actual RSS feed structure
const mockRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <item>
            <pubDate>Mon, 01 Jan 2024 12:00:00 +0200</pubDate>
            <description>USD 1.10000000 EUR 0.90000000</description>
        </item>
        <item>
            <pubDate>Wed, 03 Jan 2024 12:00:00 +0200</pubDate>
            <description>USD 1.30000000 EUR 0.70000000</description>
        </item>
        <item>
            <pubDate>Tue, 02 Jan 2024 12:00:00 +0200</pubDate>
            <description>USD 1.20000000 EUR 0.80000000</description>
        </item>
    </channel>
</rss>`

func TestGetLatestRates(t *testing.T) {
	t.Parallel()

	latestDate, _ := time.Parse(time.RFC1123Z, "Wed, 03 Jan 2024 12:00:00 +0200")

	tests := []struct {
		name          string
		rssResponse   string
		statusCode    int
		expectedRates []models.ExchangeRate
		expectErr     string
	}{
		{
			name:        "Success - Picks Latest item by date",
			rssResponse: mockRSS,
			statusCode:  http.StatusOK,
			expectedRates: []models.ExchangeRate{
				{Currency: "USD", Rate: 1.3, Date: latestDate},
				{Currency: "EUR", Rate: 0.7, Date: latestDate},
			},
		},
		{
			name:        "RSS provider returns non-200 status",
			rssResponse: "error",
			statusCode:  http.StatusInternalServerError,
			expectErr:   "unexpected status code: 500",
		},
		{
			name:        "RSS provider returns malformed XML",
			rssResponse: "invalid xml",
			statusCode:  http.StatusOK,
			expectErr:   "decode XML: EOF",
		},
		{
			name:        "RSS provider returns empty items",
			rssResponse: `<rss><channel></channel></rss>`,
			statusCode:  http.StatusOK,
			expectErr:   "no rates found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = fmt.Fprint(w, tt.rssResponse)
			}))
			defer ts.Close()

			fetcher := NewBankLatviaFetcher(ts.Client(), ts.URL)

			rates, err := fetcher.GetLatestRates(context.Background())

			if tt.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectErr)
				return
			}

			require.NoError(t, err)
			require.ElementsMatch(t, tt.expectedRates, rates)
		})
	}
}

func TestGetLatestCurrencyRate(t *testing.T) {
	t.Parallel()

	latestDate, _ := time.Parse(time.RFC1123Z, "Wed, 03 Jan 2024 12:00:00 +0200")

	tests := []struct {
		name         string
		currency     string
		rssResponse  string
		statusCode   int
		expectedRate models.ExchangeRate
		expectErr    string
	}{
		{
			name:        "Success - Returns latest rate for USD",
			currency:    "USD",
			rssResponse: mockRSS,
			statusCode:  http.StatusOK,
			expectedRate: models.ExchangeRate{
				Currency: "USD",
				Rate:     1.3,
				Date:     latestDate,
			},
		},
		{
			name:        "Success - Returns latest rate for EUR",
			currency:    "EUR",
			rssResponse: mockRSS,
			statusCode:  http.StatusOK,
			expectedRate: models.ExchangeRate{
				Currency: "EUR",
				Rate:     0.7,
				Date:     latestDate,
			},
		},
		{
			name:        "Currency not found",
			currency:    "GBP",
			rssResponse: mockRSS,
			statusCode:  http.StatusOK,
			expectErr:   "rate not found for currency",
		},
		{
			name:        "Fetch error",
			currency:    "USD",
			rssResponse: "error",
			statusCode:  http.StatusInternalServerError,
			expectErr:   "unexpected status code: 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = fmt.Fprint(w, tt.rssResponse)
			}))
			defer ts.Close()

			fetcher := NewBankLatviaFetcher(ts.Client(), ts.URL)

			rate, err := fetcher.GetLatestCurrencyRate(context.Background(), tt.currency)

			if tt.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedRate, rate)
		})
	}
}
