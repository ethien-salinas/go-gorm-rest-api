# go-gorm-rest-api

REST API construida con Go como proyecto de aprendizaje. Implementa un CRUD de usuarios y tareas usando Gorilla Mux como router, GORM como ORM y PostgreSQL como base de datos.

## Stack

- **Go** — lenguaje principal
- **[Gorilla Mux](https://github.com/gorilla/mux)** — router HTTP
- **[GORM](https://gorm.io/)** — ORM con driver de PostgreSQL
- **PostgreSQL** — base de datos relacional
- **Docker Compose** — levanta la base de datos y Adminer localmente
- **[Air](https://github.com/air-verse/air)** — hot reload en desarrollo
- **log/slog** — logging estructurado en JSON (stdlib de Go 1.21)
- **[testify](https://github.com/stretchr/testify)** — aserciones en tests unitarios

## Requisitos

- [Go 1.21+](https://go.dev/dl/)
- [Docker](https://www.docker.com/)
- [Air](https://github.com/air-verse/air) _(opcional, para hot reload)_

```bash
go install github.com/air-verse/air@latest
```

## Instalación y uso

```bash
# 1. Clonar el repositorio
git clone https://github.com/ethien-salinas/go-gorm-rest-api.git
cd go-gorm-rest-api

# 2. Configurar variables de entorno
cp .env.example .env

# 3. Levantar la base de datos
docker compose up -d

# 4. Instalar dependencias
go mod tidy

# 5. Correr el servidor
air           # con hot reload
# o
go run ./cmd/api
```

El servidor corre en `http://localhost:3000`.

Adminer (cliente web de PostgreSQL) disponible en `http://localhost:8080`.

## Tests

```bash
# Correr todos los tests
go test ./...

# Con salida verbose y cobertura por paquete
go test -v -cover ./...
```

Los tests son unitarios puros — no requieren base de datos ni servidor levantado.

| Paquete | Cobertura |
|---|---|
| `internal/handlers` | 100 % |
| `internal/config` | 100 % |
| `internal/middleware` | 100 % |

Los handlers definen sus propias interfaces de repositorio, lo que permite crear mocks con structs de campos funcionales sin herramientas de generación. El health handler usa una interfaz `Pinger` que `*sql.DB` satisface de forma nativa, eliminando la dependencia de `*gorm.DB` en los tests.

## Variables de entorno

| Variable        | Descripción                                                        | Default   |
|-----------------|--------------------------------------------------------------------|-----------|
| `DB_HOST`       | Host de la base de datos                                           | localhost |
| `DB_USER`       | Usuario de PostgreSQL                                              | postgres  |
| `DB_PASSWORD`   | Contraseña de PostgreSQL                                           | —         |
| `DB_NAME`       | Nombre de la base de datos                                         | mydb      |
| `DB_PORT`       | Puerto de PostgreSQL                                               | 5432      |
| `DB_SSLMODE`    | Modo SSL de la conexión                                            | disable   |
| `PORT`          | Puerto del servidor HTTP                                           | 3000      |
| `AUTO_MIGRATE`  | Si `true`, ejecuta `db.AutoMigrate` al arrancar                    | false     |
| `LOG_TO_FILE`   | Si `true`, escribe logs a archivos diarios rotativos en `LOG_DIR`  | false     |
| `LOG_DIR`       | Directorio donde se almacenan los archivos de log                  | logs      |

## Endpoints

### General

| Método | Ruta      | Descripción                              |
|--------|-----------|------------------------------------------|
| `GET`  | `/`       | Información general de la API (JSON)     |
| `GET`  | `/health` | Health check — verifica conexión a la BD |

### Usuarios

| Método   | Ruta                      | Descripción                                                          |
|----------|---------------------------|----------------------------------------------------------------------|
| `GET`    | `/api/v1/users`           | Listar todos los usuarios                                            |
| `GET`    | `/api/v1/users/{id}`      | Obtener un usuario                                                   |
| `POST`   | `/api/v1/users`           | Crear un usuario                                                     |
| `POST`   | `/api/v1/users/batch`     | Crear múltiples usuarios en paralelo (worker pool)                   |
| `PUT`    | `/api/v1/users/{id}`      | Actualizar un usuario                                                |
| `DELETE` | `/api/v1/users/{id}`      | Eliminar un usuario                                                  |

### Tareas

| Método   | Ruta                  | Descripción              |
|----------|-----------------------|--------------------------|
| `GET`    | `/api/v1/tasks`       | Listar todas las tareas  |
| `GET`    | `/api/v1/tasks/{id}`  | Obtener una tarea        |
| `POST`   | `/api/v1/tasks`       | Crear una tarea          |
| `PUT`    | `/api/v1/tasks/{id}`  | Actualizar una tarea     |
| `DELETE` | `/api/v1/tasks/{id}`  | Eliminar una tarea       |

### Estadísticas

| Método | Ruta              | Descripción                                                      |
|--------|-------------------|------------------------------------------------------------------|
| `GET`  | `/api/v1/stats`   | Conteo de usuarios y tareas consultados en paralelo (fan-out)    |

### Ejemplos

```bash
# Info de la API
curl http://localhost:3000/

# Health check
curl http://localhost:3000/health

# Crear un usuario
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"first_name":"tommy","last_name":"lee","email":"tommy.lee@example.com"}'

# Crear múltiples usuarios en paralelo (worker pool, máx. 100 por lote)
# workers indica cuántos inserts corren simultáneamente (default 5)
curl -X POST http://localhost:3000/api/v1/users/batch \
  -H "Content-Type: application/json" \
  -d '{"workers":3,"users":[{"first_name":"A","last_name":"B","email":"a@b.com"},{"first_name":"C","last_name":"D","email":"c@d.com"}]}'

# Crear una tarea asociada al usuario con id 1
curl -X POST http://localhost:3000/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"learn gorm","description":"build web apps with gorm","user_id":1}'

# Listar usuarios con sus tareas
curl http://localhost:3000/api/v1/users

# Estadísticas: conteo de usuarios y tareas en paralelo
curl http://localhost:3000/api/v1/stats
```

Las respuestas de error siguen el formato `{"error": "mensaje"}`.

El endpoint `/users/batch` responde `201 Created` si todas las inserciones tuvieron éxito, o `207 Multi-Status` si alguna falló, incluyendo el campo `"error"` en el resultado individual.

> Los archivos `.http` en `request/` pueden usarse directamente desde VS Code con la extensión [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).

## Arquitectura

```
cmd/api/main.go          → entry point: wiring de dependencias y registro de rutas
internal/config/         → carga de variables de entorno (Config struct + Load)
internal/database/       → conexión a PostgreSQL via GORM
internal/models/         → entidades GORM (User, Task) con campos explícitos y soft-delete
internal/repository/     → consultas GORM encapsuladas por entidad
internal/handlers/       → handlers HTTP; cada uno define su propia interfaz de repositorio
internal/middleware/      → middleware HTTP: logging estructurado con AsyncLogger (canal buffereado)
internal/logger/          → DailyRotator: rotación diaria de logs a archivos con sync.Mutex
```

**Flujo de dependencias:** `main` → `config.Load` → `database.Connect` → `repository.New*` → `handlers.New*`. Ninguna capa conoce a la superior.

**Patrones de concurrencia implementados:**

| Patrón | Dónde | Descripción |
|--------|-------|-------------|
| `sync.WaitGroup` | `cmd/api/main.go` | Bootstrap paralelo: logger y DB se inician en goroutines simultáneas |
| Canal buffereado + worker | `internal/middleware/async_logger.go` | Logging asíncrono: los requests no bloquean esperando IO de log |
| Fan-out con canales | `internal/handlers/stats.go` | Dos queries a BD corren en paralelo; canales buffer=1 evitan goroutine leaks |
| Semáforo + worker pool | `internal/repository/user.go` | `BatchCreate` limita la concurrencia con `chan struct{}` como semáforo |

**Decisiones de diseño relevantes:**

- Los handlers definen sus propias interfaces (`UserRepository`, `TaskRepository`) para desacoplarse de la implementación concreta y facilitar testing.
- Los DTOs de request (`createUserRequest`, `updateUserRequest`, etc.) evitan mass-assignment — los campos de auditoría (`id`, `created_at`) nunca se reciben del cliente.
- El body de las peticiones está limitado a 1 MB con `http.MaxBytesReader`.
- El servidor aplica graceful shutdown con timeout de 5 s al recibir `SIGINT`/`SIGTERM`. El orden de cierre es: `srv.Shutdown` → `asyncLog.Stop` (flush del canal) → `rotator.Close`.
- `AUTO_MIGRATE=true` activa la migración automática; en producción se deja en `false`.
- `database.Connect` retorna `(*gorm.DB, error)` — no llama `os.Exit` internamente; la decisión de terminar es responsabilidad del caller.
- La interfaz `Pinger` en `handlers/health.go` desacopla el health check de `*gorm.DB`; `*sql.DB` la satisface de forma nativa y permite testar el endpoint sin base de datos real.
- `StatsHandler` define interfaces propias `UserCounter`/`TaskCounter` con solo el método `Count` (interface segregation); no reutiliza las interfaces CRUD.
- Todos los paquetes exportados siguen las convenciones de documentación de [go.dev/doc/comment](https://go.dev/doc/comment).
