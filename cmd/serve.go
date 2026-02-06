package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/api"
	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/config"
	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewServeCmd(logger *slog.Logger, cfg *config.ServerConfig, rateReader api.RateReader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the currency service HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			apiController := api.NewAPI(logger, rateReader)

			mux := http.NewServeMux()

			mux.HandleFunc("GET /api/v1/rates/latest", apiController.LatestRateHandler)
			mux.HandleFunc("GET /api/v1/rates/history/{currency}", apiController.HistoryRateHandler)

			handler := middleware.LoggingMiddleware(logger, mux)

			server := &http.Server{
				Addr:         ":" + strconv.Itoa(cfg.Port),
				Handler:      handler,
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			}

			// Channel to listen for interrupt or terminate signals
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

			go func() {
				logger.Info(fmt.Sprintf("Server starting on :%d", cfg.Port))
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {

					logger.Error("Server failed to start", "error", err)

					os.Exit(1)
				}
			}()

			<-stop

			logger.Info("Server shutting down...")

			// Create a context with a timeout for the shutdown process
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				logger.Error("Server shutdown failed", "error", err)
			}

			logger.Info("Server stopped")
		},
	}

	// ----------------------- Flags -----------------------

	cmd.Flags().IntVarP(&cfg.Port, "port", "p", cfg.Port, "port to listen on")

	if err := viper.BindPFlag("server_port", cmd.Flags().Lookup("port")); err != nil {
		logger.Error("Failed to bind server port flag", "error", err)
	}

	return cmd
}
