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

## Variables de entorno

| Variable        | Descripción                                        | Default   |
|-----------------|----------------------------------------------------|-----------|
| `DB_HOST`       | Host de la base de datos                           | localhost |
| `DB_USER`       | Usuario de PostgreSQL                              | postgres  |
| `DB_PASSWORD`   | Contraseña de PostgreSQL                           | —         |
| `DB_NAME`       | Nombre de la base de datos                         | mydb      |
| `DB_PORT`       | Puerto de PostgreSQL                               | 5432      |
| `DB_SSLMODE`    | Modo SSL de la conexión                            | disable   |
| `PORT`          | Puerto del servidor HTTP                           | 3000      |
| `AUTO_MIGRATE`  | Si `true`, ejecuta `db.AutoMigrate` al arrancar    | false     |

## Endpoints

### General

| Método | Ruta      | Descripción                              |
|--------|-----------|------------------------------------------|
| `GET`  | `/`       | Información general de la API (JSON)     |
| `GET`  | `/health` | Health check — verifica conexión a la BD |

### Usuarios

| Método   | Ruta                  | Descripción               |
|----------|-----------------------|---------------------------|
| `GET`    | `/api/v1/users`       | Listar todos los usuarios |
| `GET`    | `/api/v1/users/{id}`  | Obtener un usuario        |
| `POST`   | `/api/v1/users`       | Crear un usuario          |
| `PUT`    | `/api/v1/users/{id}`  | Actualizar un usuario     |
| `DELETE` | `/api/v1/users/{id}`  | Eliminar un usuario       |

### Tareas

| Método   | Ruta                  | Descripción              |
|----------|-----------------------|--------------------------|
| `GET`    | `/api/v1/tasks`       | Listar todas las tareas  |
| `GET`    | `/api/v1/tasks/{id}`  | Obtener una tarea        |
| `POST`   | `/api/v1/tasks`       | Crear una tarea          |
| `PUT`    | `/api/v1/tasks/{id}`  | Actualizar una tarea     |
| `DELETE` | `/api/v1/tasks/{id}`  | Eliminar una tarea       |

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

# Crear una tarea asociada al usuario con id 1
curl -X POST http://localhost:3000/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"learn gorm","description":"build web apps with gorm","user_id":1}'

# Listar usuarios con sus tareas
curl http://localhost:3000/api/v1/users
```

Las respuestas de error siguen el formato `{"error": "mensaje"}`.

> Los archivos `.http` en `request/` pueden usarse directamente desde VS Code con la extensión [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).

## Arquitectura

```
cmd/api/main.go          → entry point: wiring de dependencias y registro de rutas
internal/config/         → carga de variables de entorno (Config struct + Load)
internal/database/       → conexión a PostgreSQL via GORM
internal/models/         → entidades GORM (User, Task) con campos explícitos y soft-delete
internal/repository/     → consultas GORM encapsuladas por entidad
internal/handlers/       → handlers HTTP; cada uno define su propia interfaz de repositorio
internal/middleware/      → middleware de logging estructurado (log/slog)
```

**Flujo de dependencias:** `main` → `config.Load` → `database.Connect` → `repository.New*` → `handlers.New*`. Ninguna capa conoce a la superior.

**Decisiones de diseño relevantes:**

- Los handlers definen sus propias interfaces (`UserRepository`, `TaskRepository`) para desacoplarse de la implementación concreta y facilitar testing.
- Los DTOs de request (`createUserRequest`, `updateUserRequest`, etc.) evitan mass-assignment — los campos de auditoría (`id`, `created_at`) nunca se reciben del cliente.
- El body de las peticiones está limitado a 1 MB con `http.MaxBytesReader`.
- El servidor aplica graceful shutdown con timeout de 5 s al recibir `SIGINT`/`SIGTERM`.
- `AUTO_MIGRATE=true` activa la migración automática; en producción se deja en `false`.
- Todos los paquetes exportados siguen las convenciones de documentación de [go.dev/doc/comment](https://go.dev/doc/comment).
