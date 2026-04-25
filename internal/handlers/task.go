package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/gorilla/mux"
)

type TaskRepository interface {
	FindAll(ctx context.Context) ([]models.Task, error)
	FindByID(ctx context.Context, id string) (models.Task, error)
	Create(ctx context.Context, task *models.Task) error
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, task *models.Task) error
}

type TaskHandler struct {
	repo   TaskRepository
	logger *slog.Logger
}

func NewTaskHandler(repo TaskRepository, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{repo: repo, logger: logger}
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tasks, err := h.repo.FindAll(r.Context())
	if err != nil {
		h.logger.Error("handler: failed to get tasks", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al obtener las tareas"))
		return
	}
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	task, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	json.NewEncoder(w).Encode(task)
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
	UserID      uint   `json:"user_id"`
}

type updateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode task", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar la tarea"))
		return
	}
	task := models.Task{Title: req.Title, Description: req.Description, Done: req.Done, UserID: req.UserID}
	if err := h.repo.Create(r.Context(), &task); err != nil {
		h.logger.Error("handler: failed to create task", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al crear la tarea"))
		return
	}
	h.logger.Info("task created", "id", task.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	params := mux.Vars(r)
	task, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found for update", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode task", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar la tarea"))
		return
	}
	task.Title = req.Title
	task.Description = req.Description
	task.Done = req.Done
	if err := h.repo.Update(r.Context(), &task); err != nil {
		h.logger.Error("handler: failed to update task", "id", task.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al actualizar la tarea"))
		return
	}
	h.logger.Info("task updated", "id", task.ID)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	task, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found for delete", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	if err := h.repo.Delete(r.Context(), &task); err != nil {
		h.logger.Error("handler: failed to delete task", "id", task.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al eliminar la tarea"))
		return
	}
	h.logger.Info("task deleted", "id", task.ID)
	json.NewEncoder(w).Encode(task)
}
