package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/repository"
	"github.com/gorilla/mux"
)

type TaskHandler struct {
	repo   *repository.TaskRepository
	logger *slog.Logger
}

func NewTaskHandler(repo *repository.TaskRepository, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{repo: repo, logger: logger}
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tasks, err := h.repo.FindAll()
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
	task, err := h.repo.FindByID(params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.logger.Warn("handler: failed to decode task", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar la tarea"))
		return
	}
	if err := h.repo.Create(&task); err != nil {
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
	task, err := h.repo.FindByID(params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found for update", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.logger.Warn("handler: failed to decode task", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar la tarea"))
		return
	}
	if err := h.repo.Update(&task); err != nil {
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
	task, err := h.repo.FindByID(params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found for delete", "id", params["id"])
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	if err := h.repo.Delete(&task); err != nil {
		h.logger.Error("handler: failed to delete task", "id", task.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al eliminar la tarea"))
		return
	}
	h.logger.Info("task deleted", "id", task.ID)
	json.NewEncoder(w).Encode(task)
}
