# go-gorm-rest-api

> **Proyecto de aprendizaje** — REST API construida en Go para aprender el lenguaje viniendo de un background de JavaScript y Java. Implementa un CRUD completo de usuarios y tareas con autenticación JWT.

Si llegas de JavaScript (Node.js/Express) y Java (Spring Boot), este README intenta explicar no solo *qué* hace el proyecto, sino también *por qué* Go funciona diferente y qué conceptos vale la pena tener claros. Las comparaciones en cada sección usan ambos lenguajes como referencia.

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
9. [Concurrencia y control de flujo en Go](#9-concurrencia-y-control-de-flujo-en-go)
10. [Testing](#10-testing)
11. [Conceptos de Go — comparación con JavaScript y Java](#11-conceptos-de-go--comparación-con-javascript-y-java)
12. [Herramientas de desarrollo](#12-herramientas-de-desarrollo)

---

## 1. Stack tecnológico

| Herramienta | Rol en el proyecto | Equivalente en JS | Equivalente en Java |
|---|---|---|---|
| **Go 1.21+** | Lenguaje principal | Node.js | JDK / JVM |
| **[Gorilla Mux](https://github.com/gorilla/mux)** | Router HTTP | Express / Fastify | Spring MVC / Javalin |
| **[GORM](https://gorm.io/)** | ORM para PostgreSQL | Prisma / Sequelize | Hibernate / Spring Data JPA |
| **PostgreSQL** | Base de datos relacional | — | — |
| **Docker Compose** | Levanta la BD y Adminer localmente | — | — |
| **[Air](https://github.com/air-verse/air)** | Hot reload en desarrollo | `nodemon` | Spring Boot DevTools |
| **`log/slog`** | Logging estructurado en JSON (stdlib Go 1.21) | `pino` / `winston` | SLF4J + Logback |
| **[golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)** | Generación y validación de JWT (HS256) | `jsonwebtoken` | `java-jwt` / Spring Security |
| **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** | Hashing de contraseñas | `bcrypt` (npm) | `BCryptPasswordEncoder` (Spring) |
| **[testify](https://github.com/stretchr/testify)** | Aserciones en tests unitarios | Jest / Vitest | JUnit 5 + AssertJ |

---

## 2. Estructura del proyecto

```
go-gorm-rest-api/
│
├── cmd/
│   └── api/
│       └── main.go          # Entry point: wiring de dependencias y arranque del servidor
│
├── internal/                # Código privado — no puede importarse desde fuera del módulo
│   ├── config/              # Carga de variables de entorno → struct Config
│   ├── database/            # Conexión a PostgreSQL via GORM
│   ├── models/              # Entidades GORM: User, Task
│   ├── repository/          # Queries a la BD, encapsuladas por entidad
│   ├── handlers/            # Handlers HTTP (equivalente a los controllers en Express)
│   ├── middleware/          # Auth JWT + Logging asíncrono
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

> **¿Por qué `internal/`?**  
> En Go, el directorio `internal/` tiene una regla especial del compilador: solo puede ser importado por código que esté dentro del mismo módulo. Es la forma de marcar que esas partes no son una API pública.  
> — En **Node.js** no existe este concepto de forma nativa (cualquier archivo puede hacer `require` de otro).  
> — En **Java** el equivalente más cercano es el modificador de acceso `package-private` (sin `public`) o usar módulos de Java 9+ con `module-info.java` para declarar qué paquetes se exportan.

---

## 3. Requisitos previos

- [Go 1.21+](https://go.dev/dl/) — verifica con `go version`
- [Docker Desktop](https://www.docker.com/) — para levantar PostgreSQL y Adminer
- [Air](https://github.com/air-verse/air) _(opcional)_ — hot reload

```bash
# Instalar Air globalmente
go install github.com/air-verse/air@latest
```

> **`go install` vs `npm install -g` vs Maven/Gradle**  
> `go install` descarga, compila e instala un binario en `$GOPATH/bin` (o `$GOBIN`).  
> — En **JavaScript**: equivale a `npm install -g` para herramientas CLI.  
> — En **Java**: más cercano a instalar una herramienta vía `sdk install` (SDKMAN) o descargar un JAR ejecutable. No existe un equivalente directo en Maven/Gradle porque esos gestionan dependencias de proyecto, no binarios globales.

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

> **Gestión de dependencias — comparación**
>
> | Concepto | Go | JavaScript | Java |
> |---|---|---|---|
> | Declaración de deps | `go.mod` | `package.json` | `pom.xml` / `build.gradle` |
> | Lock file (hashes) | `go.sum` | `package-lock.json` | `pom.xml` (Maven lock) / `gradle.lockfile` |
> | Instalar/sincronizar | `go mod tidy` | `npm install` | `mvn install` / `gradle build` |
> | Vendor local | `go mod vendor` | `node_modules/` | `.m2/` (caché local Maven) |
>
> `go mod tidy` descarga lo que falta y elimina lo que ya no se usa — en un solo comando.

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

> **`AUTO_MIGRATE`**: Cuando está en `true`, GORM crea/actualiza las tablas automáticamente al arrancar el servidor. Útil al principio para no tener que escribir migraciones. En producción se recomienda dejarlo en `false` y gestionar migraciones de forma explícita.

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

> Los archivos `.http` en `request/` pueden usarse directamente desde VS Code con la extensión [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client). Para declarar variables en esos archivos, usa la sintaxis `@token = valor` (no JSON).

---

## 7. Autenticación JWT

El flujo de autenticación es el mismo que en cualquier API con JWT:

```
POST /auth/signup  →  hashea password con bcrypt  →  crea User en BD
POST /auth/login   →  verifica hash bcrypt         →  emite JWT { sub: userID, exp: now+Nh }
GET  /api/v1/*     →  middleware valida Bearer      →  inyecta userID en el contexto del request
```

El token debe enviarse en el header `Authorization`:

```
Authorization: Bearer <token>
```

> **¿Cómo se inyecta el `userID` en el contexto?**  
> En Express usarías `req.user = decoded`. En Go, el middleware escribe el valor en el `context.Context` del request con una clave propia, y el handler lo lee con `r.Context().Value(...)`. Es más verboso pero más explícito y seguro.

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

Ninguna capa conoce a la superior. Los handlers no importan `repository` directamente — solo conocen una **interfaz**:

```go
// El handler define QUÉ necesita, no CÓMO se implementa
type UserRepository interface {
    FindAll() ([]models.User, error)
    FindByID(id uint) (*models.User, error)
    // ...
}
```

Esto hace que los tests sean muy sencillos: basta con crear un struct que implemente la interfaz con comportamientos controlados (mock manual, sin librerías).

> **Interfaces implícitas:** en Go no hay `implements` — cualquier tipo que tenga los métodos de la interfaz la satisface automáticamente. Ver [sección 11 → Interfaces implícitas](#interfaces-implícitas-structural-typing) para la comparación detallada con Java y TypeScript.

### Decisiones técnicas relevantes

| Decisión | Razonamiento |
|---|---|
| **PATCH usa punteros** (`*string`, `*bool`) | Permite distinguir "campo no enviado" (`nil`) de "campo enviado vacío" (`""`). Evita borrar datos accidentalmente. |
| **`Update` recibe `map[string]any`** | GORM genera un `UPDATE` selectivo que solo toca las columnas enviadas. `password_hash` nunca se sobreescribe por error. |
| **Body limitado a 1 MB** | `http.MaxBytesReader` protege contra payloads gigantes. En Express: `express.json({ limit: '1mb' })`. En Spring Boot: `spring.servlet.multipart.max-request-size=1MB`. |
| **Graceful shutdown** | Al recibir `SIGINT`/`SIGTERM`, el servidor espera 5 s a que los requests en vuelo terminen antes de cerrar. Luego hace flush del logger y cierra el archivo de logs. |
| **Interfaces por handler** | Cada handler define su propia interfaz de repositorio en lugar de importar el repositorio concreto. Esto rompe la dependencia circular y facilita el testing. |

---

## 9. Concurrencia y control de flujo en Go

Go tiene concurrencia incorporada en el lenguaje. Estos son los patrones que usa este proyecto:

### Goroutines

```go
go func() {
    // se ejecuta en paralelo, de forma concurrente
}()
```

Una goroutine es una función que corre concurrentemente. Son extremadamente baratas (unos pocos KB de stack inicial), a diferencia de los threads del SO.

| | Go | JavaScript | Java |
|---|---|---|---|
| Unidad de concurrencia | Goroutine | Event loop (single-thread); `worker_threads` para paralelismo real (poco usado) | Thread (`java.lang.Thread`) / Virtual Thread (Java 21+) |
| Coste de crear una | ~2–8 KB de stack, muy barato | N/A | ~1 MB de stack por thread de SO (caro); Virtual Threads: barato |
| Cuántas puedes tener | Miles/millones | — | Miles con Virtual Threads; cientos con threads de SO |
| Sincronización | Canales (`chan`), `sync.Mutex` | `Promise`, `async/await` | `synchronized`, `ReentrantLock`, `CompletableFuture` |

En Node.js todo corre en un solo hilo y la asincronía es cooperativa (event loop). En Java los threads de SO son costosos, aunque Java 21 introdujo Virtual Threads que se acercan al modelo de Go. En Go, el scheduler del runtime multiplexa miles de goroutines sobre los cores de CPU disponibles.

### `sync.WaitGroup` — esperar a que varias goroutines terminen

```go
var wg sync.WaitGroup
wg.Add(2)           // voy a esperar 2 goroutines

go func() {
    defer wg.Done() // avisa cuando termina
    // ... iniciar logger
}()

go func() {
    defer wg.Done()
    // ... conectar a la BD
}()

wg.Wait() // bloquea hasta que ambas llamen a Done()
```

**En este proyecto:** `main.go` usa este patrón para iniciar el logger y la BD en paralelo, ya que son operaciones de I/O independientes.

### Canales (`chan`) — comunicación entre goroutines

```go
ch := make(chan int, 1) // canal buffereado de tamaño 1

go func() {
    ch <- 42 // enviar
}()

valor := <-ch // recibir (bloquea hasta que haya algo)
```

Los canales son el mecanismo principal de comunicación entre goroutines en Go. El motto es: *"Do not communicate by sharing memory; share memory by communicating."*

**En este proyecto:**
- **Fan-out en `stats.go`**: dos queries a la BD corren en paralelo, cada una escribe en su canal. El handler lee de ambos canales. Buffer de tamaño 1 para evitar goroutine leaks si el handler abandona antes de leer.
- **Semáforo en `BatchCreate`**: un `chan struct{}` de tamaño N actúa como semáforo para limitar cuántas goroutines escriben en la BD simultáneamente.
- **AsyncLogger**: los requests no bloquean esperando que el log se escriba al disco; envían el log a un canal buffereado y un worker goroutine lo procesa en segundo plano.

### `defer` — ejecutar algo al salir de la función

```go
func miFuncion() {
    defer fmt.Println("esto se ejecuta al final")
    fmt.Println("esto se ejecuta primero")
}
```

`defer` es muy común en Go para liberar recursos (cerrar archivos, hacer unlock de un mutex, etc.). Se ejecuta cuando la función retorna, sin importar si fue por `return` normal o por `panic`.

---

## 10. Testing

```bash
# Correr todos los tests
go test ./...

# Con salida detallada y cobertura por paquete
go test -v -cover ./...

# Detectar data races (condiciones de carrera)
go test -race ./...
```

Los tests son **unitarios puros** — no requieren base de datos ni servidor levantado.

| Paquete | Cobertura |
|---|---|
| `internal/handlers` | 100 % |
| `internal/config` | 100 % |
| `internal/middleware` | 100 % |

**¿Cómo se hacen mocks en Go?**

Cada handler define una interfaz de repositorio. Para testear, se crea un struct que la implemente con el comportamiento que quieres:

```go
// Go — mock manual, sin librerías
type mockUserRepo struct {
    users []models.User
    err   error
}

func (m *mockUserRepo) FindAll() ([]models.User, error) {
    return m.users, m.err
}
```

```js
// JavaScript (Jest) — equivalente
const mockRepo = { findAll: jest.fn().mockResolvedValue(users) }
```

```java
// Java (Mockito) — equivalente
@Mock UserRepository mockRepo;
when(mockRepo.findAll()).thenReturn(List.of(user));
```

En Go el mock es un struct plano — sin anotaciones, sin framework. Más verboso que Jest o Mockito, pero 100 % explícito y sin magia en tiempo de ejecución.

> **`go test -race`**: Go tiene un detector de data races incorporado en el toolchain. Al correr tests con `-race`, el runtime instrumenta el código y avisa si dos goroutines acceden a la misma memoria sin sincronización. Muy útil cuando empiezas a escribir código concurrente.

---

## 11. Conceptos de Go — comparación con JavaScript y Java

### Manejo de errores — no hay `try/catch`

En Go los errores son **valores** que se retornan explícitamente, no excepciones:

```go
// Go — el error es un segundo valor de retorno
result, err := alguienPuedeFallar()
if err != nil {
    return fmt.Errorf("contexto: %w", err) // %w conserva el error original (unwrap)
}
```

```js
// JavaScript — excepciones con try/catch
try {
    const result = await alguienPuedeFallar()
} catch (err) {
    throw new Error(`contexto: ${err.message}`)
}
```

```java
// Java — excepciones checked y unchecked
try {
    Result result = alguienPuedeFallar(); // puede lanzar IOException (checked)
} catch (IOException e) {
    throw new RuntimeException("contexto: " + e.getMessage(), e);
}
```

**Diferencias clave:**
- En **Java** las excepciones `checked` (como `IOException`) obligan a declararlas en la firma (`throws`) o capturarlas. En Go el compilador no te obliga a manejar el error, pero ignorarlo es explícito: `_ , _ = f()` — es obvio que estás descartando algo.
- En **JavaScript** cualquier función puede lanzar cualquier cosa en cualquier momento — el compilador no te avisa. En Go la firma del tipo te dice exactamente qué puede fallar.
- **`%w` en Go** es como pasar `cause` al constructor de una excepción en Java: `new RuntimeException(msg, cause)`. Permite inspeccionar la cadena de errores con `errors.Is` / `errors.As`.

### Punteros — breve introducción

```go
// Go
x := 42
p := &x          // & obtiene la dirección de memoria
fmt.Println(*p)  // * desreferencia (accede al valor)
```

```java
// Java — los objetos siempre son referencias (punteros implícitos)
// Los primitivos (int, bool) se pasan por valor
Integer x = 42;  // autoboxing; x es una referencia al objeto
```

```js
// JavaScript — los objetos se pasan por referencia, los primitivos por valor
const obj = { x: 42 }   // referencia
const n = 42            // valor
```

**Diferencias clave:**
- En **Java** los objetos *siempre* son referencias (punteros implícitos) y los primitivos *siempre* son valores. No tienes control explícito.
- En **JavaScript** igual: no manipulas direcciones de memoria manualmente.
- En **Go** tienes control explícito. Pasas `&valor` cuando quieres que la función modifique el original o cuando el struct es grande y copiar sería costoso. Usas `*Tipo` en un campo de struct cuando ese campo puede ser `nil` (ausente) — lo que en Java sería un `Optional<T>` o un campo nullable.

**Usos en este proyecto:**
- `db.Create(&user)` — GORM necesita la dirección para poder modificar el struct (escribir el ID generado).
- `*string`, `*bool` en DTOs de `PATCH` — distingue "campo no enviado" (`nil`) de "campo enviado vacío" (`""`).

### Structs, JSON y anotaciones

```go
// Go — struct con struct tags (backticks)
type User struct {
    ID        uint   `json:"id"        gorm:"primaryKey"`
    FirstName string `json:"first_name"`
    Email     string `json:"email"`
}
```

```java
// Java — anotaciones con @
@Entity
@Table(name = "users")
public class User {
    @Id @GeneratedValue
    private Long id;

    @Column(name = "first_name")
    @JsonProperty("first_name")
    private String firstName;
}
```

```ts
// TypeScript (NestJS/TypeORM) — decoradores
@Entity()
export class User {
    @PrimaryGeneratedColumn()
    id: number

    @Column({ name: 'first_name' })
    firstName: string
}
```

**Diferencias clave:**
- Las **struct tags** de Go son strings literales en backticks inspeccionados vía `reflect` en tiempo de ejecución. Son más simples que las anotaciones de Java pero menos expresivas (no admiten lógica).
- Las **anotaciones** de Java (`@Entity`, `@Column`) son procesadas por el compilador o por el framework en runtime con AOP/proxies. En Go no hay AOP ni proxies transparentes — todo es explícito.
- En **TypeScript** los decoradores funcionan como las anotaciones de Java, pero dependen de `experimentalDecorators` y son stage 3 del estándar.

### `context.Context` — cancelación y valores entre capas

```go
// Go — context viaja explícitamente como primer parámetro
ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
r = r.WithContext(ctx)

userID := r.Context().Value(middleware.UserIDKey).(uint)
```

```java
// Java (Spring) — equivalente con SecurityContext (thread-local implícito)
SecurityContextHolder.getContext().getAuthentication().getPrincipal();
// O con @RequestAttribute / request.getAttribute("userId")
```

```js
// Express — el contexto se cuelga directamente del objeto request
req.user = decoded // en el middleware
req.user.id        // en el handler
```

**Diferencias clave:**
- En **Spring** el `SecurityContext` usa un `ThreadLocal` — el contexto viaja *implícitamente* en el hilo. Cómodo pero opaco: si cambias de hilo (async), el contexto se puede perder.
- En **Express** `req.user` es duck typing puro — cualquier propiedad que añadas al objeto request viaja sin tipos.
- En **Go** el `context.Context` es *explícito*: cada función que necesite el contexto debe recibirlo como primer parámetro. Más verboso, pero el flujo de datos es siempre visible. También propaga cancelación y timeouts automáticamente a toda la cadena de llamadas.

### Interfaces implícitas (structural typing)

```go
// Go — implementación implícita
type Animal interface {
    Sonido() string
}

type Perro struct{}

func (d Perro) Sonido() string { return "guau" }
// Perro satisface Animal automáticamente — el compilador lo comprueba en el punto de uso
```

```java
// Java — implementación explícita (nominal typing)
interface Animal {
    String sonido();
}

class Perro implements Animal {  // declaración obligatoria
    public String sonido() { return "guau"; }
}
```

```ts
// TypeScript — igual que Java, explícito
interface Animal {
    sonido(): string
}

class Perro implements Animal {
    sonido() { return 'guau' }
}
```

**Diferencias clave:**
- En **Java y TypeScript** la relación interfaz→clase debe declararse explícitamente. Si tienes una clase de una librería externa que coincide con tu interfaz, no puedes hacerla "implementarla" sin extenderla o crear un wrapper.
- En **Go** cualquier tipo que tenga los métodos correctos satisface la interfaz, aunque el tipo sea de un paquete externo. Esto da una flexibilidad enorme para adaptar código de terceros sin modificarlo.
- La contrapartida: en Go es más difícil ver *qué interfaces implementa* un struct de un vistazo. Los IDEs modernos (GoLand, VS Code + gopls) muestran esta información automáticamente.

### Métodos en structs vs clases

```go
// Go — no hay clases; los métodos se asocian a tipos con receiver
type UserService struct {
    repo UserRepository
}

func (s *UserService) GetAll() ([]User, error) {
    return s.repo.FindAll()
}
```

```java
// Java — métodos dentro de la clase
public class UserService {
    private final UserRepository repo;

    public UserService(UserRepository repo) { this.repo = repo; }

    public List<User> getAll() throws Exception {
        return repo.findAll();
    }
}
```

```js
// JavaScript/TypeScript
class UserService {
    constructor(private repo: UserRepository) {}

    async getAll(): Promise<User[]> {
        return this.repo.findAll()
    }
}
```

**Diferencias clave:**
- Go no tiene clases ni herencia. El `receiver` (`s *UserService`) es la forma en que Go asocia un método a un tipo. Es equivalente al `this` en Java/JS, pero declarado explícitamente en la firma.
- No existe `extends`. La composición se logra embebiendo structs: `type AdminService struct { UserService }` — el equivalente funcional (sin polimorfismo) de `extends`.
- No hay sobrecarga de métodos (overloading) en Go. Si necesitas variantes, usas nombres distintos o parámetros opcionales vía `...opts`.

### Compilación y binario

```bash
# Go — compila a un binario nativo estático, sin runtime externo
go build -o api ./cmd/api
./api  # se ejecuta solo, no necesita Go instalado
```

```bash
# Java — compila a bytecode, necesita JVM
javac Main.java && java Main
# Spring Boot: jar ejecutable
java -jar app.jar
```

```bash
# JavaScript — interpretado, necesita Node.js
node index.js
```

**Diferencias clave:**
- Go produce un **binario estático** por defecto: un solo archivo ejecutable sin dependencias externas. Ideal para Docker (imagen `FROM scratch`).
- Java requiere la JVM. Spring Boot empaqueta un "fat jar" con todas las dependencias, pero sigue necesitando `java` instalado.
- El tiempo de compilación de Go es notablemente más rápido que el de Java — uno de sus objetivos de diseño explícitos.

---

## 12. Herramientas de desarrollo

### Air — Hot Reload

```bash
air   # equivalente a nodemon
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

---

## Recursos para seguir aprendiendo Go

- [**A Tour of Go**](https://go.dev/tour/) — tutorial oficial interactivo, el mejor punto de partida
- [**Go by Example**](https://gobyexample.com/) — ejemplos concisos de cada concepto
- [**Effective Go**](https://go.dev/doc/effective_go) — guía oficial de idioms y buenas prácticas
- [**Go Concurrency Patterns**](https://go.dev/blog/pipelines) — patrones de canales y goroutines
- [**GORM Docs**](https://gorm.io/docs/) — documentación completa del ORM
- [**pkg.go.dev**](https://pkg.go.dev/) — documentación de cualquier paquete de Go (como npm.com pero para Go)
