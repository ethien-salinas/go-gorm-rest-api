package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// newTestAuthHandler returns an AuthHandler configured for tests (5-attempt lockout, 30-min duration).
func newTestAuthHandler(repo AuthUserRepository, historyRepo AuthPasswordHistoryRepository) *AuthHandler {
	return NewAuthHandler(repo, historyRepo, "test-secret-key", 1, 5, 30, testLogger())
}

func TestAuthHandler_Signup(t *testing.T) {
	tests := []struct {
		name       string
		repo       AuthUserRepository
		body       string
		wantStatus int
		wantErrors bool
	}{
		{
			name: "success",
			repo: &mockAuthUserRepo{
				createFn: func(_ context.Context, u *models.User) error { return nil },
			},
			body:       `{"first_name":"Ana","last_name":"Lopez","email":"ana@example.com","password":"Str0ng!Pass#24"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "weak password returns validation errors",
			repo:       &mockAuthUserRepo{},
			body:       `{"first_name":"Ana","last_name":"Lopez","email":"ana@example.com","password":"weak"}`,
			wantStatus: http.StatusBadRequest,
			wantErrors: true,
		},
		{
			name:       "empty email",
			repo:       &mockAuthUserRepo{},
			body:       `{"password":"Str0ng!Pass#24"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty password",
			repo:       &mockAuthUserRepo{},
			body:       `{"email":"ana@example.com"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad JSON",
			repo:       &mockAuthUserRepo{},
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repo error",
			repo: &mockAuthUserRepo{
				createFn: func(_ context.Context, u *models.User) error {
					return errors.New("db error")
				},
			},
			body:       `{"first_name":"Ana","last_name":"Lopez","email":"ana@example.com","password":"Str0ng!Pass#24"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAuthHandler(tt.repo, &mockPasswordHistoryRepo{})
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Signup(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantErrors {
				var body map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
				assert.Contains(t, body, "errors", "response must have 'errors' key for policy violations")
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct_pass"), bcrypt.MinCost)
	require.NoError(t, err)
	validHash := string(hash)

	future := time.Now().Add(30 * time.Minute)

	tests := []struct {
		name        string
		repoUser    *models.User
		repoFindErr error
		body        string
		wantStatus  int
		checkFields func(*testing.T, map[string]any)
	}{
		{
			name:       "success",
			repoUser:   &models.User{ID: 1, Email: "u@e.com", PasswordHash: validHash},
			body:       `{"email":"u@e.com","password":"correct_pass"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:        "user not found",
			repoFindErr: errors.New("not found"),
			body:        `{"email":"nobody@e.com","password":"correct_pass"}`,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:       "wrong password increments counter",
			repoUser:   &models.User{ID: 1, PasswordHash: validHash, FailedLoginCount: 0},
			body:       `{"email":"u@e.com","password":"wrongpassword"}`,
			wantStatus: http.StatusUnauthorized,
			checkFields: func(t *testing.T, fields map[string]any) {
				assert.Equal(t, 1, fields["failed_login_count"])
				assert.Nil(t, fields["locked_until"])
			},
		},
		{
			name:       "account locked after threshold",
			repoUser:   &models.User{ID: 1, PasswordHash: validHash, FailedLoginCount: 4},
			body:       `{"email":"u@e.com","password":"wrongpassword"}`,
			wantStatus: http.StatusUnauthorized,
			checkFields: func(t *testing.T, fields map[string]any) {
				assert.Equal(t, 5, fields["failed_login_count"])
				assert.NotNil(t, fields["locked_until"])
			},
		},
		{
			name: "locked account rejected before password check",
			repoUser: &models.User{
				ID:               1,
				PasswordHash:     validHash,
				LockedUntil:      &future,
				FailedLoginCount: 5,
			},
			body:       `{"email":"u@e.com","password":"correct_pass"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty email",
			body:       `{"password":"correct_pass"}`,
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
			var capturedFields map[string]any
			repo := &mockAuthUserRepo{
				findByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
					if tt.repoFindErr != nil {
						return nil, tt.repoFindErr
					}
					return tt.repoUser, nil
				},
				updateFn: func(_ context.Context, _ *models.User, fields map[string]any) error {
					capturedFields = fields
					return nil
				},
			}
			h := newTestAuthHandler(repo, &mockPasswordHistoryRepo{})
			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Login(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.checkFields != nil {
				tt.checkFields(t, capturedFields)
			}
		})
	}
}
