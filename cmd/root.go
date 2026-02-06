package cmd

import (
	"log/slog"
	"os"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/config"
	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/repository"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "currency-service",
	Short: "A microservice for fetching and serving currency exchange rates",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo},
		),
	)

	cfg, err := config.Load(logger)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)

		os.Exit(1)
	}

	repo, err := repository.NewMariaDBRepository(cfg.Database, logger)
	if err != nil {
		logger.Error("Failed to initialize database repository", "error", err)

		os.Exit(1)
	}

	defer repo.Close() // nolint:errcheck // We can't do much about a close error here

	rootCmd.AddCommand(NewFetchCmd(logger, &cfg.Database, repo))
	rootCmd.AddCommand(NewServeCmd(logger, &cfg.Server, repo))

	err = rootCmd.Execute()
	if err != nil {
		logger.Error("Command execution failed", "error", err)

		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
