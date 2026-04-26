// Package middleware provides HTTP middleware for the API server.
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Logging returns middleware that logs the method, path, status code, and latency of each request.
// Acepta un [AsyncLogger] para que la escritura de logs no bloquee el goroutine del request.
func Logging(logger *AsyncLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			logger.Send(slog.LevelInfo, "request received",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			next.ServeHTTP(wrapped, r)

			logger.Send(slog.LevelInfo, "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"latency_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
