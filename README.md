# go-gorm-rest-api

REST API en Go con GORM y PostgreSQL. Implementa CRUD de usuarios y tareas con autenticación JWT, logging asíncrono con rotación diaria y patrones de concurrencia del stdlib.

---

## Tabla de contenidos

1. [Stack tecnológico](#1-stack-tecnológico)
2. [Estructura del proyecto](#2-estructura-del-proyecto)
3. [Requisitos previos](#3-requisitos-previos)
4. [Instalación y puesta en marcha](#4-instalación-y-puesta-en-marcha)
5. [Variables de entorno](#5-variables-de-entorno)
6. [Endpoints de la API](#6-endpoints-de-la-api)
7. [Autenticación JWT](#7-autenticación-jwt)
8. [Arquitectura y decisiones de diseño](#8-arquitectura-y-decisiones-de-diseño)
9. [Concurrencia](#9-concurrencia)
10. [Testing](#10-testing)
11. [Herramientas de desarrollo](#11-herramientas-de-desarrollo)

---

## 1. Stack tecnológico

| Herramienta | Rol en el proyecto |
|---|---|
| **Go 1.21+** | Lenguaje principal |
| **[Gorilla Mux](https://github.com/gorilla/mux)** | Router HTTP |
| **[GORM](https://gorm.io/)** | ORM para PostgreSQL |
| **PostgreSQL** | Base de datos relacional |
| **Docker Compose** | Levanta la BD y Adminer localmente |
| **[Air](https://github.com/air-verse/air)** | Hot reload en desarrollo |
| **`log/slog`** | Logging estructurado en JSON (stdlib Go 1.21) |
| **[golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)** | Generación y validación de JWT (HS256) |
| **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** | Hashing de contraseñas |
| **[golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)** | Token bucket para rate limiting por IP |
| **[testify](https://github.com/stretchr/testify)** | Aserciones en tests unitarios |

---

## 2. Estructura del proyecto

```
go-gorm-rest-api/
│
├── cmd/
│   └── api/
│       └── main.go          # Entry point: wiring de dependencias y arranque del servidor
│
├── internal/                # Código privado del módulo
│   ├── config/              # Carga de variables de entorno → struct Config
│   ├── database/            # Conexión a PostgreSQL via GORM
│   ├── models/              # Entidades GORM: User, Task
│   ├── repository/          # Queries a la BD, encapsuladas por entidad
│   ├── handlers/            # Handlers HTTP
│   ├── middleware/          # Auth JWT + Logging asíncrono + Rate Limiting por IP
│   └── logger/              # DailyRotator: rotación diaria de archivos de log
│
├── request/                 # Archivos .http para probar la API desde VS Code
│   ├── auth.http
│   ├── 1-user-request.http
│   └── 2-task-request.http
│
├── docs/                    # Documentación Swagger generada automáticamente (swag)
├── compose.yaml             # Docker Compose: PostgreSQL + Adminer
├── .env.example             # Plantilla de variables de entorno
└── .air.toml                # Config de Air (hot reload)
```

---

## 3. Requisitos previos

- [Go 1.21+](https://go.dev/dl/) — verifica con `go version`
- [Docker Desktop](https://www.docker.com/) — para levantar PostgreSQL y Adminer
- [Air](https://github.com/air-verse/air) _(opcional)_ — hot reload

```bash
# Instalar Air globalmente
go install github.com/air-verse/air@latest
```

---

## 4. Instalación y puesta en marcha

```bash
# 1. Clonar el repositorio
git clone https://github.com/ethien-salinas/go-gorm-rest-api.git
cd go-gorm-rest-api

# 2. Copiar y configurar las variables de entorno
cp .env.example .env
# Edita .env con tus valores (ver sección 5)

# 3. Levantar PostgreSQL y Adminer con Docker
docker compose up -d

# 4. Descargar dependencias
go mod tidy

# 5. Ejecutar el servidor
air             # con hot reload (recomendado en desarrollo)
# o
go run ./cmd/api
```

El servidor corre en `http://localhost:3000`.  
Adminer (cliente web de PostgreSQL) disponible en `http://localhost:8080`.

---

## 5. Variables de entorno

Copia `.env.example` a `.env` y completa los valores:

```bash
cp .env.example .env
```

| Variable | Descripción | Default |
|---|---|---|
| `DB_HOST` | Host de la base de datos | `localhost` |
| `DB_USER` | Usuario de PostgreSQL | `postgres` |
| `DB_PASSWORD` | Contraseña de PostgreSQL | — |
| `DB_NAME` | Nombre de la base de datos | `mydb` |
| `DB_PORT` | Puerto de PostgreSQL | `5432` |
| `DB_SSLMODE` | Modo SSL de la conexión | `disable` |
| `PORT` | Puerto del servidor HTTP | `3000` |
| `AUTO_MIGRATE` | Si `true`, ejecuta `db.AutoMigrate` al arrancar | `false` |
| `LOG_TO_FILE` | Si `true`, escribe logs a archivos diarios rotativos en `LOG_DIR` | `false` |
| `LOG_DIR` | Directorio donde se almacenan los archivos de log | `logs` |
| `JWT_SECRET` | Clave de firma para los tokens JWT (HS256) — **no commitear** | — |
| `JWT_EXPIRY_HOURS` | Duración del token JWT en horas | `24` |
| `RATE_LIMIT_RPS` | Peticiones por segundo permitidas por IP (token bucket) | `10` |
| `RATE_LIMIT_BURST` | Capacidad máxima del bucket — permite ráfagas cortas sobre el RPS | `20` |

> **`AUTO_MIGRATE`**: Cuando está en `true`, GORM crea/actualiza las tablas automáticamente al arrancar. Útil en desarrollo; en producción se recomienda `false` y gestionar migraciones de forma explícita.

---

## 6. Endpoints de la API

### General

| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| `GET` | `/` | Información general de la API (JSON) | No |
| `GET` | `/health` | Health check — verifica conexión a la BD | No |
| `GET` | `/swagger/index.html` | Documentación interactiva Swagger UI | No |

### Autenticación

| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| `POST` | `/auth/signup` | Registrar nuevo usuario | No |
| `POST` | `/auth/login` | Iniciar sesión — devuelve JWT | No |

### Usuarios (`/api/v1/users`)

| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| `GET` | `/api/v1/users` | Listar todos los usuarios | ✅ |
| `GET` | `/api/v1/users/{id}` | Obtener un usuario por ID | ✅ |
| `POST` | `/api/v1/users` | Crear un usuario | ✅ |
| `POST` | `/api/v1/users/batch` | Crear múltiples usuarios en paralelo | ✅ |
| `PATCH` | `/api/v1/users/{id}` | Actualizar solo los campos enviados | ✅ |
| `PATCH` | `/api/v1/users/{id}/password` | Cambiar contraseña | ✅ |
| `DELETE` | `/api/v1/users/{id}` | Eliminar un usuario | ✅ |

### Tareas (`/api/v1/tasks`)

| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| `GET` | `/api/v1/tasks` | Listar todas las tareas | ✅ |
| `GET` | `/api/v1/tasks/{id}` | Obtener una tarea | ✅ |
| `POST` | `/api/v1/tasks` | Crear una tarea | ✅ |
| `PATCH` | `/api/v1/tasks/{id}` | Actualizar solo los campos enviados | ✅ |
| `DELETE` | `/api/v1/tasks/{id}` | Eliminar una tarea | ✅ |

### Estadísticas

| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| `GET` | `/api/v1/stats` | Conteo de usuarios y tareas en paralelo | ✅ |

### Ejemplos con `curl`

```bash
# 1. Registrar un usuario
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Tommy","last_name":"Lee","email":"tommy@example.com","password":"secret"}'

# 2. Iniciar sesión y capturar el token (requiere jq)
TOKEN=$(curl -s -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tommy@example.com","password":"secret"}' | jq -r .token)

# 3. Listar usuarios (ruta protegida)
curl http://localhost:3000/api/v1/users \
  -H "Authorization: Bearer $TOKEN"

# 4. Actualizar solo el email (PATCH parcial)
curl -X PATCH http://localhost:3000/api/v1/users/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"nuevo@example.com"}'

# 5. Cambiar contraseña
curl -X PATCH http://localhost:3000/api/v1/users/1/password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"current_password":"secret","new_password":"newsecret"}'

# 6. Marcar tarea como completada
curl -X PATCH http://localhost:3000/api/v1/tasks/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"done":true}'

# 7. Crear múltiples usuarios en paralelo (worker pool)
curl -X POST http://localhost:3000/api/v1/users/batch \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"workers":3,"users":[{"first_name":"A","last_name":"B","email":"a@b.com"},{"first_name":"C","last_name":"D","email":"c@d.com"}]}'

# 8. Estadísticas
curl http://localhost:3000/api/v1/stats \
  -H "Authorization: Bearer $TOKEN"
```

**Formato de errores:** todas las respuestas de error siguen `{"error": "mensaje"}`.

**Endpoint `/users/batch`:** responde `201 Created` si todas las inserciones tuvieron éxito, o `207 Multi-Status` si alguna falló.

Los archivos `.http` en `request/` pueden ejecutarse directamente desde VS Code con la extensión [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).

---

## 7. Autenticación JWT

```
POST /auth/signup  →  hashea password con bcrypt  →  crea User en BD
POST /auth/login   →  verifica hash bcrypt         →  emite JWT { sub: userID, exp: now+Nh }
GET  /api/v1/*     →  middleware valida Bearer      →  inyecta userID en el contexto del request
```

El token debe enviarse en el header `Authorization`:

```
Authorization: Bearer <token>
```

---

## 8. Arquitectura y decisiones de diseño

### Flujo de dependencias

```
main.go
  └─> config.Load()           # lee .env → struct Config
  └─> database.Connect()      # abre conexión GORM
  └─> repository.New*()       # capa de acceso a datos
  └─> handlers.New*()         # handlers HTTP
  └─> mux.Router              # registro de rutas
  └─> http.Server             # arranca el servidor
```

Ninguna capa conoce a la superior. Los handlers no importan `repository` directamente — solo conocen una **interfaz** definida en el mismo archivo del handler:

```go
type UserRepository interface {
    FindAll() ([]models.User, error)
    FindByID(id uint) (*models.User, error)
    // ...
}
```

### Decisiones técnicas relevantes

| Decisión | Razonamiento |
|---|---|
| **PATCH usa punteros** (`*string`, `*bool`) | Distingue "campo no enviado" (`nil`) de "campo enviado vacío" (`""`). Evita borrar datos accidentalmente. |
| **`Update` recibe `map[string]any`** | GORM genera un `UPDATE` selectivo que solo toca las columnas enviadas. `password_hash` nunca se sobreescribe por error. |
| **Body limitado a 1 MB** | `http.MaxBytesReader` protege contra payloads gigantes. |
| **Graceful shutdown** | Al recibir `SIGINT`/`SIGTERM`, el servidor espera 5 s a que los requests en vuelo terminen antes de cerrar. El orden `srv.Shutdown` → `rateLimiter.Stop()` → `asyncLog.Stop()` → `rotator.Close()` es invariante. |
| **Interfaces por handler** | Rompe la dependencia directa al repositorio concreto y facilita el testing con mocks manuales. |

---

## 9. Concurrencia

El proyecto usa cuatro patrones del stdlib de Go:

| Patrón | Dónde | Qué hace |
|---|---|---|
| **Bootstrap paralelo** | `cmd/api/main.go` | `WaitGroup` para iniciar el logger y la BD concurrentemente al arrancar |
| **Fan-out** | `internal/handlers/stats.go` | Dos queries a la BD corren en goroutines; canales con buffer=1 recolectan los resultados |
| **Worker pool** | `internal/repository/user.go:BatchCreate` | Semáforo `chan struct{}` de tamaño N limita las goroutines que escriben en la BD simultáneamente |
| **Async worker** | `internal/middleware/async_logger.go` | Canal buffereado + goroutine consumidora desacoplan el log del path de la request |
| **Background ticker** | `internal/middleware/rate_limiter.go` | Goroutine con `time.Ticker` limpia cada minuto los visitantes inactivos del map de limiters |

---

## 10. Testing

```bash
# Correr todos los tests
go test ./...

# Con salida detallada y cobertura por paquete
go test -v -cover ./...

# Detectar data races
go test -race ./...
```

Los tests son **unitarios puros** — no requieren base de datos ni servidor levantado.

| Paquete | Cobertura |
|---|---|
| `internal/handlers` | 100 % |
| `internal/config` | 100 % |
| `internal/middleware` | 100 % |

Los mocks son structs manuales que implementan la interfaz del repositorio — sin librerías de mocking:

```go
type mockUserRepo struct {
    users []models.User
    err   error
}

func (m *mockUserRepo) FindAll() ([]models.User, error) {
    return m.users, m.err
}
```

---

## 11. Herramientas de desarrollo

### Air — Hot Reload

```bash
air
```

La configuración está en `.air.toml`. Monitorea cambios en archivos `.go` y recompila automáticamente.

### Adminer — Cliente web de PostgreSQL

Disponible en `http://localhost:8080` cuando Docker está corriendo.

- **Sistema:** PostgreSQL
- **Servidor:** `db` (nombre del servicio en `compose.yaml`)
- **Usuario/contraseña/BD:** los que configuraste en `.env`

### Swagger UI — Documentación interactiva

Disponible en `http://localhost:3000/swagger/index.html`.

Para regenerar la documentación si modificas los comentarios Swagger:

```bash
# Instalar swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generar documentación
swag init -g cmd/api/main.go
```

### REST Client — Archivos `.http`

La carpeta `request/` contiene archivos `.http` con ejemplos de todas las peticiones. Instala la extensión [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) en VS Code para ejecutarlos directamente.

```http
# Declarar una variable reutilizable
@token = eyJhbGci...

# Usar la variable
GET http://localhost:3000/api/v1/users HTTP/1.1
Authorization: Bearer {{token}}
```
