# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Development (hot reload via Air)
air

# Run without hot reload
go run ./cmd/api

# Build
go build -o ./tmp/main ./cmd/api

# Dependencies
go mod tidy
```

The project requires a running PostgreSQL instance. Start it with:
```bash
docker compose up -d
```

Environment variables are loaded from `.env` automatically (via godotenv). Copy `.env.example` to `.env` before running.

## Architecture

```
cmd/api/main.go          ← wiring: crea config, DB, repos, handlers y registra rutas
internal/config/         ← lee variables de entorno y devuelve un Config struct
internal/database/       ← abre la conexión GORM; retorna (*gorm.DB, error), no llama os.Exit
internal/models/         ← structs de GORM (User, Task); User tiene []Task via foreignKey
internal/repository/     ← consultas GORM encapsuladas por entidad (UserRepository, TaskRepository)
internal/handlers/       ← HTTP handlers como métodos de struct; reciben su repo por inyección
internal/middleware/     ← middleware HTTP: Logging usa AsyncLogger (canal buffereado + goroutine worker)
internal/logger/         ← DailyRotator: io.Writer con rotación diaria y sync.Mutex
```

**Flujo de dependencias:** `main` instancia `config.Load()` → bootstrap paralelo con `sync.WaitGroup` (logger + DB) → `repository.New*(db)` → `handlers.New*(repo)`. Ninguna capa conoce a la superior.

**Rutas registradas:**

```
GET    /                             → handlers.HomeHandler          (info JSON de la API)
GET    /health                       → handlers.NewHealthHandler      (ping a la BD)

POST   /auth/signup                  → AuthHandler.Signup             (registro; público)
POST   /auth/login                   → AuthHandler.Login              (login; devuelve JWT; público)

— subrouter /api/v1 protegido por middleware.Auth —
GET    /api/v1/users
POST   /api/v1/users/batch           → BatchCreate: worker pool paralelo (máx. 100 usuarios)
GET    /api/v1/users/{id}
POST   /api/v1/users
PATCH  /api/v1/users/{id}            → actualización parcial (solo campos enviados)
PATCH  /api/v1/users/{id}/password   → cambio de contraseña (requiere contraseña actual)
DELETE /api/v1/users/{id}
GET    /api/v1/tasks
GET    /api/v1/tasks/{id}
POST   /api/v1/tasks
PATCH  /api/v1/tasks/{id}            → actualización parcial (solo campos enviados)
DELETE /api/v1/tasks/{id}
GET    /api/v1/stats                 → StatsHandler: fan-out de Count(users) y Count(tasks)
```

> `/users/batch` debe registrarse **antes** de `/users/{id}` para que gorilla/mux no interprete "batch" como un ID.

**Añadir un nuevo recurso** sigue este orden:
1. Modelo en `internal/models/` con campos explícitos (sin `gorm.Model`) y json tags en snake_case
2. Repositorio en `internal/repository/` con métodos `FindAll`, `FindByID`, `Create`, `Update`, `Delete`
3. Handler en `internal/handlers/`: definir primero la interfaz del repositorio en el mismo archivo, luego el struct handler, constructor y métodos
4. Registro de rutas en `cmd/api/main.go` bajo el subrouter `/api/v1`

**Añadir un endpoint con concurrencia** — patrones disponibles en el proyecto:
- **Fan-out (queries independientes):** ver `handlers/stats.go` — goroutines + canales buffereados de tamaño 1
- **Worker pool (batch IO):** ver `repository/user.go:BatchCreate` — semáforo con `chan struct{}{N}` + WaitGroup
- **Bootstrap paralelo:** ver `cmd/api/main.go` — `sync.WaitGroup` con goroutines para IO independiente
- **Async worker:** ver `middleware/async_logger.go` — canal buffereado + goroutine consumer + graceful shutdown

## Convenciones

**Estructura y nombrado**
- Archivos: `snake_case.go`
- El paquete `internal/` impide que módulos externos importen código de aplicación

**Interfaces de repositorio**
- Cada handler define su propia interfaz (e.g., `UserRepository` en `handlers/user.go`), no importa el tipo concreto del paquete `repository`. Esto desacopla las capas y facilita el testing con mocks.

**DTOs de request**
- Nunca usar el modelo directamente para leer el body. Definir structs privados (`createXRequest`, `updateXRequest`) con solo los campos que el cliente puede enviar. Esto evita mass-assignment de campos de auditoría (`id`, `created_at`, `deleted_at`).
- Los DTOs de endpoints PATCH usan **punteros** (`*string`, `*bool`) para distinguir "campo no enviado" (nil) de "campo enviado vacío". Al construir el `map[string]any` para GORM solo se incluyen los campos no-nil; retornar 400 si el map queda vacío.

**Respuestas HTTP**
- Todos los endpoints establecen `Content-Type: application/json`
- Los endpoints que reciben body limitan su tamaño a 1 MB con `http.MaxBytesReader`
- Los errores se devuelven con `writeError(w, status, msg)` definido en `handlers/errors.go`, que produce `{"error": "mensaje"}`

**Logging**
- Se usa `log/slog` con `slog.NewJSONHandler` (JSON estructurado a stdout)
- El logger se inyecta por constructor en repos y handlers — no hay logger global
- Usar pares clave-valor semánticos: `"id", user.ID`, `"error", err`
- El middleware usa `AsyncLogger` (canal buffereado): los requests no bloquean en IO de log
- Orden de shutdown: `srv.Shutdown` → `asyncLog.Stop()` → `rotator.Close()` — nunca invertir

**Documentación**
- Todo identificador exportado lleva un doc comment siguiendo [go.dev/doc/comment](https://go.dev/doc/comment): empieza con el nombre del identificador, oración completa con punto final
- Los paquetes llevan `// Package foo ...` en uno de sus archivos
- Usar `[Identifier]` para cross-links entre tipos del mismo paquete

