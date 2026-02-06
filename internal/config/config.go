package config

import (
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	MaxOpenConns int
	MaxIdleConns int
	ConnLifetime time.Duration
}

type ServerConfig struct {
	Port int
}

func Load(logger *slog.Logger) (*Config, error) {
	viper.SetEnvPrefix("CURRENCY_SERVICE")
	viper.AutomaticEnv()

	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 3306)
	viper.SetDefault("DB_USER", "currency")
	viper.SetDefault("DB_PASSWORD", "currency")
	viper.SetDefault("DB_NAME", "currency_service")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_LIFETIME", "5m")

	// Server defaults
	viper.SetDefault("SERVER_PORT", 8080)

	connLifetime, err := time.ParseDuration(viper.GetString("DB_CONN_LIFETIME"))
	if err != nil {
		logger.Warn("invalid DB_CONN_LIFETIME, using default",
			slog.String("value", viper.GetString("DB_CONN_LIFETIME")),
			slog.Any("error", err))
		connLifetime = 5 * time.Minute
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:         viper.GetString("DB_HOST"),
			Port:         viper.GetInt("DB_PORT"),
			User:         viper.GetString("DB_USER"),
			Password:     viper.GetString("DB_PASSWORD"),
			Name:         viper.GetString("DB_NAME"),
			MaxOpenConns: viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnLifetime: connLifetime,
		},
		Server: ServerConfig{
			Port: viper.GetInt("SERVER_PORT"),
		},
	}

	logger.Debug("configuration loaded",
		slog.String("db_host", cfg.Database.Host),
		slog.Int("db_port", cfg.Database.Port),
		slog.String("db_name", cfg.Database.Name),
		slog.Int("server_port", cfg.Server.Port))

	return cfg, nil
}
