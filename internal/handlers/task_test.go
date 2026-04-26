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
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockTaskRepo{
				findAllFn: func(_ context.Context) ([]models.Task, error) {
					return []models.Task{{Title: "learn gorm", Description: "build with gorm"}}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "repo error",
			repo: &mockTaskRepo{
				findAllFn: func(_ context.Context) ([]models.Task, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
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
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "learn gorm"}, nil
				},
			},
			id:         "1",
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
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			testMux := http.NewServeMux()
			testMux.HandleFunc("GET /api/v1/tasks/{id}", h.GetByID)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+tt.id, nil)
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
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockTaskRepo{
				createFn: func(_ context.Context, t *models.Task) error { return nil },
			},
			body:       `{"title":"learn gorm","description":"build with gorm","user_id":1}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "bad JSON",
			repo:       &mockTaskRepo{},
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockTaskRepo{
				createFn: func(_ context.Context, tk *models.Task) error {
					return errors.New("db error")
				},
			},
			body:       `{"title":"learn gorm","description":"build with gorm","user_id":1}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
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
		wantStatus int
	}{
		{
			name: "partial update one field",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old", Description: "desc", Done: false}, nil
				},
				updateFn: func(_ context.Context, tk *models.Task, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"title":"new"}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "partial update done flag",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old"}, nil
				},
				updateFn: func(_ context.Context, tk *models.Task, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"done":true}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "full update all fields",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old"}, nil
				},
				updateFn: func(_ context.Context, tk *models.Task, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"title":"new","description":"updated","done":true}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "empty body returns 400",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old"}, nil
				},
			},
			id:         "1",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
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
			wantStatus: http.StatusNotFound,
		},
		{
			name: "bad JSON",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old"}, nil
				},
			},
			id:         "1",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "old"}, nil
				},
				updateFn: func(_ context.Context, tk *models.Task, fields map[string]any) error {
					return errors.New("db error")
				},
			},
			id:         "1",
			body:       `{"title":"new"}`,
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
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "learn gorm"}, nil
				},
				deleteFn: func(_ context.Context, tk *models.Task) error { return nil },
			},
			id:         "1",
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
			wantStatus: http.StatusNotFound,
		},
		{
			name: "repo error",
			repo: &mockTaskRepo{
				findByIDFn: func(_ context.Context, id string) (models.Task, error) {
					return models.Task{Title: "learn gorm"}, nil
				},
				deleteFn: func(_ context.Context, tk *models.Task) error {
					return errors.New("db error")
				},
			},
			id:         "1",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(tt.repo, testLogger())
			testMux := http.NewServeMux()
			testMux.HandleFunc("DELETE /api/v1/tasks/{id}", h.Delete)
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+tt.id, nil)
			w := httptest.NewRecorder()
			testMux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