**Servidor**
- Graceful shutdown con timeout de 5 s al recibir `SIGINT`/`SIGTERM`
- `AUTO_MIGRATE=true` ejecuta `db.AutoMigrate` al arrancar; mantenerlo en `false` en producción
- El bootstrap usa `sync.WaitGroup` para inicializar logger y DB en paralelo; usar `bootstrap` (logger stdout) hasta que `log` esté listo

**Autenticación JWT**
- `POST /auth/signup` y `POST /auth/login` son públicos; el resto de `/api/v1/` está protegido por `middleware.Auth`.
- `middleware.Auth` valida el Bearer token (HS256), rechaza tokens expirados y cualquier algoritmo distinto a HMAC, e inyecta el `userID` en el contexto bajo la clave privada `middleware.UserIDKey`.
- `JWT_SECRET` es obligatorio en producción; `JWT_EXPIRY_HOURS` controla la duración del token (default 24 h).
- Para cambiar contraseña usar `PATCH /api/v1/users/{id}/password` con `current_password` y `new_password`; la contraseña actual siempre se verifica con `bcrypt.CompareHashAndPassword` antes de actualizar.

**Updates selectivos con map**
- `UserRepository.Update` y `TaskRepository.Update` reciben `(ctx, entity, map[string]any)`. GORM genera `UPDATE … SET col=val WHERE id=?` solo para las columnas del map — `password_hash` y asociaciones no se tocan aunque no estén en el map.
- Esta firma es la que deben satisfacer los mocks en tests. El mock incluye nil-guard para no requerir `updateFn` en tests que no ejercen el update.

**Concurrencia — invariantes a mantener**
- Siempre verificar ausencia de data races con `go test -race ./...` tras agregar goroutines
- Los canales que pueden quedar sin lector deben ser buffereados (mínimo tamaño 1) para evitar goroutine leaks
- Escribir a `slice[idx]` desde goroutines es seguro solo si cada goroutine escribe a un índice único
