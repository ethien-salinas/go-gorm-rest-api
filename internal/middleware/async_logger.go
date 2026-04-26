package middleware

import (
	"context"
	"log/slog"
)

// LogEntry holds a single log message with its level and key-value args.
type LogEntry struct {
	level slog.Level
	msg   string
	args  []any
}

// AsyncLogger wraps [slog.Logger] with a buffered channel so that callers
// write without blocking on IO. A single goroutine worker drains the channel.
type AsyncLogger struct {
	ch     chan LogEntry
	logger *slog.Logger
	done   chan struct{}
}

// NewAsyncLogger creates an [AsyncLogger] backed by logger with a channel
// buffer of bufSize entries. Call [AsyncLogger.Start] before using it.
func NewAsyncLogger(logger *slog.Logger, bufSize int) *AsyncLogger {
	return &AsyncLogger{
		ch:     make(chan LogEntry, bufSize),
		logger: logger,
		done:   make(chan struct{}),
	}
}

// Start launches the goroutine worker that drains the channel.
// Each call to Start must be paired with exactly one call to [AsyncLogger.Stop].
func (a *AsyncLogger) Start() {
	go func() {
		defer close(a.done)
		// range sobre el canal: procesa entradas hasta que el canal se cierre con close(ch).
		// Este es el patrón "consumer" clásico: el producer llena el canal (Send),
		// el consumer lo vacía aquí, y Stop() sincroniza el cierre.
		for entry := range a.ch {
			a.logger.Log(context.Background(), entry.level, entry.msg, entry.args...)
		}
	}()
}

// Send enqueues a log entry without blocking. If the channel is full, the entry
// is silently dropped (fire-and-forget). This non-blocking select/default idiom
// evita que el goroutine del request HTTP quede bloqueado esperando al worker de logs.
func (a *AsyncLogger) Send(level slog.Level, msg string, args ...any) {
	entry := LogEntry{level: level, msg: msg, args: args}
	select {
	case a.ch <- entry:
		// entrada encolada correctamente
	default:
		// buffer lleno: se descarta. En producción aquí iría un contador atómico.
	}
}

// Stop cierra el canal y espera a que el worker drene todas las entradas pendientes.
// Debe llamarse después de que no haya más productores (e.g., tras srv.Shutdown).
func (a *AsyncLogger) Stop() {
	close(a.ch) // señala al worker que no habrá más entradas
	<-a.done    // espera a que el worker termine de procesar las que quedan
}
