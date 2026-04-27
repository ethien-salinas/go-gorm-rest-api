package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/middleware"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
)

type mockUserRepo struct {
	findAllFn     func(ctx context.Context) ([]models.User, error)
	findByIDFn    func(ctx context.Context, id string) (models.User, error)
	createFn      func(ctx context.Context, u *models.User) error
	updateFn      func(ctx context.Context, u *models.User, fields map[string]any) error
	deleteFn      func(ctx context.Context, u *models.User) error
	batchCreateFn func(ctx context.Context, users []*models.User, workers int) []error
}

type mockAuthUserRepo struct {
	findByEmailFn func(ctx context.Context, email string) (*models.User, error)
	createFn      func(ctx context.Context, u *models.User) error
	updateFn      func(ctx context.Context, u *models.User, fields map[string]any) error
}

func (m *mockAuthUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockAuthUserRepo) Create(ctx context.Context, u *models.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, u)
	}
	return nil
}

func (m *mockAuthUserRepo) Update(ctx context.Context, u *models.User, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, u, fields)
	}
	return nil
}

type mockPasswordHistoryRepo struct {
	createFn             func(ctx context.Context, h *models.PasswordHistory) error
	findRecentByUserIDFn func(ctx context.Context, userID uint, limit int) ([]models.PasswordHistory, error)
}

func (m *mockPasswordHistoryRepo) Create(ctx context.Context, h *models.PasswordHistory) error {
	if m.createFn != nil {
		return m.createFn(ctx, h)
	}
	return nil
}

func (m *mockPasswordHistoryRepo) FindRecentByUserID(ctx context.Context, userID uint, limit int) ([]models.PasswordHistory, error) {
	if m.findRecentByUserIDFn != nil {
		return m.findRecentByUserIDFn(ctx, userID, limit)
	}
	return nil, nil
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]models.User, error) {
	return m.findAllFn(ctx)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (models.User, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockUserRepo) Create(ctx context.Context, u *models.User) error {
	return m.createFn(ctx, u)
}

func (m *mockUserRepo) Update(ctx context.Context, u *models.User, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, u, fields)
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, u *models.User) error {
	return m.deleteFn(ctx, u)
}

func (m *mockUserRepo) BatchCreate(ctx context.Context, users []*models.User, workers int) []error {
	if m.batchCreateFn != nil {
		return m.batchCreateFn(ctx, users, workers)
	}
	return make([]error, len(users))
}

type mockTaskRepo struct {
	findAllFn       func(ctx context.Context) ([]models.Task, error)
	findAllByUserFn func(ctx context.Context, userID uint) ([]models.Task, error)
	findByIDFn      func(ctx context.Context, id string) (models.Task, error)
	createFn        func(ctx context.Context, t *models.Task) error
	updateFn        func(ctx context.Context, t *models.Task, fields map[string]any) error
	deleteFn        func(ctx context.Context, t *models.Task) error
}

func (m *mockTaskRepo) FindAll(ctx context.Context) ([]models.Task, error) {
	return m.findAllFn(ctx)
}

func (m *mockTaskRepo) FindAllByUser(ctx context.Context, userID uint) ([]models.Task, error) {
	return m.findAllByUserFn(ctx, userID)
}

func (m *mockTaskRepo) FindByID(ctx context.Context, id string) (models.Task, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockTaskRepo) Create(ctx context.Context, t *models.Task) error {
	return m.createFn(ctx, t)
}

func (m *mockTaskRepo) Update(ctx context.Context, t *models.Task, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, t, fields)
	}
	return nil
}

func (m *mockTaskRepo) Delete(ctx context.Context, t *models.Task) error {
	return m.deleteFn(ctx, t)
}

// withUserID returns a copy of r with the authenticated user ID injected into the context.
func withUserID(r *http.Request, id uint) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, float64(id))
	return r.WithContext(ctx)
}

type mockPinger struct{ err error }

func (m mockPinger) PingContext(_ context.Context) error { return m.err }

type mockUserCounter struct {
	countFn func(ctx context.Context) (int64, error)
}

func (m *mockUserCounter) Count(ctx context.Context) (int64, error) {
	return m.countFn(ctx)
}

type mockTaskCounter struct {
	countFn func(ctx context.Context) (int64, error)
}

func (m *mockTaskCounter) Count(ctx context.Context) (int64, error) {
	return m.countFn(ctx)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
