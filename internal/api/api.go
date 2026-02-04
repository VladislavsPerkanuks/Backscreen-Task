package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type RateService interface {
	GetLatestRates() (map[string]float64, error)
	GetHistoricalRates(currency string) ([]float64, error)
}

type API struct {
	rateService RateService
}

func NewAPI(rateService RateService) *API {
	return &API{rateService: rateService}
}

func (a *API) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("json encode failed", "err", err)
		// already sent headers → can't change status, just log
	}
}

func (a *API) errorResponse(w http.ResponseWriter, err error, userMsg string) {
	slog.Error("handler error", "err", err)
	a.jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": userMsg})
}

func (a *API) LatestRateHandler(w http.ResponseWriter, r *http.Request) {
	latestRate, err := a.rateService.GetLatestRates()
	if err != nil {
		a.errorResponse(w, err, "Internal Server Error")

		return
	}

	a.jsonResponse(w, http.StatusOK, latestRate)
}

func (a *API) HistoryRateHandler(w http.ResponseWriter, r *http.Request) {
	currency := r.PathValue("currency")

	// Basic validation: ISO 4217 codes are 3 characters
	if len(currency) != 3 {
		a.jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid currency format: must be 3 characters"})

		return
	}

	historyRate, err := a.rateService.GetHistoricalRates(currency)
	if err != nil {
		a.errorResponse(w, err, "Internal Server Error")

		return
	}

	a.jsonResponse(w, http.StatusOK, historyRate)
}
