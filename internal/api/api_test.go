package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
	"github.com/stretchr/testify/require"
)

type mockRateReader struct {
	latestRates     []models.ExchangeRate
	latestErr       error
	historicalRates []models.ExchangeRate
	historicalErr   error
}

func (m *mockRateReader) GetLatestRates(ctx context.Context) ([]models.ExchangeRate, error) {
	return m.latestRates, m.latestErr
}

func (m *mockRateReader) GetHistoricalRates(ctx context.Context, currency string) ([]models.ExchangeRate, error) {
	return m.historicalRates, m.historicalErr
}

func TestLatestRateHandler(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Minute) // Truncate to avoid precision issues
	tests := []struct {
		name           string
		mockRates      []models.ExchangeRate
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			mockRates: []models.ExchangeRate{
				{Currency: "USD", Rate: 1.1, Date: now},
				{Currency: "GBP", Rate: 99.9, Date: now},
			},
			expectedStatus: http.StatusOK,
			expectedBody: `
			{
				"rates": [
					{
						"currency": "USD",
						"rate": 1.1,
						"date": "` + now.Format(time.RFC3339) + `"
					},
					{
						"currency": "GBP",
						"rate": 99.9,
						"date": "` + now.Format(time.RFC3339) + `"
					}
				],
				"updated_at": "` + now.Format(time.RFC3339) + `"
			}`,
		},
		{
			name:           "Error - Fetch Failed",
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `
			{
				"error": "failed to fetch latest rates"
			}`,
		},
		{
			name:           "Error - No Rates",
			mockRates:      []models.ExchangeRate{},
			expectedStatus: http.StatusNotFound,
			expectedBody: `
			{
				"error": "no rates found"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockRateReader{
				latestRates: tt.mockRates,
				latestErr:   tt.mockErr,
			}
			api := NewAPI(mock)

			req := httptest.NewRequest(http.MethodGet, "/latest", nil)
			rr := httptest.NewRecorder()

			api.LatestRateHandler(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.NotNil(t, req.Body)
			require.JSONEq(t, tt.expectedBody, rr.Body.String())
		})
	}
}

func TestHistoryRateHandler(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Minute) // Truncate to avoid precision issues
	tests := []struct {
		name           string
		currency       string
		mockRates      []models.ExchangeRate
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Success",
			currency: "USD",
			mockRates: []models.ExchangeRate{
				{Currency: "USD", Rate: 1.1, Date: now},
				{Currency: "USD", Rate: 1.2, Date: now.Add(-24 * time.Hour)},
				{Currency: "USD", Rate: 1.3, Date: now.Add(-48 * time.Hour)},
			},
			expectedStatus: http.StatusOK,
			expectedBody: `
			{
				"currency": "USD",
				"history": [
					{
						"currency": "USD",
						"rate": 1.1,
						"date": "` + now.Format(time.RFC3339) + `"
					},
					{
						"currency": "USD",
						"rate": 1.2,
						"date": "` + now.Add(-24*time.Hour).Format(time.RFC3339) + `"
					},
					{
						"currency": "USD",
						"rate": 1.3,
						"date": "` + now.Add(-48*time.Hour).Format(time.RFC3339) + `"
					}
				]
			}
			`,
		},
		{
			name:           "Error - Invalid Currency Format",
			currency:       "US",
			expectedStatus: http.StatusBadRequest,
			expectedBody: `
			{
				"error": "invalid currency format"
			}`,
		},
		{
			name:           "Error - Fetch Failed",
			currency:       "USD",
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `
			{
				"error": "failed to fetch historical rates"
			}`,
		},
		{
			name:           "Error - No Rates Found",
			currency:       "GBP",
			mockRates:      []models.ExchangeRate{},
			expectedStatus: http.StatusNotFound,
			expectedBody: `
			{
				"error": "no rates found for currency: GBP"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockRateReader{
				historicalRates: tt.mockRates,
				historicalErr:   tt.mockErr,
			}
			api := NewAPI(mock)

			req := httptest.NewRequest(http.MethodGet, "/history/"+tt.currency, nil)
			req.SetPathValue("currency", tt.currency)

			rr := httptest.NewRecorder()

			api.HistoryRateHandler(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.NotNil(t, req.Body)
			require.JSONEq(t, tt.expectedBody, rr.Body.String())
		})
	}
}
