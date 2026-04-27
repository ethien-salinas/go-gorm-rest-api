package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
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
			testMux := http.NewServeMux()
			testMux.HandleFunc("GET /api/v1/users/{id}", h.GetByID)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.id, nil)
			w := httptest.NewRecorder()
			testMux.ServeHTTP(w, req)
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
		userID     uint
		wantStatus int
	}{
		{
			name: "partial update one field",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, FirstName: "old", LastName: "lee", Email: "old@example.com"}, nil
				},
				updateFn: func(_ context.Context, u *models.User, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"first_name":"new"}`,
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name: "full update all fields",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, FirstName: "old"}, nil
				},
				updateFn: func(_ context.Context, u *models.User, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"first_name":"new","last_name":"lee","email":"new@example.com"}`,
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name: "forbidden",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 2, FirstName: "other"}, nil
				},
			},
			id:         "2",
			body:       `{"first_name":"new"}`,
			userID:     1,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "empty body returns 400",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, FirstName: "old"}, nil
				},
			},
			id:         "1",
			body:       `{}`,
			userID:     1,
			wantStatus: http.StatusBadRequest,
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
			userID:     1,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "bad JSON",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, FirstName: "old"}, nil
				},
			},
			id:         "1",
			body:       `{invalid}`,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, FirstName: "old"}, nil
				},
				updateFn: func(_ context.Context, u *models.User, fields map[string]any) error {
					return errors.New("db error")
				},
			},
			id:         "1",
			body:       `{"first_name":"new"}`,
			userID:     1,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.repo, testLogger())
			testMux := http.NewServeMux()
			testMux.HandleFunc("PATCH /api/v1/users/{id}", h.Update)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+tt.id, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = withUserID(req, tt.userID)
			w := httptest.NewRecorder()
			testMux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserHandler_BatchCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		batchFn    func(ctx context.Context, users []*models.User, workers int) []error
		wantStatus int
		wantErrors bool
	}{
		{
			name:       "all success",
			body:       `{"users":[{"first_name":"A","last_name":"B","email":"a@b.com"},{"first_name":"C","last_name":"D","email":"c@d.com"}]}`,
			batchFn:    func(_ context.Context, users []*models.User, _ int) []error { return make([]error, len(users)) },
			wantStatus: http.StatusCreated,
		},
		{
			name: "partial failure returns 207",
			body: `{"users":[{"first_name":"A","last_name":"B","email":"a@b.com"},{"first_name":"C","last_name":"D","email":"c@d.com"}]}`,
			batchFn: func(_ context.Context, users []*models.User, _ int) []error {
				errs := make([]error, len(users))
				errs[1] = errors.New("duplicate email")
				return errs
			},
			wantStatus: http.StatusMultiStatus,
			wantErrors: true,
		},
		{
			name:       "empty users",
			body:       `{"users":[]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad JSON",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{batchCreateFn: tt.batchFn}
			h := NewUserHandler(repo, testLogger())

			req := httptest.NewRequest(http.MethodPost, "/api/v1/users/batch", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.BatchCreate(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrors {
				var results []BatchCreateResult
				require.NoError(t, json.NewDecoder(w.Body).Decode(&results))
				hasErr := false
				for _, r := range results {
					if r.Error != "" {
						hasErr = true
					}
				}
				assert.True(t, hasErr, "expected at least one result with error")
			}
		})
	}
}

func TestUserHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		repo       UserRepository
		id         string
		userID     uint
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, FirstName: "tommy"}, nil
				},
				deleteFn: func(_ context.Context, u *models.User) error { return nil },
			},
			id:         "1",
			userID:     1,
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
			userID:     1,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "forbidden",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 2, FirstName: "other"}, nil
				},
			},
			id:         "2",
			userID:     1,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, FirstName: "tommy"}, nil
				},
				deleteFn: func(_ context.Context, u *models.User) error {
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
			h := NewUserHandler(tt.repo, testLogger())
			testMux := http.NewServeMux()
			testMux.HandleFunc("DELETE /api/v1/users/{id}", h.Delete)
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+tt.id, nil)
			req = withUserID(req, tt.userID)
			w := httptest.NewRecorder()
			testMux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserHandler_ChangePassword(t *testing.T) {
	// bcrypt.MinCost (4) para que los tests sean rápidos
	hash, err := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.MinCost)
	require.NoError(t, err)
	validHash := string(hash)

	tests := []struct {
		name       string
		repo       UserRepository
		id         string
		body       string
		userID     uint
		wantStatus int
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, PasswordHash: validHash}, nil
				},
				updateFn: func(_ context.Context, u *models.User, fields map[string]any) error { return nil },
			},
			id:         "1",
			body:       `{"current_password":"oldpass","new_password":"newpass"}`,
			userID:     1,
			wantStatus: http.StatusNoContent,
		},
		{
			name: "wrong current password",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, PasswordHash: validHash}, nil
				},
			},
			id:         "1",
			body:       `{"current_password":"wrong","new_password":"newpass"}`,
			userID:     1,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{}, errors.New("not found")
				},
			},
			id:         "99",
			body:       `{"current_password":"oldpass","new_password":"newpass"}`,
			userID:     1,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "forbidden",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 2, PasswordHash: validHash}, nil
				},
			},
			id:         "2",
			body:       `{"current_password":"oldpass","new_password":"newpass"}`,
			userID:     1,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "missing fields",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, PasswordHash: validHash}, nil
				},
			},
			id:         "1",
			body:       `{"current_password":"oldpass"}`,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "bad JSON",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, PasswordHash: validHash}, nil
				},
			},
			id:         "1",
			body:       `{invalid}`,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				findByIDFn: func(_ context.Context, id string) (models.User, error) {
					return models.User{ID: 1, PasswordHash: validHash}, nil
				},
				updateFn: func(_ context.Context, u *models.User, fields map[string]any) error {
					return errors.New("db error")
				},
			},
			id:         "1",
			body:       `{"current_password":"oldpass","new_password":"newpass"}`,
			userID:     1,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.repo, testLogger())
			testMux := http.NewServeMux()
			testMux.HandleFunc("PATCH /api/v1/users/{id}/password", h.ChangePassword)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+tt.id+"/password", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = withUserID(req, tt.userID)
			w := httptest.NewRecorder()
			testMux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
