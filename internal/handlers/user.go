package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository defines the data-access operations required by [UserHandler].
type UserRepository interface {
	FindAll(ctx context.Context) ([]models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User, fields map[string]any) error
	Delete(ctx context.Context, user *models.User) error
	BatchCreate(ctx context.Context, users []*models.User, workers int) []error
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
//
//	@Summary		Listar usuarios
//	@Description	Retorna todos los usuarios activos.
//	@Tags			users
//	@Produce		json
//	@Success		200	{array}		models.User
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/users [get]
func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := h.repo.FindAll(r.Context())
	if err != nil {
		h.logger.Error("handler: failed to get users", "error", err)
		writeError(w, http.StatusInternalServerError, "error al obtener los usuarios")
		return
	}
	if err := json.NewEncoder(w).Encode(users); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// GetByID writes the user identified by the route parameter {id} to the response.
//
//	@Summary		Obtener usuario por ID
//	@Description	Retorna un usuario por su ID.
//	@Tags			users
//	@Produce		json
//	@Param			id	path		int	true	"ID del usuario"
//	@Success		200	{object}	models.User
//	@Failure		404	{object}	ErrorResponse
//	@Router			/api/v1/users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	user, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found", "id", params["id"])
		writeError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// CreateUserRequest holds the fields accepted when creating a new user.
type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// updateUserRequest holds the fields accepted when partially updating an existing user.
// Pointer fields distinguish "not sent" (nil) from "sent as empty" ("").
type updateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
}

// Create decodes a user from the request body and persists it, responding 201 on success.
//
//	@Summary		Crear usuario
//	@Description	Crea un nuevo usuario con los datos del body.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			user	body		CreateUserRequest	true	"Datos del usuario"
//	@Success		201		{object}	models.User
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req CreateUserRequest
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
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// Update applies a partial update to the user identified by {id}.
// Only the fields present in the request body are modified; omitted fields are left unchanged.
//
//	@Summary		Actualizar usuario (parcial)
//	@Description	Actualiza solo los campos enviados en el body del usuario identificado por {id}.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"ID del usuario"
//	@Param			user	body		updateUserRequest	true	"Campos a actualizar"
//	@Success		200		{object}	models.User
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/users/{id} [patch]
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
	fields := make(map[string]any)
	if req.FirstName != nil {
		fields["first_name"] = *req.FirstName
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		fields["last_name"] = *req.LastName
		user.LastName = *req.LastName
	}
	if req.Email != nil {
		fields["email"] = *req.Email
		user.Email = *req.Email
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no se proporcionaron campos para actualizar")
		return
	}
	if err := h.repo.Update(r.Context(), &user, fields); err != nil {
		h.logger.Error("handler: failed to update user", "id", user.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "error al actualizar el usuario")
		return
	}
	h.logger.Info("user updated", "id", user.ID)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

// Delete soft-deletes the user identified by {id}.
//
//	@Summary		Eliminar usuario
//	@Description	Realiza un soft-delete del usuario identificado por {id}.
//	@Tags			users
//	@Produce		json
//	@Param			id	path		int	true	"ID del usuario"
//	@Success		200	{object}	models.User
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/users/{id} [delete]
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
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("handler: failed to encode response", "error", err)
	}
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword verifies the current password and replaces it with a new one.
// Responds 204 No Content on success.
//
//	@Summary		Cambiar contraseña
//	@Description	Verifica la contraseña actual y la reemplaza con la nueva.
//	@Tags			users
//	@Accept			json
//	@Param			id		path	int						true	"ID del usuario"
//	@Param			body	body	changePasswordRequest	true	"Contraseñas"
//	@Success		204
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/users/{id}/password [patch]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	params := mux.Vars(r)
	user, err := h.repo.FindByID(r.Context(), params["id"])
	if err != nil {
		h.logger.Warn("handler: user not found for password change", "id", params["id"])
		writeError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "current_password y new_password son requeridos")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		writeError(w, http.StatusUnauthorized, "contraseña actual incorrecta")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("handler: failed to hash new password", "error", err)
		writeError(w, http.StatusInternalServerError, "error al procesar la solicitud")
		return
	}

	if err := h.repo.Update(r.Context(), &user, map[string]any{"password_hash": string(newHash)}); err != nil {
		h.logger.Error("handler: failed to update password", "id", user.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "error al actualizar la contraseña")
		return
	}

	h.logger.Info("password changed", "id", user.ID)
	w.WriteHeader(http.StatusNoContent)
}

// batchCreateRequest holds the fields accepted by the batch create endpoint.
type batchCreateRequest struct {
	Users   []CreateUserRequest `json:"users"`
	Workers int                 `json:"workers"`
}

// BatchCreateResult reports the outcome for one user in a batch operation.
type BatchCreateResult struct {
	Index uint   `json:"index"`
	ID    uint   `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

const (
	maxBatchSize    = 100
	defaultWorkers  = 5
)

// BatchCreate inserts multiple users concurrently via a worker pool and responds 201 on full
// success or 207 Multi-Status if any insertion failed.
//
// Uses a buffered-channel semaphore to cap concurrent inserts at the requested workers value
// (default 5). Maximum batch size is 100 entries.
func (h *UserHandler) BatchCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req batchCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("handler: failed to decode batch request", "error", err)
		writeError(w, http.StatusBadRequest, "error al decodificar la solicitud")
		return
	}
	if len(req.Users) == 0 {
		writeError(w, http.StatusBadRequest, "la lista de usuarios no puede estar vacía")
		return
	}
	if len(req.Users) > maxBatchSize {
		writeError(w, http.StatusBadRequest, "máximo 100 usuarios por lote")
		return
	}

	workers := req.Workers
	if workers <= 0 {
		workers = defaultWorkers
	}

	users := make([]*models.User, len(req.Users))
	for i, u := range req.Users {
		users[i] = &models.User{FirstName: u.FirstName, LastName: u.LastName, Email: u.Email}
	}

	errs := h.repo.BatchCreate(r.Context(), users, workers)

	results := make([]BatchCreateResult, len(users))
	hasError := false
	for i, u := range users {
		results[i] = BatchCreateResult{Index: uint(i), ID: u.ID}
		if errs[i] != nil {
			results[i].Error = errs[i].Error()
			hasError = true
		}
	}

	// 207 indica que algunas operaciones del lote fallaron; 201 que todas tuvieron éxito.
	if hasError {
		w.WriteHeader(http.StatusMultiStatus)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error("handler: failed to encode batch response", "error", err)
	}
}
