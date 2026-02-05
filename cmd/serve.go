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
	"github.com/VladislavsPerkanuks/Backscreen-Task/internal/middleware"
	"github.com/spf13/cobra"
)

var port int //

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the currency service HTTP server",
	Run:   runServer,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to listen on")
}

func runServer(cmd *cobra.Command, args []string) {
	apiController := api.NewAPI(nil) // TODO

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/rates/latest", apiController.LatestRateHandler)
	mux.HandleFunc("GET /api/v1/rates/history/{currency}", apiController.HistoryRateHandler)

	handler := middleware.LoggingMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: handler,
	}

	// Channel to listen for interrupt or terminate signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info(fmt.Sprintf("Server starting on :%d", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)

			os.Exit(1)
		}
	}()

	<-stop

	slog.Info("Server shutting down...")

	// Create a context with a timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}

	slog.Info("Server stopped")
}
