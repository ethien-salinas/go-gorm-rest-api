package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthUserRepository defines the data-access operations required by [AuthHandler].
type AuthUserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

// AuthHandler handles signup and login endpoints.
type AuthHandler struct {
	repo   AuthUserRepository
	secret string
	expiry time.Duration
	log    *slog.Logger
}

// NewAuthHandler returns an [AuthHandler] configured with the given JWT secret and expiry.
func NewAuthHandler(repo AuthUserRepository, secret string, expiryHours int, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		repo:   repo,
		secret: secret,
		expiry: time.Duration(expiryHours) * time.Hour,
		log:    log,
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("handler: failed to hash password", "error", err)
		writeError(w, http.StatusInternalServerError, "error al procesar la solicitud")
		return
	}

	user := models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hash),
	}
	if err := h.repo.Create(r.Context(), &user); err != nil {
		h.log.Error("handler: failed to create user on signup", "error", err)
		writeError(w, http.StatusInternalServerError, "error al registrar el usuario")
		return
	}

	h.log.Info("user signed up", "id", user.ID)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]any{"id": user.ID, "email": user.Email}); err != nil {
		h.log.Error("handler: failed to encode signup response", "error", err)
	}
}

// Login verifies credentials and responds 200 with a signed JWT on success.
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
		// Mensaje genérico para evitar enumeración de usuarios
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
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
