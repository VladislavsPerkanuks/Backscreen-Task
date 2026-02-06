package fetcher

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
	"github.com/shopspring/decimal"
	"golang.org/x/net/html/charset"
)

// ----------------------- RSS feed structure based on actual Bank.lv feed -----------------------
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// ----------------------------------------------------------------------------------------------

type BankLatviaFetcher struct {
	logger *slog.Logger
	client *http.Client
	url    string
}

func NewBankLatviaFetcher(logger *slog.Logger, client *http.Client, url string) *BankLatviaFetcher {
	b := &BankLatviaFetcher{
		logger: logger,
		client: client,
		url:    url,
	}

	b.logger = b.logger.With(slog.String("fetcher", "BankLatviaFetcher"), slog.String("url", url))

	return b
}

const (
	dateLayout = time.RFC1123Z
)

var (
	ErrNoRatesFound = errors.New("no rates found")
	ErrRateNotFound = errors.New("rate not found for currency")
)

// GetAllRates returns all exchange rates from the feed grouped by date
func (b *BankLatviaFetcher) GetAllRates(ctx context.Context) ([][]models.ExchangeRate, error) {
	items, err := b.fetchRSS(ctx)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, ErrNoRatesFound
	}

	var allRates [][]models.ExchangeRate
	for _, item := range items {
		rates, err := b.parseRates(item.Description, b.parseDate(item.PubDate))
		if err != nil {
			b.logger.Warn("failed to parse rates for item", "date", item.PubDate, "err", err)

			continue
		}

		allRates = append(allRates, rates)
	}

	return allRates, nil
}

func (b *BankLatviaFetcher) GetCurrencyRates(ctx context.Context, currency string) ([]models.ExchangeRate, error) {
	items, err := b.fetchRSS(ctx)
	if err != nil {
		return nil, err
	}

	var currencyRates []models.ExchangeRate
	for _, item := range items {
		rates, err := b.parseRates(item.Description, b.parseDate(item.PubDate))
		if err != nil {
			continue
		}

		idx := slices.IndexFunc(rates, func(r models.ExchangeRate) bool {
			return r.Currency == currency
		})

		if idx != -1 {
			currencyRates = append(currencyRates, rates[idx])
		}
	}

	if len(currencyRates) == 0 {
		return nil, fmt.Errorf("%w '%s'", ErrRateNotFound, currency)
	}

	return currencyRates, nil
}

func (b *BankLatviaFetcher) fetchRSS(ctx context.Context) ([]Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch RSS feed: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck // We can't do much about a close error here

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel

	var rss RSS
	if err := decoder.Decode(&rss); err != nil {
		return nil, fmt.Errorf("decode XML: %w", err)
	}

	return rss.Channel.Items, nil
}

// parseRates parses the description field into a slice of models.ExchangeRate
// Format: "AUD 1.70010000 BRL 6.22330000 CAD 1.61200000 ..."
func (b *BankLatviaFetcher) parseRates(description string, date time.Time) ([]models.ExchangeRate, error) {
	var rates []models.ExchangeRate
	fields := strings.Fields(description)

	// Process pairs: code rate
	for i := 0; i < len(fields); i += 2 {
		currency, rateStr := fields[i], fields[i+1]

		rate, err := decimal.NewFromString(rateStr)
		if err != nil {
			b.logger.Error("failed to parse rate", "currency", currency, "rateStr", rateStr, "err", err)

			continue
		}

		rates = append(rates, models.ExchangeRate{
			Currency: currency,
			Rate:     rate,
			Date:     date,
		})
	}

	if len(rates) == 0 {
		return nil, ErrNoRatesFound
	}

	return rates, nil
}

// parseDate parses the pubDate string into time.Time
func (b *BankLatviaFetcher) parseDate(dateStr string) time.Time {
	t, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		b.logger.Error("failed to parse date", "dateStr", dateStr, "err", err)

		return time.Time{} // Return zero time on error
	}

	return t.UTC() // Ensure consistent timezone
}
