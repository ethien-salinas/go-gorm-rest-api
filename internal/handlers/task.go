package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/middleware"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
)

// TaskRepository defines the data-access operations required by [TaskHandler].
type TaskRepository interface {
	FindAll(ctx context.Context) ([]models.Task, error)
	FindAllByUser(ctx context.Context, userID uint) ([]models.Task, error)
	FindByID(ctx context.Context, id string) (models.Task, error)
	Create(ctx context.Context, task *models.Task) error
	Update(ctx context.Context, task *models.Task, fields map[string]any) error
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
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tasks, err := h.repo.FindAllByUser(r.Context(), userID)
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
	id := r.PathValue("id")
	task, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("handler: task not found", "id", id)
		writeError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || task.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := json.NewEncoder(w).Encode(task); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// CreateTaskRequest holds the fields accepted when creating a new task.
// UserID is not accepted from the client; it is always set from the authenticated JWT.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// updateTaskRequest holds the fields accepted when partially updating an existing task.
// Pointer fields distinguish "not sent" (nil) from "sent as zero value".
type updateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Done        *bool   `json:"done"`
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
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode task", "error", err)
		writeError(w, http.StatusBadRequest, "error al decodificar la tarea")
		return
	}
	task := models.Task{Title: req.Title, Description: req.Description, Done: req.Done, UserID: userID}
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

// Update applies a partial update to the task identified by {id}.
// Only the fields present in the request body are modified; omitted fields are left unchanged.
//
//	@Summary		Actualizar tarea (parcial)
//	@Description	Actualiza solo los campos enviados en el body de la tarea identificada por {id}.
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"ID de la tarea"
//	@Param			task	body		updateTaskRequest	true	"Campos a actualizar"
//	@Success		200		{object}	models.Task
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/tasks/{id} [patch]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	id := r.PathValue("id")
	task, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("handler: task not found for update", "id", id)
		writeError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || task.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode task", "error", err)
		writeError(w, http.StatusBadRequest, "error al decodificar la tarea")
		return
	}
	fields := make(map[string]any)
	if req.Title != nil {
		fields["title"] = *req.Title
		task.Title = *req.Title
	}
	if req.Description != nil {
		fields["description"] = *req.Description
		task.Description = *req.Description
	}
	if req.Done != nil {
		fields["done"] = *req.Done
		task.Done = *req.Done
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no se proporcionaron campos para actualizar")
		return
	}
	if err := h.repo.Update(r.Context(), &task, fields); err != nil {
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
	id := r.PathValue("id")
	task, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("handler: task not found for delete", "id", id)
		writeError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || task.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
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
