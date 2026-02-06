package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

type contextKey string

const (
	requestIDKey contextKey = "reqID"
)

func LoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate or propagate request ID
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
			r.Header.Set("X-Request-ID", reqID)
		}

		// Attach to context & logger
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		logger := logger.With("reqID", reqID)

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		logger.Info("request started",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		defer func() {
			duration := time.Since(start)

			attrs := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Duration("duration", duration),
			}

			if rw.status >= 500 {
				logger.Error("request failed", attrs...)
			} else if rw.status >= 400 {
				logger.Warn("request client error", attrs...)
			} else {
				logger.Info("request completed", attrs...)
			}
		}()

		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}
