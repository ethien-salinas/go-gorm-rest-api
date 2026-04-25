# go-gorm-rest-api

REST API construida con Go como proyecto de aprendizaje. Implementa un CRUD de usuarios y tareas usando Gorilla Mux como router, GORM como ORM y PostgreSQL como base de datos.

## Stack

- **Go** — lenguaje principal
- **[Gorilla Mux](https://github.com/gorilla/mux)** — router HTTP
- **[GORM](https://gorm.io/)** — ORM con driver de PostgreSQL
- **PostgreSQL** — base de datos relacional
- **Docker Compose** — levanta la base de datos y Adminer localmente
- **[Air](https://github.com/air-verse/air)** — hot reload en desarrollo

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

| Variable      | Descripción                  | Default   |
|---------------|------------------------------|-----------|
| `DB_HOST`     | Host de la base de datos     | localhost |
| `DB_USER`     | Usuario de PostgreSQL        | postgres  |
| `DB_PASSWORD` | Contraseña de PostgreSQL     | —         |
| `DB_NAME`     | Nombre de la base de datos   | mydb      |
| `DB_PORT`     | Puerto de PostgreSQL         | 5432      |
| `DB_SSLMODE`  | Modo SSL de la conexión      | disable   |
| `PORT`        | Puerto del servidor HTTP     | 3000      |

## Endpoints

### Usuarios

| Método   | Ruta            | Descripción               |
|----------|-----------------|---------------------------|
| `GET`    | `/users`        | Listar todos los usuarios |
| `GET`    | `/users/{id}`   | Obtener un usuario        |
| `POST`   | `/users`        | Crear un usuario          |
| `PUT`    | `/users/{id}`   | Actualizar un usuario     |
| `DELETE` | `/users/{id}`   | Eliminar un usuario       |

### Tareas

| Método   | Ruta            | Descripción              |
|----------|-----------------|--------------------------|
| `GET`    | `/tasks`        | Listar todas las tareas  |
| `GET`    | `/tasks/{id}`   | Obtener una tarea        |
| `POST`   | `/tasks`        | Crear una tarea          |
| `PUT`    | `/tasks/{id}`   | Actualizar una tarea     |
| `DELETE` | `/tasks/{id}`   | Eliminar una tarea       |

### Ejemplos

```bash
# Crear un usuario
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"first_name":"tommy","last_name":"lee","email":"tommy.lee@example.com"}'

# Crear una tarea asociada al usuario con id 1
curl -X POST http://localhost:3000/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"learn gorm","description":"build web apps with gorm","user_id":1}'

# Listar usuarios con sus tareas
curl http://localhost:3000/users
```

> Los archivos `.http` en `request/` pueden usarse directamente desde VS Code con la extensión [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).

## Arquitectura

```
cmd/api/main.go          → entry point: wiring de dependencias y registro de rutas
internal/config/         → carga de variables de entorno
internal/database/       → conexión a PostgreSQL via GORM
internal/models/         → definición de entidades (User, Task)
internal/repository/     → consultas a la base de datos por entidad
internal/handlers/       → handlers HTTP organizados por entidad
```

Cada capa recibe sus dependencias por inyección — no hay estado global.
