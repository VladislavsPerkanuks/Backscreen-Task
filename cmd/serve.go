package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/rates/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Latest rates endpoint")) // nolint:errcheck // will be replaced soon
	})

	mux.HandleFunc("GET /api/v1/rates/history/{currency}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Historical rates endpoint for currency")) // nolint:errcheck // will be replaced soon
	})

	slog.Info(fmt.Sprintf("Server starting on :%d", port))
	if err := http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil {
		slog.Error("Server failed to start", "error", err)
	}
}
