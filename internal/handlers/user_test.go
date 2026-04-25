package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		repo       UserRepository
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				findAllFn: func(_ context.Context) ([]models.User, error) {
					return []models.User{{FirstName: "tommy", LastName: "lee", Email: "tommy@example.com"}}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				findAllFn: func(_ context.Context) ([]models.User, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
			w := httptest.NewRecorder()
			h.GetAll(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		repo       UserRepository
		id         string
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{FirstName: "tommy"}, nil
				},
			},
			id:         "1",
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{}, errors.New("not found")
				},
			},
			id:         "99",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()
			h.GetByID(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		repo       UserRepository
		body       string
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				createFn: func(_ context.Context, u *models.User) error { return nil },
			},
			body:       `{"first_name":"tommy","last_name":"lee","email":"tommy@example.com"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "bad JSON",
			repo:       &mockUserRepo{},
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				createFn: func(_ context.Context, u *models.User) error {
					return errors.New("db error")
				},
			},
			body:       `{"first_name":"tommy","last_name":"lee","email":"tommy@example.com"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Create(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		repo       UserRepository
		id         string
		body       string
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{FirstName: "old"}, nil
				},
				updateFn: func(_ context.Context, u *models.User) error { return nil },
			},
			id:         "1",
			body:       `{"first_name":"new","last_name":"lee","email":"new@example.com"}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{}, errors.New("not found")
				},
			},
			id:         "99",
			body:       `{"first_name":"new"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "bad JSON",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{FirstName: "old"}, nil
				},
			},
			id:         "1",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{FirstName: "old"}, nil
				},
				updateFn: func(_ context.Context, u *models.User) error {
					return errors.New("db error")
				},
			},
			id:         "1",
			body:       `{"first_name":"new","last_name":"lee","email":"new@example.com"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+tt.id, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()
			h.Update(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		repo       UserRepository
		id         string
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{FirstName: "tommy"}, nil
				},
				deleteFn: func(_ context.Context, u *models.User) error { return nil },
			},
			id:         "1",
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{}, errors.New("not found")
				},
			},
			id:         "99",
			wantStatus: http.StatusNotFound,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{FirstName: "tommy"}, nil
				},
				deleteFn: func(_ context.Context, u *models.User) error {
					return errors.New("db error")
				},
			},
			id:         "1",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.repo, testLogger())
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()
			h.Delete(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
