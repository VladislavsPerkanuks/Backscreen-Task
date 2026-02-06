package fetcher

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// actual RSS feed response
const mockRSS = `<?xml version="1.0" encoding="utf-8"?>
<!-- generator="Joomla! - Open Source Content Management" -->
<?xml-stylesheet href="/plugins/system/jce/css/content.css?badb4208be409b1335b815dde676300e" type="text/css"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
	<channel>
		<title>Exchange Rates | Latvijas Banka</title>
		<description><![CDATA[ECB EXCHANGE RATES]]></description>
		<link>https://www.bank.lv/</link>
		<lastBuildDate>Fri, 06 Feb 2026 20:35:34 +0200</lastBuildDate>
		<generator>Joomla! - Open Source Content Management</generator>
		<atom:link rel="self" type="application/rss+xml" href="https://www.bank.lv/component/graph/?view=graph&amp;format=feed&amp;mode=exc&amp;ecb=true&amp;Itemid=150&amp;type=rss"/>
		<image>
			<url>https://www.bank.lv/images/latvijas_banka_smalllogo.gif</url>
			<title>Exchange Rates</title>
			<link>https://www.bank.lv/</link>
			<width>142</width>
			<height>48</height>
		</image>
		<language>lv-LV</language>
		<ttl>5</ttl>
		<item>
			<title>ECB EXCHANGE RATES.</title>
			<link>https://www.bank.lv/</link>
			<guid isPermaLink="false">https://www.bank.lv/#02.02</guid>
			<description><![CDATA[AUD 1.70420000 BRL 6.22220000 CAD 1.61570000 CHF 0.91990000 CNY 8.22100000 CZK 24.30500000 DKK 7.46910000 GBP 0.86580000 HKD 9.24840000 HUF 381.20000000 IDR 19890.00000000 ILS 3.67380000 INR 108.41700000 ISK 145.00000000 JPY 183.59000000 KRW 1719.74000000 MXN 20.61310000 MYR 4.66730000 NOK 11.46550000 NZD 1.97050000 PHP 69.75400000 PLN 4.21730000 RON 5.09600000 SEK 10.59350000 SGD 1.50550000 THB 37.43200000 TRY 51.49410000 USD 1.18400000 ZAR 18.97740000 ]]></description>
			<pubDate>Mon, 02 Feb 2026 02:00:00 +0200</pubDate>
		</item>
		<item>
			<title>ECB EXCHANGE RATES.</title>
			<link>https://www.bank.lv/</link>
			<guid isPermaLink="false">https://www.bank.lv/#03.02</guid>
			<description><![CDATA[AUD 1.68300000 BRL 6.16850000 CAD 1.61160000 CHF 0.91730000 CNY 8.18770000 CZK 24.31200000 DKK 7.46870000 GBP 0.86230000 HKD 9.22010000 HUF 380.40000000 IDR 19790.00000000 ILS 3.64190000 INR 106.37100000 ISK 145.00000000 JPY 183.92000000 KRW 1709.44000000 MXN 20.42450000 MYR 4.64070000 NOK 11.42200000 NZD 1.95320000 PHP 69.73200000 PLN 4.22050000 RON 5.09510000 SEK 10.54850000 SGD 1.49940000 THB 37.25000000 TRY 51.32460000 USD 1.18010000 ZAR 18.82180000 ]]></description>
			<pubDate>Tue, 03 Feb 2026 02:00:00 +0200</pubDate>
		</item>
	</channel>
</rss>`

