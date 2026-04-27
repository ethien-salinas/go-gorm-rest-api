package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestTaskHandler_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		repo       TaskRepository
		injectUser bool
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockTaskRepo{
				findAllByUserFn: func(_ context.Context, _ uint) ([]models.Task, error) {
					return []models.Task{{Title: "learn gorm", Description: "build with gorm"}}, nil
				},
			},
			injectUser: true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing auth returns 401",
			repo:       &mockTaskRepo{},
			injectUser: false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "repo error",
			repo: &mockTaskRepo{
				findAllByUserFn: func(_ context.Context, _ uint) ([]models.Task, error) {
					return nil, errors.New("db error")
				},
			},
			injectUser: true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
			if tt.injectUser {
				req = withUserID(req, 1)
			}
			w := httptest.NewRecorder()
			h.GetAll(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestTaskHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		repo       TaskRepository
		id         string
		userID     uint
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "learn gorm", UserID: 1}, nil
				},
			},
			id:         "1",
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{}, errors.New("not found")
				},
			},
			id:         "99",
			userID:     1,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "forbidden",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "someone else's task", UserID: 2}, nil
				},
			},
			id:         "1",
			userID:     1,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			testMux := http.NewServeMux()
			testMux.HandleFunc("GET /api/v1/tasks/{id}", h.GetByID)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+tt.id, nil)
			req = withUserID(req, tt.userID)
			w := httptest.NewRecorder()
			testMux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestTaskHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		repo       TaskRepository
		body       string
		injectUser bool
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockTaskRepo{
				createFn: func(_ context.Context, t *models.Task) error { return nil },
			},
			body:       `{"title":"learn gorm","description":"build with gorm"}`,
			injectUser: true,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing auth returns 401",
			repo:       &mockTaskRepo{},
			body:       `{"title":"learn gorm"}`,
			injectUser: false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad JSON",
			repo:       &mockTaskRepo{},
			body:       `{invalid}`,
			injectUser: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockTaskRepo{
				createFn: func(_ context.Context, tk *models.Task) error {
					return errors.New("db error")
				},
			},
			body:       `{"title":"learn gorm","description":"build with gorm"}`,
			injectUser: true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.injectUser {
				req = withUserID(req, 1)
			}
			w := httptest.NewRecorder()
			h.Create(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestTaskHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		repo       TaskRepository
		id         string
		body       string
		userID     uint
		wantStatus int
	}{
		{
			name: "partial update one field",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old", Description: "desc", Done: false, UserID: 1}, nil
				},
				updateFn: func(_ context.Context, tk *models.Task, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"title":"new"}`,
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name: "partial update done flag",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old", UserID: 1}, nil
				},
				updateFn: func(_ context.Context, tk *models.Task, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"done":true}`,
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name: "full update all fields",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old", UserID: 1}, nil
				},
				updateFn: func(_ context.Context, tk *models.Task, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"title":"new","description":"updated","done":true}`,
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name: "empty body returns 400",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old", UserID: 1}, nil
				},
			},
			id:         "1",
			body:       `{}`,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "forbidden",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old", UserID: 2}, nil
				},
			},
			id:         "1",
			body:       `{"title":"new"}`,
			userID:     1,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "not found",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{}, errors.New("not found")
				},
			},
			id:         "99",
			body:       `{"title":"new"}`,
			userID:     1,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "bad JSON",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old", UserID: 1}, nil
				},
			},
			id:         "1",
			body:       `{invalid}`,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old", UserID: 1}, nil
				},
				updateFn: func(_ context.Context, tk *models.Task, fields map[string]any) error {
					return errors.New("db error")
				},
			},
			id:         "1",
			body:       `{"title":"new"}`,
			userID:     1,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			testMux := http.NewServeMux()
			testMux.HandleFunc("PATCH /api/v1/tasks/{id}", h.Update)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+tt.id, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = withUserID(req, tt.userID)
			w := httptest.NewRecorder()
			testMux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestTaskHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		repo       TaskRepository
		id         string
		userID     uint
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "learn gorm", UserID: 1}, nil
				},
				deleteFn: func(_ context.Context, tk *models.Task) error { return nil },
			},
			id:         "1",
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{}, errors.New("not found")
				},
			},
			id:         "99",
			userID:     1,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "forbidden",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "learn gorm", UserID: 2}, nil
				},
			},
			id:         "1",
			userID:     1,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "repo error",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "learn gorm", UserID: 1}, nil
				},
				deleteFn: func(_ context.Context, tk *models.Task) error {
					return errors.New("db error")
				},
			},
			id:         "1",
			userID:     1,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			testMux := http.NewServeMux()
			testMux.HandleFunc("DELETE /api/v1/tasks/{id}", h.Delete)
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+tt.id, nil)
			req = withUserID(req, tt.userID)
			w := httptest.NewRecorder()
			testMux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
