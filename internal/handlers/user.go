package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/repository"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	repo   *repository.UserRepository
	logger *slog.Logger
}

func NewUserHandler(repo *repository.UserRepository, logger *slog.Logger) *UserHandler {
	return &UserHandler{repo: repo, logger: logger}
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := h.repo.FindAll()
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
	user, err := h.repo.FindByID(params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.logger.Warn("handler: failed to decode user", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar el usuario"))
		return
	}
	if err := h.repo.Create(&user); err != nil {
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
	user, err := h.repo.FindByID(params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found for update", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.logger.Warn("handler: failed to decode user", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar el usuario"))
		return
	}
	if err := h.repo.Update(&user); err != nil {
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
	user, err := h.repo.FindByID(params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found for delete", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
		return
	}
	if err := h.repo.Delete(&user); err != nil {
		h.logger.Error("handler: failed to delete user", "id", user.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al eliminar el usuario"))
		return
	}
	h.logger.Info("user deleted", "id", user.ID)
	json.NewEncoder(w).Encode(user)
}
