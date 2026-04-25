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

REST API en Go con tres capas:

```
cmd/api/main.go          ← wiring: crea config, DB, repos, handlers y registra rutas
internal/config/         ← lee variables de entorno y devuelve un Config struct
internal/database/       ← abre la conexión GORM y la devuelve (sin estado global)
internal/models/         ← structs de GORM (User, Task); User tiene []Task via foreignKey
internal/repository/     ← consultas GORM encapsuladas por entidad (UserRepository, TaskRepository)
internal/handlers/       ← HTTP handlers como métodos de struct; reciben su repo por inyección
```

**Flujo de dependencias:** `main` instancia `config.Load()` → `database.Connect(cfg)` → `repository.New*(db)` → `handlers.New*(repo)`. Ninguna capa conoce a la superior.

**Añadir un nuevo recurso** sigue este orden: modelo en `internal/models/`, repositorio en `internal/repository/`, handler en `internal/handlers/`, registro de rutas en `cmd/api/main.go`.

## Convenciones

- Archivos: `snake_case.go`
- El paquete `internal/` impide que módulos externos importen código de aplicación
- Los handlers establecen `Content-Type: application/json` y limitan el body a 1 MB (`http.MaxBytesReader`) en todos los endpoints que reciben body
- El servidor usa graceful shutdown con timeout de 5 s al recibir SIGINT/SIGTERM
