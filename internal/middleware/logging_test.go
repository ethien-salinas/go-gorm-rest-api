package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	Logging(logger)(next).ServeHTTP(w, req)

	assert.True(t, called, "next handler should be called")
	assert.Equal(t, http.StatusCreated, w.Code)
}
