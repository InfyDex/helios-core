# Helios Core — agent / AI context

Use this file (and `README.md` for runbooks) when onboarding a new chat or automation to this repository.

## What this is

**Helios Core** is a self-hosted **Go** backend: central **identity** for the Helios platform. It verifies **Google Sign-In ID tokens**, stores users in **PostgreSQL**, and issues **Helios JWTs** for other services. It does **not** implement Admin SDK, microservices, Redis, or event buses.

**Companion client:** the **Helios Flutter host app** (separate repository) is the modular shell: it performs Google Sign-In, exchanges the ID token with this API (`POST /core/v1/auth/google`), stores the Helios JWT, and exposes auth to **feature plugins** (todo, movies, etc.) via a small contract package—plugins do not re-implement Google login. See **`HELIOS_FLUTTER_HOST_PROMPT.md`** for a scaffold prompt.

**Feature backends (e.g. todo):** separate services/repos. Copy-paste prompts: **`HELIOS_TODO_SERVICE_PROMPT.md`** (Go API), **`HELIOS_FLUTTER_TODO_PLUGIN_PROMPT.md`** (Flutter plugin wiring).

## Inter-service communication (Core ↔ todo ↔ client)

1. **Login (once):** the client talks only to **Helios Core** → `POST /core/v1/auth/google` → receives **Helios JWT** (HS256; claims include `user_id`, `email`, `sub` = user id; standard `exp` / `iat`).
2. **Authenticated API calls:** the client sends **`Authorization: Bearer <Helios JWT>`** to **each feature service** (e.g. todo). The todo service **does not** call Core on every request for MVP; it **verifies the JWT locally** using the **same `JWT_SECRET`** as Helios Core and reads **`user_id` / `sub`** to scope data. Google ID tokens are **not** sent to feature services for normal CRUD.
3. **Trust boundary:** any service that holds `JWT_SECRET` can mint or validate user tokens—treat all such services as **highly trusted**; rotate the secret **together** across Core + consumers. **Future options** (not in Core today): asymmetric signing (RS256) + JWKS from Core, short-lived tokens + refresh, or **token introspection** (`POST /core/v1/introspect`) if you want a network check per request.
4. **Service-to-service (backend only):** if one microservice calls another **without** an end-user JWT, use **mTLS**, **internal API keys**, or **workload identity**—separate from the user Bearer flow above.
5. **Browser / Flutter web:** if the todo API is on another origin than the web app, configure **CORS** on the todo service for allowed origins and headers (`Authorization`, `Content-Type`).

- **Module:** `github.com/infydex/helios-core`
- **HTTP:** [Fiber](https://github.com/gofiber/fiber) v2
- **DB:** PostgreSQL via [pgx](https://github.com/jackc/pgx) pool + **[sqlc](https://sqlc.dev/)** generated queries in `pkg/db`
- **Google:** `google.golang.org/api/idtoken` — signature/expiry/issuer; **audience** must match one of the configured OAuth client IDs
- **JWT:** `github.com/golang-jwt/jwt/v5`, HS256, secret + expiry from env

## HTTP surface

All public routes live under **`/core/v1`** (see `internal/handler/routes.go` → `CoreAPIPrefix`).

| Method | Path | Role |
|--------|------|------|
| `GET` | `/core/v1/health` | Liveness |
| `POST` | `/core/v1/auth/google` | Body: `{"idToken":"..."}`. Returns `{ "user": { id, email, name, avatar, phone }, "token": "<Helios JWT>" }` |

Handlers: `internal/handler/`. Middleware: `internal/middleware/` (e.g. request ID).

## Auth flow (MVP)

1. Client sends Google **ID token** (JWT from Google Sign-In / GIS), not the Admin SDK service account.
2. `pkg/google.VerifyIDToken` validates the token with **empty** audience passed to `idtoken.Validate` (signature/expiry), then checks **`aud`** is in the allowlist, **`iss`** is a Google issuer, and reads claims: `sub`, `email`, `name`, `picture`, `phone_number` (phone only if the client requested phone scope and Google included it).
3. `internal/user.Store.GetOrCreateByGoogle` loads or creates `users` by **`google_id` = `sub`**. Phone can be set on create or updated on later logins when the token carries `phone_number`.
4. `pkg/jwt.Sign` issues the app JWT (`user_id`, `email`, standard `iat`/`exp`, `sub`).

Invalid Google tokens map to **401**; duplicate DB edge cases to **409**; other failures **500**. See `internal/handler/auth.go` and `internal/auth/service.go`.

## Configuration (env)

| Variable | Notes |
|----------|--------|
| `PORT` | Default `8080` |
| `DATABASE_URL` | Required PostgreSQL URL |
| `JWT_SECRET` | Required, **≥ 32** characters |
| `JWT_EXPIRY` | Seconds, minimum **60** |
| `GOOGLE_CLIENT_ID` | Required. **One or more** OAuth 2.0 client IDs, **comma-separated** (trimmed). ID token `aud` must equal **one** of them — supports Web + Android + iOS clients. Parsed in `internal/config/config.go` → `GoogleClientIDs`. |

Example `.env` shape: see `.env.example`. Docker: `docker-compose.yml` + `README.md`.

## Data model

- **Table `users`:** `id` (UUID), `email` (unique), `name`, `avatar_url`, `phone` (nullable), `google_id` (unique), `created_at`, `updated_at`.
- **Migrations:** `migrations/*.sql` — Postgres Docker init runs them on **first** DB create only. Older DBs may need manual `002_user_phone.sql` if `phone` was added later (see `README.md`).

Schema changes: edit SQL → update `sql/queries/*.sql` → run **`sqlc generate`** (or `go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate`) → fix any compile errors in `internal/` / `pkg/`.

## Directory map

```
cmd/server/main.go       # entry: config, DB pool, Fiber, routes
internal/auth/           # Google login orchestration + JWT
internal/config/         # env loading
internal/handler/        # HTTP (auth, health); CoreAPIPrefix
internal/middleware/
internal/user/           # persistence helpers around sqlc
pkg/db/                  # generated — do not hand-edit
pkg/google/              # ID token verification + claim mapping
pkg/jwt/                 # Helios JWT signing
migrations/              # Postgres DDL
sql/queries/             # sqlc query definitions
docker/Dockerfile
docker-compose.yml
req.md                   # original product spec (may drift slightly)
README.md                # human-oriented setup, Google Console, API notes
```

## Conventions for changes

- Prefer **small, focused** edits; match existing style and imports.
- **Do not edit** `pkg/db/*.go` by hand — regenerate with **sqlc**.
- New authenticated routes: reuse `handler.CoreAPIPrefix` or a nested `Group`; keep versioning under `/core/v1` unless the product decision is to version elsewhere.
- **Future-friendly** (not implemented yet): Apple/email login, refresh tokens, roles, service-to-service auth — keep new code modular (e.g. auth provider interface) rather than hard-coding only Google where a second provider will plug in.

## Tests and static analysis

```bash
go test ./...
go vet ./...
# Lint (config: .golangci.yml; generated sqlc code under pkg/db is excluded)
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./...
```

Add or extend **`*_test.go`** when behavior changes (`internal/config`, `pkg/jwt`, `pkg/google`, `internal/handler`, `internal/middleware`, etc.). After edits, **run `go test ./...` and golangci-lint** before considering work done.

Quick compile: `go build -o bin/helios-core ./cmd/server`

Full Docker / Google Cloud steps: **`README.md`**.

## Cursor

Workspace rule: **`.cursor/rules/helios-core.mdc`** (`alwaysApply: true`) — points agents at this file and the invariants above.
