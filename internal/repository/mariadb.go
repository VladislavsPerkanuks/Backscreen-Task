package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/config"
	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

type MariaDBRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMariaDBRepository(cfg config.DatabaseConfig, logger *slog.Logger) (*MariaDBRepository, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	repoLogger := logger.With(
		slog.String("component", "repository"),
		slog.String("subsystem", "mariadb"),
	)

	return &MariaDBRepository{db: db, logger: repoLogger}, nil
}

func (r *MariaDBRepository) SaveRate(ctx context.Context, rate models.ExchangeRate) error {
	return r.SaveRates(ctx, []models.ExchangeRate{rate})
}

func (r *MariaDBRepository) SaveRates(ctx context.Context, rates []models.ExchangeRate) error {
	if len(rates) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(rates))
	args := make([]interface{}, 0, len(rates)*3)

	for _, rate := range rates {
		placeholders = append(placeholders, "(?, ?, ?)")
		args = append(args, rate.Currency, rate.Rate, rate.Date)
	}

	query := fmt.Sprintf(
		"INSERT INTO exchange_rates (currency, rate, date) VALUES %s ON DUPLICATE KEY UPDATE rate = VALUES(rate)",
		strings.Join(placeholders, ","),
	)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed bulk upsert", "count", len(rates), "err", err)
		return err
	}
	return nil
}

func (r *MariaDBRepository) GetLatestRates(ctx context.Context) ([]models.ExchangeRate, error) {
	query := `SELECT currency, rate, date FROM exchange_rates er1
              WHERE date = (SELECT MAX(date) FROM exchange_rates er2 WHERE er1.currency = er2.currency)
              ORDER BY currency`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("failed to fetch latest rates", slog.Any("error", err))

		return nil, fmt.Errorf("fetch latest rates: %w", err)
	}
	defer rows.Close() // nolint:errcheck // We can't do much about a close error here

	var rates []models.ExchangeRate
	for rows.Next() {
		var rate models.ExchangeRate
		if err := rows.Scan(&rate.Currency, &rate.Rate, &rate.Date); err != nil {
			r.logger.Error("failed to scan rate row", slog.Any("error", err))

			return nil, fmt.Errorf("scan rate: %w", err)
		}
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return rates, nil
}

func (r *MariaDBRepository) GetHistoricalRates(ctx context.Context, currency string) ([]models.ExchangeRate, error) {
	query := `SELECT currency, rate, date FROM exchange_rates
              WHERE currency = ? ORDER BY date ASC`

	rows, err := r.db.QueryContext(ctx, query, currency)
	if err != nil {
		r.logger.Error("failed to fetch historical rates",
			slog.String("currency", currency),
			slog.Any("error", err))

		return nil, fmt.Errorf("fetch historical rates for %s: %w", currency, err)
	}
	defer rows.Close() // nolint:errcheck // We can't do much about a close error here

	var rates []models.ExchangeRate
	for rows.Next() {
		var rate models.ExchangeRate
		if err := rows.Scan(&rate.Currency, &rate.Rate, &rate.Date); err != nil {
			r.logger.Error("failed to scan rate row", slog.Any("error", err))

			return nil, fmt.Errorf("scan rate: %w", err)
		}
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return rates, nil
}

func (r *MariaDBRepository) Close() error {
	return r.db.Close()
}
