# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
air                          # desarrollo con hot reload
go run ./cmd/api             # sin hot reload
go build -o ./tmp/main ./cmd/api
go mod tidy
go test -race ./...          # siempre usar -race al agregar goroutines
docker compose up -d         # levanta PostgreSQL + Adminer
```

Copiar `.env.example` a `.env` antes de correr. Variables cargadas automáticamente por godotenv.

## Architecture

```
cmd/api/main.go      ← wiring: config → DB → repos → handlers → rutas
internal/config/     ← env vars → Config struct
internal/database/   ← abre conexión GORM; retorna error, no llama os.Exit
internal/models/     ← structs GORM (User, Task)
internal/repository/ ← queries GORM por entidad
internal/handlers/   ← HTTP handlers; cada uno define su propia interfaz de repo
internal/middleware/ ← Auth (JWT) + Logging (AsyncLogger) + RateLimiter (token bucket per-IP)
internal/logger/     ← DailyRotator: rotación diaria con sync.Mutex
```

**Flujo de dependencias:** `config.Load` → bootstrap paralelo (logger + DB con WaitGroup) → `repository.New*` → `handlers.New*`. Ninguna capa conoce a la superior.

## Rutas

```
GET  /          GET  /health

POST /auth/signup    POST /auth/login          ← públicas

— /api/v1 requiere Bearer JWT —
GET/POST        /api/v1/users
POST            /api/v1/users/batch            ← precede a /{id} automáticamente (net/http 1.22+)
GET/PATCH/DELETE /api/v1/users/{id}
PATCH           /api/v1/users/{id}/password
GET/POST        /api/v1/tasks
GET/PATCH/DELETE /api/v1/tasks/{id}
GET             /api/v1/stats
```

## Convenciones críticas

**Añadir un recurso nuevo:**
1. Modelo en `internal/models/` — campos explícitos, sin `gorm.Model`, json tags snake_case
2. Repositorio en `internal/repository/` — `FindAll`, `FindByID`, `Create`, `Update(ctx, entity, map[string]any)`, `Delete`
3. Handler en `internal/handlers/` — definir la interfaz del repo en el mismo archivo, luego struct + constructor + métodos
4. Registrar rutas en `cmd/api/main.go` — rutas protegidas con `chain(http.HandlerFunc(...), protected...)`, públicas con `chain(..., global...)`

**PATCH parcial:** los DTOs usan `*string`/`*bool` — nil significa campo no enviado. Construir `map[string]any` con solo los campos no-nil y pasarlo a `repo.Update`; retornar 400 si el map queda vacío. GORM genera `UPDATE … SET col=val` solo para esas columnas; `password_hash` y asociaciones no se tocan.

**JWT:** `middleware.Auth` valida el Bearer token (HS256), rechaza expirados y algoritmos no-HMAC, e inyecta `userID` en contexto bajo `middleware.UserIDKey`. `/auth/signup` y `/auth/login` son las únicas rutas públicas bajo la raíz.

**Shutdown:** el orden `srv.Shutdown` → `rateLimiter.Stop()` → `asyncLog.Stop()` → `rotator.Close()` es invariante — nunca invertir.

**Concurrencia — patrones disponibles:**
- Fan-out: `handlers/stats.go` — goroutines + canales buffer=1
- Worker pool: `repository/user.go:BatchCreate` — semáforo `chan struct{}`
- Async worker: `middleware/async_logger.go` — canal buffereado + goroutine consumer
- Bootstrap paralelo: `cmd/api/main.go` — WaitGroup para IO independiente
- Background ticker: `middleware/rate_limiter.go` — goroutine con `time.Ticker` limpia visitors inactivos cada minuto

**Rate limiting:** `middleware.RateLimiter` aplica token bucket per-IP en el router principal (antes del logger). Configurable con `RATE_LIMIT_RPS` (default 10 req/s) y `RATE_LIMIT_BURST` (default 20). Responde `429` con `{"error":"too many requests"}`. Instanciar con `NewRateLimiter`, registrar con `.Middleware()`, detener con `.Stop()` en shutdown.
