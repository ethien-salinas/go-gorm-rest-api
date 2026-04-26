package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/gorilla/mux"
)

// TaskRepository defines the data-access operations required by [TaskHandler].
type TaskRepository interface {
	FindAll(ctx context.Context) ([]models.Task, error)
	FindByID(ctx context.Context, id string) (models.Task, error)
	Create(ctx context.Context, task *models.Task) error
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, task *models.Task) error
}

// TaskHandler handles HTTP requests for the /api/v1/tasks resource.
type TaskHandler struct {
	repo   TaskRepository
	logger *slog.Logger
}

// NewTaskHandler returns a [TaskHandler] that delegates persistence to repo.
func NewTaskHandler(repo TaskRepository, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{repo: repo, logger: logger}
}

// GetAll writes a JSON array of all tasks to the response.
//
//	@Summary		Listar tareas
//	@Description	Retorna todas las tareas activas con su usuario asociado.
//	@Tags			tasks
//	@Produce		json
//	@Success		200	{array}		models.Task
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/tasks [get]
func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tasks, err := h.repo.FindAll(r.Context())
	if err != nil {
		h.logger.Error("handler: failed to get tasks", "error", err)
		writeError(w, http.StatusInternalServerError, "error al obtener las tareas")
		return
	}
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// GetByID writes the task identified by the route parameter {id} to the response.
//
//	@Summary		Obtener tarea por ID
//	@Description	Retorna una tarea por su ID, incluyendo el usuario asociado.
//	@Tags			tasks
//	@Produce		json
//	@Param			id	path		int	true	"ID de la tarea"
//	@Success		200	{object}	models.Task
//	@Failure		404	{object}	ErrorResponse
//	@Router			/api/v1/tasks/{id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	task, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found", "id", params["id"])
		writeError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	if err := json.NewEncoder(w).Encode(task); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// CreateTaskRequest holds the fields accepted when creating a new task.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
	UserID      uint   `json:"user_id"`
}

// UpdateTaskRequest holds the fields accepted when updating an existing task.
type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// Create decodes a task from the request body and persists it, responding 201 on success.
//
//	@Summary		Crear tarea
//	@Description	Crea una nueva tarea con los datos del body.
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			task	body		CreateTaskRequest	true	"Datos de la tarea"
//	@Success		201		{object}	models.Task
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode task", "error", err)
		writeError(w, http.StatusBadRequest, "error al decodificar la tarea")
		return
	}
	task := models.Task{Title: req.Title, Description: req.Description, Done: req.Done, UserID: req.UserID}
	if err := h.repo.Create(r.Context(), &task); err != nil {
		h.logger.Error("handler: failed to create task", "error", err)
		writeError(w, http.StatusInternalServerError, "error al crear la tarea")
		return
	}
	h.logger.Info("task created", "id", task.ID)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// Update replaces the fields of the task identified by {id} with the values from the request body.
//
//	@Summary		Actualizar tarea
//	@Description	Reemplaza los campos de la tarea identificada por {id}.
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"ID de la tarea"
//	@Param			task	body		UpdateTaskRequest	true	"Nuevos datos de la tarea"
//	@Success		200		{object}	models.Task
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/tasks/{id} [put]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	params := mux.Vars(r)
	task, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found for update", "id", params["id"])
		writeError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode task", "error", err)
		writeError(w, http.StatusBadRequest, "error al decodificar la tarea")
		return
	}
	task.Title = req.Title
	task.Description = req.Description
	task.Done = req.Done
	if err := h.repo.Update(r.Context(), &task); err != nil {
		h.logger.Error("handler: failed to update task", "id", task.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "error al actualizar la tarea")
		return
	}
	h.logger.Info("task updated", "id", task.ID)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// Delete soft-deletes the task identified by {id}.
//
//	@Summary		Eliminar tarea
//	@Description	Realiza un soft-delete de la tarea identificada por {id}.
//	@Tags			tasks
//	@Produce		json
//	@Param			id	path		int	true	"ID de la tarea"
//	@Success		200	{object}	models.Task
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	task, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: task not found for delete", "id", params["id"])
		writeError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	if err := h.repo.Delete(r.Context(), &task); err != nil {
		h.logger.Error("handler: failed to delete task", "id", task.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "error al eliminar la tarea")
		return
	}
	h.logger.Info("task deleted", "id", task.ID)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}