func TestGetLatestRates(t *testing.T) {
	t.Parallel()

	latestDate, _ := time.Parse(time.RFC1123Z, "Tue, 03 Feb 2026 02:00:00 +0200")
	latestDate = latestDate.UTC()

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
				{Currency: "AUD", Rate: decimal.NewFromFloat(1.70420000), Date: latestDate},
				{Currency: "BRL", Rate: decimal.NewFromFloat(6.16850000), Date: latestDate},
				{Currency: "CAD", Rate: decimal.NewFromFloat(1.61160000), Date: latestDate},
				{Currency: "CHF", Rate: decimal.NewFromFloat(0.91730000), Date: latestDate},
				{Currency: "CNY", Rate: decimal.NewFromFloat(8.18770000), Date: latestDate},
				{Currency: "CZK", Rate: decimal.NewFromFloat(24.31200000), Date: latestDate},
				{Currency: "DKK", Rate: decimal.NewFromFloat(7.46870000), Date: latestDate},
				{Currency: "GBP", Rate: decimal.NewFromFloat(0.86230000), Date: latestDate},
				{Currency: "HKD", Rate: decimal.NewFromFloat(9.22010000), Date: latestDate},
				{Currency: "HUF", Rate: decimal.NewFromFloat(380.40000000), Date: latestDate},
				{Currency: "IDR", Rate: decimal.NewFromFloat(19790.00000000), Date: latestDate},
				{Currency: "ILS", Rate: decimal.NewFromFloat(3.64190000), Date: latestDate},
				{Currency: "INR", Rate: decimal.NewFromFloat(106.37100000), Date: latestDate},
				{Currency: "ISK", Rate: decimal.NewFromFloat(145.00000000), Date: latestDate},
				{Currency: "JPY", Rate: decimal.NewFromFloat(183.92000000), Date: latestDate},
				{Currency: "KRW", Rate: decimal.NewFromFloat(1709.44000000), Date: latestDate},
				{Currency: "MXN", Rate: decimal.NewFromFloat(20.42450000), Date: latestDate},
				{Currency: "MYR", Rate: decimal.NewFromFloat(4.64070000), Date: latestDate},
				{Currency: "NOK", Rate: decimal.NewFromFloat(11.42200000), Date: latestDate},
				{Currency: "NZD", Rate: decimal.NewFromFloat(1.95320000), Date: latestDate},
				{Currency: "PHP", Rate: decimal.NewFromFloat(69.73200000), Date: latestDate},
				{Currency: "PLN", Rate: decimal.NewFromFloat(4.22050000), Date: latestDate},
				{Currency: "RON", Rate: decimal.NewFromFloat(5.09510000), Date: latestDate},
				{Currency: "SEK", Rate: decimal.NewFromFloat(10.54850000), Date: latestDate},
				{Currency: "SGD", Rate: decimal.NewFromFloat(1.49940000), Date: latestDate},
				{Currency: "THB", Rate: decimal.NewFromFloat(37.25000000), Date: latestDate},
				{Currency: "TRY", Rate: decimal.NewFromFloat(51.32460000), Date: latestDate},
				{Currency: "USD", Rate: decimal.NewFromFloat(1.18010000), Date: latestDate},
				{Currency: "ZAR", Rate: decimal.NewFromFloat(18.82180000), Date: latestDate},
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

			fetcher := NewBankLatviaFetcher(slog.Default(), ts.Client(), ts.URL)

			rates, err := fetcher.GetLatestRates(t.Context())

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

	latestDate, _ := time.Parse(time.RFC1123Z, "Tue, 03 Feb 2026 02:00:00 +0200")
	latestDate = latestDate.UTC()

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
				Rate:     decimal.NewFromFloat(1.18010000),
				Date:     latestDate,
			},
		},
		{
			name:        "Success - Returns latest rate for GBP",
			currency:    "GBP",
			rssResponse: mockRSS,
			statusCode:  http.StatusOK,
			expectedRate: models.ExchangeRate{
				Currency: "GBP",
				Rate:     decimal.NewFromFloat(0.86230000),
				Date:     latestDate,
			},
		},
		{
			name:        "Currency not found",
			currency:    "EUR",
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

			fetcher := NewBankLatviaFetcher(slog.Default(), ts.Client(), ts.URL)

			rate, err := fetcher.GetLatestCurrencyRate(t.Context(), tt.currency)

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
