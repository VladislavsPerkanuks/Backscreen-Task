package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
	apiController := api.NewAPI(nil)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/rates/latest", apiController.LatestRateHandler)
	mux.HandleFunc("GET /api/v1/rates/history/{currency}", apiController.HistoryRateHandler)

	handler := middleware.LoggingMiddleware(mux)

	slog.Info(fmt.Sprintf("Server starting on :%d", port))
	if err := http.ListenAndServe(":"+strconv.Itoa(port), handler); err != nil {
		slog.Error("Server failed to start", "error", err)
	}
}
