# gp_payment_orchestration — Project Architecture

## Overview

```
cmd/            → Application entry points & CLI commands
internal/
  api/          → HTTP server, routes, middleware, request context
  common/       → Shared types, errors, constants, utilities, base models
  global/       → Singleton initializers for all infrastructure dependencies
  modules/      → Business domain logic (one package per feature)
  mq/           → RabbitMQ consumer and message handlers
  repo/         → Generic CRUD helpers for GORM (PostgreSQL) and MongoDB
  tasks/        → Asynq (Redis) background job server and handlers
```

---

## `cmd/` — Entry Points & CLI

| Path | Purpose |
|------|---------|
| `cmd/app/main.go` | **Application entry point** — loads `.env`, calls `global.New()` to bootstrap all dependencies, then runs the Cobra CLI. |
| `cmd/root.go` | Cobra root command definition — the binary is named `hr`. |
| `cmd/serve.go` | `serve` subcommand — starts the API server, Asynq task runner, and RabbitMQ consumer concurrently, then waits for a shutdown signal. |

---

## `internal/api/` — HTTP Layer

| Path | Purpose |
|------|---------|
| `server.go` | Initialises the Gin engine, attaches global middleware, registers routes, and starts listening on `LISTEN_ADDR`. |
| `routes/` | Route registration — one file per resource group (e.g. `users.go`, `payments.go`). |
| `middleware/` | Custom Gin middleware — auth, CORS, rate limiting, etc. |
| `context/` | Request-scoped helpers — extract user claims, request ID, etc. from `gin.Context`. |

**Convention:** Controller/handler functions live here. They call into `modules/` for business logic.

---

## `internal/common/` — Shared Primitives

| File | Purpose |
|------|---------|
| `structs.go` | Reusable DTOs: `Response`, `PageInfo`. |
| `models.go` | Base `Entity` struct with `ExtID`, timestamps, GORM hooks, and `Autofill()`. Embed this in all domain models. |
| `constants.go` | App-wide constants (e.g. `MAX_LOGGED_IN_ACCOUNTS`). |
| `errors.go` | Sentinel errors organised by category (4xx, 5xx, auth, payments). |
| `service.go` | Base `Service` struct wrapping `*global.App` — embed in domain services. |
| `services.go` | Utility functions: `GenerateID()`, `GenerateRandomString()`, `EncryptString()`, `DecryptString()`. |

---

## `internal/global/` — Infrastructure Singletons

All initialisers follow the `sync.Once` singleton pattern. `global.New()` creates the `App` struct once and returns it.

| File | Singleton | Backend |
|------|-----------|---------|
| `init.go` | `App` struct — central container for all dependencies | — |
| `db.go` | `GetDB()` — GORM client | PostgreSQL |
| `mongo.go` | `GetMongo()` — MongoDB client | MongoDB |
| `cache.go` | `GetCache()` — Redis client | Redis |
| `asynq.go` | `GetTaskQueue()` — Asynq client | Redis |
| `mq.go` | `GetMQ()` — RabbitMQ connection/channel, plus `Emit()` and `Consume()` helpers | RabbitMQ |
| `socket.go` | `GetSocketIO()` — Socket.IO server | Socket.IO |
| `validator.go` | `GetValidator()` — struct validator | go-playground/validator |

**`App` also provides:**
- `Register(models ...Model)` — auto-migrates GORM models.
- `GetPermissions(group)` — collects permissions across all registered models.
- `Close()` — graceful shutdown of all dependencies.

---

## `internal/modules/` — Business Logic

Each domain feature gets its own package here. Example layout for a `users` module:

```
modules/users/
  model.go      → User struct (embeds common.Entity, implements global.Model)
  service.go    → UserService (embeds common.Service)
  repo.go       → User-specific queries
```

**Convention:**
- Modules **do not** import `internal/api/`.
- Modules expose public methods that handlers/controllers call.

---

## `internal/repo/` — Generic CRUD

| File | Purpose |
|------|---------|
| `crud.go` | Type-parameterised GORM helpers: `CreateOne`, `GetOne`, `GetMany`, `Count`, `UpdateOne`, `InsertMany`, `DeleteOne`. Accepts any type implementing `Gormable` (`TableName() string`). |
| `mgodb.go` | Type-parameterised MongoDB helpers: `MongoCreateOne`, `MongoGetOne`, `MongoGetMany`, `MongoUpdateOne`, `MongoDeleteOne`, `MongoInsertMany`. Accepts any type implementing `Mongoable` (`CollectionName() string`). |

Use these to avoid repetitive DB code. For complex queries, add custom methods in the module's own package.

---

## `internal/mq/` — Message Queue Consumer

| Path | Purpose |
|------|---------|
| `receiver.go` | Central dispatcher — switches on `msg.RoutingKey` and delegates to the appropriate handler. |
| `handlers/` | One file per routing key group (e.g. `payment.go`, `notification.go`). |

Messages are consumed via `rmq.Consume()` in `internal/global/mq.go`. To emit a message, call `app.MQ.Emit(event, data)`.

---

## `internal/tasks/` — Background Jobs (Asynq)

| Path | Purpose |
|------|---------|
| `server.go` | Sets up the Asynq mux and starts the task server. |
| `handlers/` | One file per task type (e.g. `send_email.go`, `generate_report.go`). |

To enqueue a task from anywhere with access to `*global.App`, use `app.TaskQ.Enqueue()`.

---

## Summary of data flow

```
HTTP Request
  → cmd/serve.go (starts Gin server)
    → internal/api/routes/* (route matched)
      → internal/api/middleware/* (auth, etc.)
        → internal/modules/<feature>/service.go (business logic)
          → internal/repo/crud.go or mgodb.go (DB access)

Background task
  → internal/tasks/server.go (Asynq mux)
    → internal/tasks/handlers/* (task handler)
      → internal/modules/<feature>/service.go

MQ message
  → internal/mq/receiver.go (routing key dispatch)
    → internal/mq/handlers/* (message handler)
      → internal/modules/<feature>/service.go
```
