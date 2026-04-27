package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/password"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthUserRepository defines the data-access operations required by [AuthHandler].
type AuthUserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User, fields map[string]any) error
}

// AuthPasswordHistoryRepository defines the data-access operations for password history
// needed by [AuthHandler].
type AuthPasswordHistoryRepository interface {
	Create(ctx context.Context, h *models.PasswordHistory) error
}

// AuthHandler handles signup and login endpoints.
type AuthHandler struct {
	repo             AuthUserRepository
	historyRepo      AuthPasswordHistoryRepository
	secret           string
	expiry           time.Duration
	log              *slog.Logger
	lockoutThreshold int
	lockoutDuration  time.Duration
}

// NewAuthHandler returns an [AuthHandler] configured with the given dependencies and policy values.
func NewAuthHandler(
	repo AuthUserRepository,
	historyRepo AuthPasswordHistoryRepository,
	secret string,
	expiryHours int,
	lockoutThreshold int,
	lockoutDurationMinutes int,
	log *slog.Logger,
) *AuthHandler {
	return &AuthHandler{
		repo:             repo,
		historyRepo:      historyRepo,
		secret:           secret,
		expiry:           time.Duration(expiryHours) * time.Hour,
		log:              log,
		lockoutThreshold: lockoutThreshold,
		lockoutDuration:  time.Duration(lockoutDurationMinutes) * time.Minute,
	}
}

type signupRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

// Signup creates a new user account and responds 201 with {id, email}.
//
//	@Summary		Registro de usuario
//	@Description	Crea una nueva cuenta con email y contraseña hasheada con bcrypt.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		signupRequest	true	"Datos de registro"
//	@Success		201		{object}	map[string]any
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/auth/signup [post]
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email y password son requeridos")
		return
	}

	if violations := password.Validate(req.Password, req.Email); len(violations) > 0 {
		writeValidationErrors(w, violations)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), password.BcryptCost)
	if err != nil {
		h.log.Error("handler: failed to hash password", "error", err)
		writeError(w, http.StatusInternalServerError, "error al procesar la solicitud")
		return
	}

	user := models.User{
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Email:             req.Email,
		PasswordHash:      string(hash),
		PasswordChangedAt: time.Now(),
	}
	if err := h.repo.Create(r.Context(), &user); err != nil {
		h.log.Error("handler: failed to create user on signup", "error", err)
		writeError(w, http.StatusInternalServerError, "error al registrar el usuario")
		return
	}

	if err := h.historyRepo.Create(r.Context(), &models.PasswordHistory{
		UserID:       user.ID,
		PasswordHash: string(hash),
	}); err != nil {
		h.log.Error("handler: failed to persist password history on signup", "userID", user.ID, "error", err)
		// Non-fatal: user is already created
	}

	h.log.Info("user signed up", "id", user.ID)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]any{"id": user.ID, "email": user.Email}); err != nil {
		h.log.Error("handler: failed to encode signup response", "error", err)
	}
}

// Login verifies credentials and responds 200 with a signed JWT on success.
// Failed attempts increment a counter; reaching the lockout threshold blocks the account.
//
//	@Summary		Inicio de sesión
//	@Description	Verifica email y contraseña; devuelve un JWT Bearer token firmado con HS256.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"Credenciales"
//	@Success		200		{object}	tokenResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email y password son requeridos")
		return
	}

	user, err := h.repo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		// Generic message to prevent user enumeration
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}

	// Check account lockout before verifying password (PCI DSS 8.3.4)
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		writeError(w, http.StatusUnauthorized, "cuenta bloqueada temporalmente, intenta más tarde")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		user.FailedLoginCount++
		var lockedUntil *time.Time
		if user.FailedLoginCount >= h.lockoutThreshold {
			t := time.Now().Add(h.lockoutDuration)
			lockedUntil = &t
			h.log.Warn("account locked due to failed login attempts", "userID", user.ID)
		}
		if updateErr := h.repo.Update(r.Context(), user, map[string]any{
			"failed_login_count": user.FailedLoginCount,
			"locked_until":       lockedUntil,
		}); updateErr != nil {
			h.log.Error("handler: failed to update failed login count", "error", updateErr)
		}
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}

	// Successful authentication: reset lockout state
	if updateErr := h.repo.Update(r.Context(), user, map[string]any{
		"failed_login_count": 0,
		"locked_until":       nil,
	}); updateErr != nil {
		h.log.Error("handler: failed to reset login counter", "error", updateErr)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(h.expiry).Unix(),
	})
	signed, err := token.SignedString([]byte(h.secret))
	if err != nil {
		h.log.Error("handler: failed to sign token", "error", err)
		writeError(w, http.StatusInternalServerError, "error al generar el token")
		return
	}

	h.log.Info("user logged in", "id", user.ID)
	if err := json.NewEncoder(w).Encode(tokenResponse{Token: signed}); err != nil {
		h.log.Error("handler: failed to encode login response", "error", err)
	}
}
