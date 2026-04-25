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
internal/database/       ← abre la conexión GORM y la devuelve (sin estado global)
internal/models/         ← structs de GORM (User, Task); User tiene []Task via foreignKey
internal/repository/     ← consultas GORM encapsuladas por entidad (UserRepository, TaskRepository)
internal/handlers/       ← HTTP handlers como métodos de struct; reciben su repo por inyección
internal/middleware/     ← middleware HTTP (actualmente: Logging con log/slog)
```

**Flujo de dependencias:** `main` instancia `config.Load()` → `database.Connect(cfg)` → `repository.New*(db)` → `handlers.New*(repo)`. Ninguna capa conoce a la superior.

**Rutas registradas:**

```
GET  /              → handlers.HomeHandler         (info JSON de la API)
GET  /health        → handlers.NewHealthHandler     (ping a la BD)
GET  /api/v1/users
GET  /api/v1/users/{id}
POST /api/v1/users
PUT  /api/v1/users/{id}
DELETE /api/v1/users/{id}
GET  /api/v1/tasks
GET  /api/v1/tasks/{id}
POST /api/v1/tasks
PUT  /api/v1/tasks/{id}
DELETE /api/v1/tasks/{id}
```

**Añadir un nuevo recurso** sigue este orden:
1. Modelo en `internal/models/` con campos explícitos (sin `gorm.Model`) y json tags en snake_case
2. Repositorio en `internal/repository/` con métodos `FindAll`, `FindByID`, `Create`, `Update`, `Delete`
3. Handler en `internal/handlers/`: definir primero la interfaz del repositorio en el mismo archivo, luego el struct handler, constructor y métodos
4. Registro de rutas en `cmd/api/main.go` bajo el subrouter `/api/v1`

## Convenciones

**Estructura y nombrado**
- Archivos: `snake_case.go`
- El paquete `internal/` impide que módulos externos importen código de aplicación

**Interfaces de repositorio**
- Cada handler define su propia interfaz (e.g., `UserRepository` en `handlers/user.go`), no importa el tipo concreto del paquete `repository`. Esto desacopla las capas y facilita el testing con mocks.

**DTOs de request**
- Nunca usar el modelo directamente para leer el body. Definir structs privados (`createXRequest`, `updateXRequest`) con solo los campos que el cliente puede enviar. Esto evita mass-assignment de campos de auditoría (`id`, `created_at`, `deleted_at`).

**Respuestas HTTP**
- Todos los endpoints establecen `Content-Type: application/json`
- Los endpoints que reciben body limitan su tamaño a 1 MB con `http.MaxBytesReader`
- Los errores se devuelven con `writeError(w, status, msg)` definido en `handlers/errors.go`, que produce `{"error": "mensaje"}`

**Logging**
- Se usa `log/slog` con `slog.NewJSONHandler` (JSON estructurado a stdout)
- El logger se inyecta por constructor en repos y handlers — no hay logger global
- Usar pares clave-valor semánticos: `"id", user.ID`, `"error", err`

**Documentación**
- Todo identificador exportado lleva un doc comment siguiendo [go.dev/doc/comment](https://go.dev/doc/comment): empieza con el nombre del identificador, oración completa con punto final
- Los paquetes llevan `// Package foo ...` en uno de sus archivos
- Usar `[Identifier]` para cross-links entre tipos del mismo paquete

**Servidor**
- Graceful shutdown con timeout de 5 s al recibir `SIGINT`/`SIGTERM`
- `AUTO_MIGRATE=true` ejecuta `db.AutoMigrate` al arrancar; mantenerlo en `false` en producción
