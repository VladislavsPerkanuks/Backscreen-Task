package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/fetcher"
	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch latest currency rates from Bank.lv",
	RunE:  runFetchCommand,
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}

type ExchangeRateFetcher interface {
	GetLatestRates(ctx context.Context) ([]models.ExchangeRate, error)
	GetLatestCurrencyRate(ctx context.Context, currency string) (models.ExchangeRate, error)
}

type fetchResult struct {
	Rate models.ExchangeRate
	Err  error
}

func executeFetch(ctx context.Context, exchangeRateFetcher ExchangeRateFetcher, currencies []string) ([]models.ExchangeRate, error) {
	results := make(chan fetchResult, len(currencies))
	var wg sync.WaitGroup

	for _, curr := range currencies {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()

			rate, err := exchangeRateFetcher.GetLatestCurrencyRate(ctx, c)
			results <- fetchResult{Rate: rate, Err: err}
		}(curr)
	}

	wg.Wait()
	close(results)

	rates := make([]models.ExchangeRate, 0, len(currencies))
	errs := make([]error, 0, len(currencies))

	for res := range results {
		if res.Err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", res.Rate.Currency, res.Err))
			continue
		}

		rates = append(rates, res.Rate)
	}

	return rates, errors.Join(errs...)
}

func runFetchCommand(cmd *cobra.Command, args []string) error {
	fetcherSvc := fetcher.NewBankLatviaFetcher(&http.Client{Timeout: 30 * time.Second}, "https://www.bank.lv/vk/ecb_rss.xml")

	ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
	defer cancel()

	currencies := []string{"USD", "GBP", "JPY"}
	rates, err := executeFetch(ctx, fetcherSvc, currencies)
	if err != nil {
		return fmt.Errorf("failed to fetch rates: %w", err)
	}

	for _, rate := range rates {
		fmt.Printf("Currency: %s, Rate: %f\n", rate.Currency, rate.Rate)
	}

	return nil
}
