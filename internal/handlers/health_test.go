package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		pingerErr  error
		wantStatus int
	}{
		{"db reachable", nil, http.StatusOK},
		{"db unreachable", errors.New("connection refused"), http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHealthHandler(mockPinger{err: tt.pingerErr})
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			h(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
