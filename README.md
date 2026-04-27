# go-gorm-rest-api

REST API en Go con GORM y PostgreSQL. Implementa CRUD de usuarios y tareas con autenticación JWT, política de contraseñas de grado financiero (PCI DSS v4.0 + NIST SP 800-63B), logging asíncrono con rotación diaria y patrones de concurrencia del stdlib.

---

## Tabla de contenidos

1. [Stack tecnológico](#1-stack-tecnológico)
2. [Estructura del proyecto](#2-estructura-del-proyecto)
3. [Requisitos previos](#3-requisitos-previos)
4. [Instalación y puesta en marcha](#4-instalación-y-puesta-en-marcha)
5. [Variables de entorno](#5-variables-de-entorno)
6. [Endpoints de la API](#6-endpoints-de-la-api)
7. [Autenticación JWT](#7-autenticación-jwt)
8. [Política de contraseñas](#8-política-de-contraseñas)
9. [Arquitectura y decisiones de diseño](#9-arquitectura-y-decisiones-de-diseño)
10. [Concurrencia](#10-concurrencia)
11. [Testing](#11-testing)
12. [Herramientas de desarrollo](#12-herramientas-de-desarrollo)

---

## 1. Stack tecnológico

| Herramienta | Rol en el proyecto |
|---|---|
| **Go 1.22+** | Lenguaje principal |
| **net/http** | Router HTTP (stdlib, Go 1.22+) |
| **[GORM](https://gorm.io/)** | ORM para PostgreSQL |
| **PostgreSQL** | Base de datos relacional |
| **Docker Compose** | Levanta la BD y Adminer localmente |
| **[Air](https://github.com/air-verse/air)** | Hot reload en desarrollo |
| **`log/slog`** | Logging estructurado en JSON (stdlib Go 1.21) |
| **[golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)** | Generación y validación de JWT (HS256) |
| **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** | Hashing de contraseñas (cost 12) |
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
│   ├── models/              # Entidades GORM: User, Task, PasswordHistory
│   ├── password/            # Política de contraseñas: Validate(), BcryptCost, lista común embebida
│   ├── repository/          # Queries a la BD, encapsuladas por entidad (User, Task, PasswordHistory)
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

- [Go 1.22+](https://go.dev/dl/) — verifica con `go version`
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
| `PASSWORD_HISTORY_COUNT` | Contraseñas anteriores que no se pueden reutilizar (PCI DSS 8.3.7) | `12` |
| `PASSWORD_MAX_AGE_DAYS` | Días antes de forzar cambio de contraseña (reservado) | `90` |
| `ACCOUNT_LOCKOUT_THRESHOLD` | Intentos fallidos antes de bloquear la cuenta (PCI DSS 8.3.4) | `5` |
| `ACCOUNT_LOCKOUT_DURATION_MINUTES` | Minutos que dura el bloqueo | `30` |

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
# 1. Registrar un usuario (la contraseña debe cumplir la política)
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Tommy","last_name":"Lee","email":"tommy@example.com","password":"Str0ng!Pass#24"}'

# 2. Iniciar sesión y capturar el token (requiere jq)
TOKEN=$(curl -s -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tommy@example.com","password":"Str0ng!Pass#24"}' | jq -r .token)

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
  -d '{"current_password":"Str0ng!Pass#24","new_password":"N3wP@ssw0rd#99"}'

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

**Formato de errores simples:** `{"error": "mensaje"}`.

**Errores de validación de contraseña:** `{"errors": ["mensaje 1", "mensaje 2", ...]}` (se reportan todas las reglas violadas en una sola respuesta).

**Endpoint `/users/batch`:** responde `201 Created` si todas las inserciones tuvieron éxito, o `207 Multi-Status` si alguna falló.

Los archivos `.http` en `request/` pueden ejecutarse directamente desde VS Code con la extensión [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).

---

## 7. Autenticación JWT

```
POST /auth/signup  →  valida política  →  hashea password (bcrypt cost 12)  →  crea User en BD
POST /auth/login   →  verifica bloqueo →  verifica hash bcrypt               →  emite JWT { sub: userID, exp: now+Nh }
GET  /api/v1/*     →  middleware valida Bearer  →  inyecta userID en el contexto del request
```

El token debe enviarse en el header `Authorization`:

```
Authorization: Bearer <token>
```

---

## 8. Política de contraseñas

La API cumple **PCI DSS v4.0** y **NIST SP 800-63B**. Toda la lógica reside en el paquete `internal/password`, sin dependencias de handlers ni repositorios (capa de dominio pura).

### Reglas de fortaleza (signup y cambio de contraseña)

| Regla | Requisito |
|---|---|
| Longitud mínima | 12 caracteres (PCI DSS 8.3.6) |
| Longitud máxima | 128 caracteres (NIST) |
| Complejidad | Al menos una mayúscula, una minúscula, un dígito y un carácter especial |
| Contraseñas comunes | Rechazadas si aparecen en la lista top-100 embebida (NIST) |
| Contiene email | Rechazada si incluye la parte local del correo del usuario |

Cuando la contraseña viola una o más reglas, la respuesta incluye **todas** las violaciones a la vez:

```json
{
  "errors": [
    "la contraseña debe tener al menos 12 caracteres",
    "la contraseña debe contener al menos un carácter especial"
  ]
}
```

### Historial de contraseñas (PCI DSS 8.3.7)

El endpoint `PATCH /api/v1/users/{id}/password` compara la nueva contraseña contra los últimos **12 hashes** almacenados en la tabla `password_histories`. Si coincide con alguno, la solicitud se rechaza con `400`.

### Bloqueo de cuenta (PCI DSS 8.3.4)

Tras **5 intentos fallidos** de login, la cuenta se bloquea durante **30 minutos**. El bloqueo se gestiona con los campos `failed_login_count` y `locked_until` del modelo `User`:

- Cada intento fallido incrementa `failed_login_count` y persiste el cambio en la BD.
- Al alcanzar el umbral, se registra la fecha y hora de desbloqueo en `locked_until`.
- Un login exitoso resetea ambos campos a cero.

Los valores por defecto son configurables vía variables de entorno (ver sección 5).

---

## 9. Arquitectura y decisiones de diseño

### Flujo de dependencias

```
main.go
  └─> config.Load()                # lee .env → struct Config
  └─> database.Connect()           # abre conexión GORM
  └─> repository.New*()            # capa de acceso a datos
  └─> handlers.New*()              # handlers HTTP
  └─> http.ServeMux                # registro de rutas con chain() middleware
  └─> http.Server                  # arranca el servidor
```

El paquete `internal/password` no aparece en el flujo de wiring porque no tiene estado — los handlers lo invocan directamente como una función pura.

Ninguna capa conoce a la superior. Los handlers no importan `repository` directamente — solo conocen una **interfaz** definida en el mismo archivo del handler:

```go
type UserRepository interface {
    FindAll(ctx context.Context) ([]models.User, error)
    FindByID(ctx context.Context, id string) (models.User, error)
    // ...
}
```

### Decisiones técnicas relevantes

| Decisión | Razonamiento |
|---|---|
| **`internal/password` sin estado** | Capa de dominio pura: `Validate()` y `BcryptCost` no dependen de la BD ni de los handlers. Facilita el testing en aislamiento. |
| **Lista de contraseñas comunes embebida** | `//go:embed common_passwords.txt` → binario autocontenido, sin I/O en runtime. Cargada con `sync.Once`. |
| **`BcryptCost = 12`** | Única fuente de verdad para el costo de hashing. Todos los handlers lo importan desde `internal/password`. |
| **PATCH usa punteros** (`*string`, `*bool`) | Distingue "campo no enviado" (`nil`) de "campo enviado vacío" (`""`). Evita borrar datos accidentalmente. |
| **`Update` recibe `map[string]any`** | GORM genera un `UPDATE` selectivo que solo toca las columnas enviadas. `password_hash` nunca se sobreescribe por error. |
| **`map[string]any` para actualizar `locked_until`** | A diferencia de structs, GORM sí persiste valores `nil` en maps — lo que permite resetear la columna nullable a `NULL`. |
| **Body limitado a 1 MB** | `http.MaxBytesReader` protege contra payloads gigantes. |
| **Graceful shutdown** | Al recibir `SIGINT`/`SIGTERM`, el servidor espera 5 s a que los requests en vuelo terminen. El orden `srv.Shutdown` → `rateLimiter.Stop()` → `asyncLog.Stop()` → `rotator.Close()` es invariante. |
| **Interfaces por handler** | Rompe la dependencia directa al repositorio concreto y facilita el testing con mocks manuales. |

---

## 10. Concurrencia

El proyecto usa cinco patrones del stdlib de Go:

| Patrón | Dónde | Qué hace |
|---|---|---|
| **Bootstrap paralelo** | `cmd/api/main.go` | `WaitGroup` para iniciar el logger y la BD concurrentemente al arrancar |
| **Fan-out** | `internal/handlers/stats.go` | Dos queries a la BD corren en goroutines; canales con buffer=1 recolectan los resultados |
| **Worker pool** | `internal/repository/user.go:BatchCreate` | Semáforo `chan struct{}` de tamaño N limita las goroutines que escriben en la BD simultáneamente |
| **Async worker** | `internal/middleware/async_logger.go` | Canal buffereado + goroutine consumidora desacoplan el log del path de la request |
| **Background ticker** | `internal/middleware/rate_limiter.go` | Goroutine con `time.Ticker` limpia cada minuto los visitantes inactivos del map de limiters |
| **`sync.Once`** | `internal/password/policy.go` | Inicializa el mapa de contraseñas comunes exactamente una vez, de forma segura para goroutines concurrentes |

---

## 11. Testing

```bash
# Correr todos los tests
go test ./...

# Con salida detallada y cobertura por paquete
go test -v -cover ./...

# Detectar data races
go test -race ./...
```

Los tests son **unitarios puros** — no requieren base de datos ni servidor levantado.

| Paquete | Qué se prueba |
|---|---|
| `internal/password` | Todas las reglas de `Validate()`, incluyendo casos límite de longitud, cada tipo de complejidad, contraseñas comunes y coincidencia con email |
| `internal/handlers` | Todos los handlers HTTP (signup, login, bloqueo de cuenta, cambio de contraseña con historial, CRUD de usuarios y tareas, stats) |
| `internal/config` | Carga de variables de entorno, defaults y nuevas variables de política |
| `internal/middleware` | Auth JWT y rate limiter |

Los mocks son structs manuales que implementan la interfaz correspondiente — sin librerías de mocking:

```go
type mockPasswordHistoryRepo struct {
    createFn             func(ctx context.Context, h *models.PasswordHistory) error
    findRecentByUserIDFn func(ctx context.Context, userID uint, limit int) ([]models.PasswordHistory, error)
}
```

---

## 12. Herramientas de desarrollo

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
