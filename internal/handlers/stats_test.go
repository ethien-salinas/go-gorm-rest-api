package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsHandler_GetStats(t *testing.T) {
	tests := []struct {
		name           string
		userCount      int64
		taskCount      int64
		userErr        error
		taskErr        error
		wantStatus     int
		wantTotalUsers int64
		wantTotalTasks int64
		wantErrors     bool
	}{
		{
			name:           "success",
			userCount:      42,
			taskCount:      153,
			wantStatus:     http.StatusOK,
			wantTotalUsers: 42,
			wantTotalTasks: 153,
		},
		{
			name:       "user count error",
			userErr:    errors.New("db error"),
			taskCount:  10,
			wantStatus: http.StatusOK,
			wantErrors: true,
		},
		{
			name:       "task count error",
			userCount:  5,
			taskErr:    errors.New("db error"),
			wantStatus: http.StatusOK,
			wantErrors: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			users := &mockUserCounter{
				countFn: func(_ context.Context) (int64, error) {
					return tc.userCount, tc.userErr
				},
			}
			tasks := &mockTaskCounter{
				countFn: func(_ context.Context) (int64, error) {
					return tc.taskCount, tc.taskErr
				},
			}

			h := NewStatsHandler(users, tasks, testLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
			w := httptest.NewRecorder()

			h.GetStats(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)

			var resp statsResponse
			require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

			if !tc.wantErrors {
				assert.Equal(t, tc.wantTotalUsers, resp.TotalUsers)
				assert.Equal(t, tc.wantTotalTasks, resp.TotalTasks)
				assert.Empty(t, resp.Errors)
			} else {
				assert.NotEmpty(t, resp.Errors)
			}
		})
	}
}
