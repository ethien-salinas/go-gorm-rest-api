package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/gorilla/mux"
)

// UserRepository defines the data-access operations required by [UserHandler].
type UserRepository interface {
	FindAll(ctx context.Context) ([]models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, user *models.User) error
}

// UserHandler handles HTTP requests for the /api/v1/users resource.
type UserHandler struct {
	repo   UserRepository
	logger *slog.Logger
}

// NewUserHandler returns a [UserHandler] that delegates persistence to repo.
func NewUserHandler(repo UserRepository, logger *slog.Logger) *UserHandler {
	return &UserHandler{repo: repo, logger: logger}
}

// GetAll writes a JSON array of all users to the response.
func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := h.repo.FindAll(r.Context())
	if err != nil {
		h.logger.Error("handler: failed to get users", "error", err)
		writeError(w, http.StatusInternalServerError, "error al obtener los usuarios")
		return
	}
	json.NewEncoder(w).Encode(users)
}

// GetByID writes the user identified by the route parameter {id} to the response.
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	user, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found", "id", params["id"])
		writeError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	json.NewEncoder(w).Encode(user)
}

type createUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type updateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// Create decodes a user from the request body and persists it, responding 201 on success.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode user", "error", err)
		writeError(w, http.StatusBadRequest, "error al decodificar el usuario")
		return
	}
	user := models.User{FirstName: req.FirstName, LastName: req.LastName, Email: req.Email}
	if err := h.repo.Create(r.Context(), &user); err != nil {
		h.logger.Error("handler: failed to create user", "error", err)
		writeError(w, http.StatusInternalServerError, "error al crear el usuario")
		return
	}
	h.logger.Info("user created", "id", user.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Update replaces the fields of the user identified by {id} with the values from the request body.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	params := mux.Vars(r)
	user, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found for update", "id", params["id"])
		writeError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode user", "error", err)
		writeError(w, http.StatusBadRequest, "error al decodificar el usuario")
		return
	}
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	if err := h.repo.Update(r.Context(), &user); err != nil {
		h.logger.Error("handler: failed to update user", "id", user.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "error al actualizar el usuario")
		return
	}
	h.logger.Info("user updated", "id", user.ID)
	json.NewEncoder(w).Encode(user)
}

// Delete soft-deletes the user identified by {id}.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	user, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found for delete", "id", params["id"])
		writeError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if err := h.repo.Delete(r.Context(), &user); err != nil {
		h.logger.Error("handler: failed to delete user", "id", user.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "error al eliminar el usuario")
		return
	}
	h.logger.Info("user deleted", "id", user.ID)
	json.NewEncoder(w).Encode(user)
}
