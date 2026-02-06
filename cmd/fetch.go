package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/fetcher"
	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ExchangeRateFetcher interface {
	GetAllRates(ctx context.Context) ([][]models.ExchangeRate, error)
	GetCurrencyRates(ctx context.Context, currency string) ([]models.ExchangeRate, error)
}

type ExchangeRateWriter interface {
	SaveRate(ctx context.Context, rate models.ExchangeRate) error
	SaveRates(ctx context.Context, rates []models.ExchangeRate) error
}

type fetchResult struct {
	Rates []models.ExchangeRate
	Err   error
}

func NewFetchCmd(logger *slog.Logger, writerSvc ExchangeRateWriter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch latest currency rates from Bank.lv",
		RunE: func(cmd *cobra.Command, args []string) error {
			fetcherSvc := fetcher.NewBankLatviaFetcher(logger, &http.Client{Timeout: 30 * time.Second}, "https://www.bank.lv/vk/ecb_rss.xml")

			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
			defer cancel()

			rates, err := executeFetch(ctx, fetcherSvc, writerSvc, viper.GetStringSlice("currencies"))
			if err != nil {
				return fmt.Errorf("failed to fetch rates: %w", err)
			}

			for _, rate := range rates {
				fmt.Printf("Currency: %s, Rate: %s, Date: %s\n", rate.Currency, rate.Rate, rate.Date.Format(time.DateOnly))
			}

			return nil
		},
	}

	// ------------ Flags --------------------

	cmd.Flags().StringSliceP("currencies", "c", []string{"USD", "GBP", "JPY"}, "Comma-separated list of currency codes to fetch")

	if err := viper.BindPFlag("currencies", cmd.Flags().Lookup("currencies")); err != nil {
		logger.Error("bind flag failed", "flag", "currencies", "error", err)
	}

	return cmd
}

func executeFetch(
	ctx context.Context,
	exchangeRateFetcher ExchangeRateFetcher,
	rateWriter ExchangeRateWriter,
	currencies []string,
) ([]models.ExchangeRate, error) {
	results := make(chan fetchResult, len(currencies))
	var wg sync.WaitGroup

	for _, curr := range currencies {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()

			rates, err := exchangeRateFetcher.GetCurrencyRates(ctx, c)
			results <- fetchResult{Rates: rates, Err: err}
		}(curr)
	}

	wg.Wait()
	close(results)

	var allRates []models.ExchangeRate
	var errs []error

	for res := range results {
		if res.Err != nil {
			errs = append(errs, res.Err)
			continue
		}

		allRates = append(allRates, res.Rates...)
	}

	if err := rateWriter.SaveRates(ctx, allRates); err != nil {
		errs = append(errs, fmt.Errorf("failed to save rates: %w", err))
	}

	return allRates, errors.Join(errs...)
}
