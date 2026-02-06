package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
)

type API struct {
	logger     *slog.Logger
	rateReader RateReader
}

func NewAPI(logger *slog.Logger, rateReader RateReader) *API {
	a := &API{
		logger:     logger,
		rateReader: rateReader,
	}

	a.logger = a.logger.With(slog.String("component", "API"))

	return a
}

// LatestRatesResponse represents the API response for latest rates
type LatestRatesResponse struct {
	Rates []models.ExchangeRate `json:"rates"`
}

// HistoricalRatesResponse represents the API response for historical rates
type HistoricalRatesResponse struct {
	Currency string                `json:"currency"`
	History  []models.ExchangeRate `json:"history"`
}

// RateReader defines the interface for reading exchange rates
type RateReader interface {
	GetLatestRates(ctx context.Context) ([]models.ExchangeRate, error)
	GetHistoricalRates(ctx context.Context, currency string) ([]models.ExchangeRate, error)
}

func (a *API) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		a.logger.Error("json encode failed", "err", err)

		return
	}
}

func (a *API) errorResponse(w http.ResponseWriter, status int, err error, userMsg string) {
	a.logger.Error("handler error", slog.Any("error", err), slog.Int("status", status))
	a.jsonResponse(w, status, map[string]string{"error": userMsg})
}

func (a *API) LatestRateHandler(w http.ResponseWriter, r *http.Request) {
	rates, err := a.rateReader.GetLatestRates(r.Context())
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Errorf(
			"get latest rates: %w", err,
		), "failed to fetch latest rates")

		return
	}

	if len(rates) == 0 {
		a.errorResponse(w, http.StatusNotFound, errors.New("no rates found"), "no rates found")
		return
	}

	a.jsonResponse(w, http.StatusOK, LatestRatesResponse{
		Rates: rates,
	})
}

func (a *API) HistoryRateHandler(w http.ResponseWriter, r *http.Request) {
	currency := r.PathValue("currency")

	// ISO 4217
	if len(currency) != 3 {
		a.errorResponse(w, http.StatusBadRequest, errors.New("invalid currency format: must be 3 characters"), "invalid currency format")

		return
	}

	rates, err := a.rateReader.GetHistoricalRates(r.Context(), currency)
	if err != nil {
		a.errorResponse(
			w,
			http.StatusInternalServerError,
			fmt.Errorf("get historical rates: %w", err),
			"failed to fetch historical rates")

		return
	}

	if len(rates) == 0 {
		a.errorResponse(
			w,
			http.StatusNotFound,
			errors.New("no rates found"),
			"no rates found for currency: "+currency)

		return
	}

	a.jsonResponse(w, http.StatusOK, HistoricalRatesResponse{
		Currency: currency,
		History:  rates,
	})
}
