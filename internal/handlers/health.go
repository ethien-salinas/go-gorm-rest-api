package handlers

import (
	"context"
	"net/http"
)

// Pinger is satisfied by any type that can check database connectivity.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// NewHealthHandler returns an [http.HandlerFunc] that responds 200 OK when the database is reachable.
//
//	@Summary		Health check
//	@Description	Responde 200 si la base de datos está accesible, 503 si no.
//	@Tags			general
//	@Success		200
//	@Failure		503
//	@Router			/health [get]
func NewHealthHandler(p Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := p.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
