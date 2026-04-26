// Package logger provides a daily-rotating io.Writer for structured log output.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DailyRotator is an [io.WriteCloser] that writes to a new log file each calendar day.
// Files are named app-YYYY-MM-DD.log inside the configured directory.
// All methods are safe for concurrent use.
type DailyRotator struct {
	mu      sync.Mutex
	dir     string
	current *os.File
	day     string
}

// NewDailyRotator creates the log file for today inside dir.
// It creates dir with 0755 if it does not exist.
func NewDailyRotator(dir string) (*DailyRotator, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("logger: create log dir %q: %w", dir, err)
	}
	r := &DailyRotator{dir: dir}
	if err := r.rotate(time.Now()); err != nil {
		return nil, err
	}
	return r, nil
}

// Write implements [io.Writer]. Rotates to a new file when the calendar day changes.
func (r *DailyRotator) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if today := time.Now().Format("2006-01-02"); today != r.day {
		if err := r.rotate(time.Now()); err != nil {
			return 0, err
		}
	}
	return r.current.Write(p)
}

// Close flushes and closes the current log file.
func (r *DailyRotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.current != nil {
		return r.current.Close()
	}
	return nil
}

// rotate opens (or creates) the file for t's date. Caller must hold r.mu.
func (r *DailyRotator) rotate(t time.Time) error {
	if r.current != nil {
		_ = r.current.Close()
	}
	day := t.Format("2006-01-02")
	path := filepath.Join(r.dir, "app-"+day+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("logger: open log file %q: %w", path, err)
	}
	r.current, r.day = f, day
	return nil
}

// NewLogger builds the application [slog.Logger].
// When logToFile is true it writes to both os.Stdout and a [DailyRotator] under logDir.
// Returns the rotator so the caller can close it on shutdown; nil when logToFile is false.
func NewLogger(logToFile bool, logDir string) (*slog.Logger, *DailyRotator, error) {
	if !logToFile {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil)), nil, nil
	}
	rotator, err := NewDailyRotator(logDir)
	if err != nil {
		return nil, nil, err
	}
	w := io.MultiWriter(os.Stdout, rotator)
	return slog.New(slog.NewJSONHandler(w, nil)), rotator, nil
}
