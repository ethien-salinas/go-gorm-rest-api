package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/gorilla/mux"
)

type UserRepository interface {
	FindAll(ctx context.Context) ([]models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, user *models.User) error
}

type UserHandler struct {
	repo   UserRepository
	logger *slog.Logger
}

func NewUserHandler(repo UserRepository, logger *slog.Logger) *UserHandler {
	return &UserHandler{repo: repo, logger: logger}
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := h.repo.FindAll(r.Context())
	if err != nil {
		h.logger.Error("handler: failed to get users", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al obtener los usuarios"))
		return
	}
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	user, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
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

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode user", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar el usuario"))
		return
	}
	user := models.User{FirstName: req.FirstName, LastName: req.LastName, Email: req.Email}
	if err := h.repo.Create(r.Context(), &user); err != nil {
		h.logger.Error("handler: failed to create user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al crear el usuario"))
		return
	}
	h.logger.Info("user created", "id", user.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	params := mux.Vars(r)
	user, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found for update", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
		return
	}
	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode user", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar el usuario"))
		return
	}
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	if err := h.repo.Update(r.Context(), &user); err != nil {
		h.logger.Error("handler: failed to update user", "id", user.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al actualizar el usuario"))
		return
	}
	h.logger.Info("user updated", "id", user.ID)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	user, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found for delete", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
		return
	}
	if err := h.repo.Delete(r.Context(), &user); err != nil {
		h.logger.Error("handler: failed to delete user", "id", user.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al eliminar el usuario"))
		return
	}
	h.logger.Info("user deleted", "id", user.ID)
	json.NewEncoder(w).Encode(user)
}
