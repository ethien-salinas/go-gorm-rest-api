# go-gorm-rest-api

REST API construida con Go como proyecto de aprendizaje. Implementa un CRUD de usuarios y tareas con autenticación JWT usando Gorilla Mux como router, GORM como ORM y PostgreSQL como base de datos.

## Stack

- **Go** — lenguaje principal
- **[Gorilla Mux](https://github.com/gorilla/mux)** — router HTTP
- **[GORM](https://gorm.io/)** — ORM con driver de PostgreSQL
- **PostgreSQL** — base de datos relacional
- **Docker Compose** — levanta la base de datos y Adminer localmente
- **[Air](https://github.com/air-verse/air)** — hot reload en desarrollo
- **log/slog** — logging estructurado en JSON (stdlib de Go 1.21)
- **[golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)** — generación y validación de JWT (HS256)
- **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** — hashing de contraseñas
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

# Verificar ausencia de data races
go test -race ./...
```

Los tests son unitarios puros — no requieren base de datos ni servidor levantado.

| Paquete | Cobertura |
|---|---|
| `internal/handlers` | 100 % |
| `internal/config` | 100 % |
| `internal/middleware` | 100 % |

Los handlers definen sus propias interfaces de repositorio, lo que permite crear mocks con structs de campos funcionales sin herramientas de generación.

## Variables de entorno

| Variable           | Descripción                                                        | Default   |
|--------------------|--------------------------------------------------------------------|-----------|
| `DB_HOST`          | Host de la base de datos                                           | localhost |
| `DB_USER`          | Usuario de PostgreSQL                                              | postgres  |
| `DB_PASSWORD`      | Contraseña de PostgreSQL                                           | —         |
| `DB_NAME`          | Nombre de la base de datos                                         | mydb      |
| `DB_PORT`          | Puerto de PostgreSQL                                               | 5432      |
| `DB_SSLMODE`       | Modo SSL de la conexión                                            | disable   |
| `PORT`             | Puerto del servidor HTTP                                           | 3000      |
| `AUTO_MIGRATE`     | Si `true`, ejecuta `db.AutoMigrate` al arrancar                    | false     |
| `LOG_TO_FILE`      | Si `true`, escribe logs a archivos diarios rotativos en `LOG_DIR`  | false     |
| `LOG_DIR`          | Directorio donde se almacenan los archivos de log                  | logs      |
| `JWT_SECRET`       | Clave de firma para los tokens JWT (HS256)                         | —         |
| `JWT_EXPIRY_HOURS` | Duración del token JWT en horas                                    | 24        |

## Endpoints

### General

| Método | Ruta      | Descripción                              | Auth |
|--------|-----------|------------------------------------------|------|
| `GET`  | `/`       | Información general de la API (JSON)     | No   |
| `GET`  | `/health` | Health check — verifica conexión a la BD | No   |

### Autenticación

| Método | Ruta            | Descripción                                        | Auth |
|--------|-----------------|----------------------------------------------------|------|
| `POST` | `/auth/signup`  | Registrar usuario con email y contraseña           | No   |
| `POST` | `/auth/login`   | Iniciar sesión; devuelve un JWT Bearer token       | No   |

### Usuarios

| Método    | Ruta                          | Descripción                                            | Auth |
|-----------|-------------------------------|--------------------------------------------------------|------|
| `GET`     | `/api/v1/users`               | Listar todos los usuarios                              | Sí   |
| `GET`     | `/api/v1/users/{id}`          | Obtener un usuario                                     | Sí   |
| `POST`    | `/api/v1/users`               | Crear un usuario                                       | Sí   |
| `POST`    | `/api/v1/users/batch`         | Crear múltiples usuarios en paralelo (worker pool)     | Sí   |
| `PATCH`   | `/api/v1/users/{id}`          | Actualizar solo los campos enviados                    | Sí   |
| `PATCH`   | `/api/v1/users/{id}/password` | Cambiar contraseña (requiere contraseña actual)        | Sí   |
| `DELETE`  | `/api/v1/users/{id}`          | Eliminar un usuario                                    | Sí   |

### Tareas

| Método   | Ruta                  | Descripción                             | Auth |
|----------|-----------------------|-----------------------------------------|------|
| `GET`    | `/api/v1/tasks`       | Listar todas las tareas                 | Sí   |
| `GET`    | `/api/v1/tasks/{id}`  | Obtener una tarea                       | Sí   |
| `POST`   | `/api/v1/tasks`       | Crear una tarea                         | Sí   |
| `PATCH`  | `/api/v1/tasks/{id}`  | Actualizar solo los campos enviados     | Sí   |
| `DELETE` | `/api/v1/tasks/{id}`  | Eliminar una tarea                      | Sí   |

### Estadísticas

| Método | Ruta            | Descripción                                                   | Auth |
|--------|-----------------|---------------------------------------------------------------|------|
| `GET`  | `/api/v1/stats` | Conteo de usuarios y tareas consultados en paralelo (fan-out) | Sí   |

### Ejemplos

```bash
# Registrar un usuario
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Tommy","last_name":"Lee","email":"tommy@example.com","password":"secret"}'

# Iniciar sesión y guardar el token
TOKEN=$(curl -s -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tommy@example.com","password":"secret"}' | jq -r .token)

# Usar el token en peticiones protegidas
curl http://localhost:3000/api/v1/users \
  -H "Authorization: Bearer $TOKEN"

# Actualizar solo el email (PATCH parcial — los demás campos no se tocan)
curl -X PATCH http://localhost:3000/api/v1/users/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"nuevo@example.com"}'

# Cambiar contraseña (requiere la contraseña actual)
curl -X PATCH http://localhost:3000/api/v1/users/1/password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"current_password":"secret","new_password":"newsecret"}'

# Marcar una tarea como completada (PATCH parcial)
curl -X PATCH http://localhost:3000/api/v1/tasks/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"done":true}'

# Crear múltiples usuarios en paralelo (worker pool, máx. 100 por lote)
curl -X POST http://localhost:3000/api/v1/users/batch \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"workers":3,"users":[{"first_name":"A","last_name":"B","email":"a@b.com"},{"first_name":"C","last_name":"D","email":"c@d.com"}]}'

# Estadísticas: conteo de usuarios y tareas en paralelo
curl http://localhost:3000/api/v1/stats \
  -H "Authorization: Bearer $TOKEN"
```

Las respuestas de error siguen el formato `{"error": "mensaje"}`.

El endpoint `/users/batch` responde `201 Created` si todas las inserciones tuvieron éxito, o `207 Multi-Status` si alguna falló.

> Los archivos `.http` en `request/` pueden usarse directamente desde VS Code con la extensión [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).

## Arquitectura

```
cmd/api/main.go          → entry point: wiring de dependencias y registro de rutas
internal/config/         → carga de variables de entorno (Config struct + Load)
internal/database/       → conexión a PostgreSQL via GORM
internal/models/         → entidades GORM (User, Task) con campos explícitos y soft-delete
internal/repository/     → consultas GORM encapsuladas por entidad
internal/handlers/       → handlers HTTP; cada uno define su propia interfaz de repositorio
internal/middleware/     → Auth (JWT) y Logging (AsyncLogger con canal buffereado)
internal/logger/         → DailyRotator: rotación diaria de logs a archivos con sync.Mutex
```

**Flujo de dependencias:** `main` → `config.Load` → `database.Connect` → `repository.New*` → `handlers.New*`. Ninguna capa conoce a la superior.

**Flujo de autenticación:**

```
POST /auth/signup  →  hashea password con bcrypt  →  crea User en BD
POST /auth/login   →  verifica hash  →  emite JWT {sub: userID, exp: now+24h}
GET  /api/v1/*     →  middleware Auth valida Bearer token  →  inyecta userID en contexto
```

**Patrones de concurrencia implementados:**

| Patrón | Dónde | Descripción |
|--------|-------|-------------|
| `sync.WaitGroup` | `cmd/api/main.go` | Bootstrap paralelo: logger y DB se inician en goroutines simultáneas |
| Canal buffereado + worker | `internal/middleware/async_logger.go` | Logging asíncrono: los requests no bloquean esperando IO de log |
| Fan-out con canales | `internal/handlers/stats.go` | Dos queries a BD corren en paralelo; canales buffer=1 evitan goroutine leaks |
| Semáforo + worker pool | `internal/repository/user.go` | `BatchCreate` limita la concurrencia con `chan struct{}` como semáforo |

**Decisiones de diseño relevantes:**

- Los handlers definen sus propias interfaces (`UserRepository`, `TaskRepository`) para desacoplarse de la implementación concreta y facilitar testing.
- Los DTOs de request son privados (`createUserRequest`, `updateUserRequest`, etc.) y usan punteros (`*string`, `*bool`) en los endpoints PATCH para distinguir "campo no enviado" (nil) de "campo enviado vacío". Esto evita borrar datos accidentalmente.
- `UserRepository.Update` y `TaskRepository.Update` reciben un `map[string]any` con solo las columnas a modificar; GORM genera un `UPDATE` selectivo que no toca `password_hash` ni asociaciones.
- El body de las peticiones está limitado a 1 MB con `http.MaxBytesReader`.
- El servidor aplica graceful shutdown con timeout de 5 s al recibir `SIGINT`/`SIGTERM`. El orden de cierre es: `srv.Shutdown` → `asyncLog.Stop` (flush del canal) → `rotator.Close`.
- `AUTO_MIGRATE=true` activa la migración automática; en producción se deja en `false`.
- Todos los paquetes exportados siguen las convenciones de documentación de [go.dev/doc/comment](https://go.dev/doc/comment).
