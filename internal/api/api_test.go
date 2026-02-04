package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockRateService struct {
	getLatestRatesFn     func() (map[string]float64, error)
	getHistoricalRatesFn func(currency string) ([]float64, error)
}

func (m *mockRateService) GetLatestRates() (map[string]float64, error) {
	return m.getLatestRatesFn()
}

func (m *mockRateService) GetHistoricalRates(currency string) ([]float64, error) {
	return m.getHistoricalRatesFn(currency)
}

func TestLatestRateHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockFn         func() (map[string]float64, error)
		expectedStatus int
		expectedBody   any
	}{
		{
			name: "Success",
			mockFn: func() (map[string]float64, error) {
				return map[string]float64{"USD": 1.0, "EUR": 0.85}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]any{"USD": 1.0, "EUR": 0.85},
		},
		{
			name: "Service Error",
			mockFn: func() (map[string]float64, error) {
				return nil, errors.New("provider failure")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]any{"error": "Internal Server Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockRateService{getLatestRatesFn: tt.mockFn}
			api := NewAPI(svc)

			req := httptest.NewRequest(http.MethodGet, "/latest", nil)
			w := httptest.NewRecorder()

			api.LatestRateHandler(w, req)

			require.Equal(t, tt.expectedStatus, w.Code, "unexpected status code")

			var resp map[string]any
			require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
			require.Equal(t, tt.expectedBody, resp)
		})
	}
}

func TestHistoryRateHandler(t *testing.T) {
	tests := []struct {
		name           string
		currency       string
		mockFn         func(string) ([]float64, error)
		expectedStatus int
		expectedBody   any
	}{
		{
			name:     "Success",
			currency: "USD",
			mockFn: func(c string) ([]float64, error) {
				return []float64{1.1, 1.2, 1.3}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []float64{1.1, 1.2, 1.3},
		},
		{
			name:           "Invalid Currency Length",
			currency:       "US",
			mockFn:         nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "invalid currency format: must be 3 characters"},
		},
		{
			name:     "Service Error",
			currency: "EUR",
			mockFn: func(c string) ([]float64, error) {
				return nil, errors.New("database failure")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "Internal Server Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockRateService{getHistoricalRatesFn: tt.mockFn}
			api := NewAPI(svc)

			req := httptest.NewRequest(http.MethodGet, "/history/"+tt.currency, nil)
			req.SetPathValue("currency", tt.currency)
			w := httptest.NewRecorder()

			api.HistoryRateHandler(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp []float64
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				require.Equal(t, tt.expectedBody, resp)

				return
			}

			var resp map[string]string
			require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
			require.Equal(t, tt.expectedBody, resp)
		})
	}
}
