package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// UserCounter describes the Count operation needed by [StatsHandler].
// Interfaz mínima (interface segregation): StatsHandler solo necesita Count,
// no todas las operaciones CRUD de UserRepository.
type UserCounter interface {
	Count(ctx context.Context) (int64, error)
}

// TaskCounter describes the Count operation needed by [StatsHandler].
type TaskCounter interface {
	Count(ctx context.Context) (int64, error)
}

// StatsHandler handles requests for aggregated resource counts.
type StatsHandler struct {
	users  UserCounter
	tasks  TaskCounter
	logger *slog.Logger
}

// NewStatsHandler returns a [StatsHandler] that queries users and tasks.
func NewStatsHandler(users UserCounter, tasks TaskCounter, logger *slog.Logger) *StatsHandler {
	return &StatsHandler{users: users, tasks: tasks, logger: logger}
}

// statsResponse is the JSON shape returned by [StatsHandler.GetStats].
type statsResponse struct {
	TotalUsers int64    `json:"total_users"`
	TotalTasks int64    `json:"total_tasks"`
	Errors     []string `json:"errors,omitempty"`
}

// countResult carries the count value or the error from a single count query.
type countResult struct {
	count int64
	err   error
}

// GetStats returns total user and task counts, queried in parallel.
//
// Patrón: fan-out con dos goroutines y canales buffereados de tamaño 1.
// El buffer=1 es crítico: si este handler retorna antes de recibir del canal
// (e.g., por ctx cancelado), las goroutines pueden escribir y terminar sin
// quedar bloqueadas para siempre (goroutine leak).
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userCh := make(chan countResult, 1)
	taskCh := make(chan countResult, 1)

	// Fan-out: lanzar ambas queries en paralelo
	go func() {
		n, err := h.users.Count(ctx)
		userCh <- countResult{n, err}
	}()

	go func() {
		n, err := h.tasks.Count(ctx)
		taskCh <- countResult{n, err}
	}()

	// Fan-in: recoger ambos resultados (el orden no importa porque cada canal
	// tiene su propio receptor; no se usa select aquí porque queremos AMBOS)
	userRes := <-userCh
	taskRes := <-taskCh

	resp := statsResponse{
		TotalUsers: userRes.count,
		TotalTasks: taskRes.count,
	}

	if userRes.err != nil {
		h.logger.Error("handler: failed to count users", "error", userRes.err)
		resp.Errors = append(resp.Errors, "error al contar usuarios")
	}
	if taskRes.err != nil {
		h.logger.Error("handler: failed to count tasks", "error", taskRes.err)
		resp.Errors = append(resp.Errors, "error al contar tareas")
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("handler: failed to encode stats response", "error", err)
	}
}
