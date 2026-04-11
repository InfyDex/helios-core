# Helios Core

Central identity service for the Helios platform: Google Sign-In, PostgreSQL users, and JWT issuance for other services.

## Prerequisites

- Go 1.22+
- PostgreSQL 16+ (or Docker)
- [sqlc](https://docs.sqlc.dev/en/stable/overview/install.html) — only if you change SQL under `sql/queries/` or `migrations/`

## Configuration

Copy `.env.example` to `.env` and set:

| Variable | Description |
|----------|-------------|
| `PORT` | HTTP listen port (default `8080`) |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | HMAC secret; **minimum 32 characters** |
| `JWT_EXPIRY` | Access token lifetime in seconds (minimum `60`) |
| `GOOGLE_CLIENT_ID` | OAuth 2.0 Web client ID used to validate ID tokens |

`GOOGLE_CLIENT_ID` must match the client ID your frontend uses with Google Identity Services so the token `aud` claim validates.

## Run with Docker Compose

From the repository root, set your Web client ID, then build and start:

```powershell
$env:GOOGLE_CLIENT_ID = "your-id.apps.googleusercontent.com"
docker compose up --build
```

Alternatively, put `GOOGLE_CLIENT_ID=...` in a `.env` file beside `docker-compose.yml`; Compose substitutes it into the service environment.

On first start, Postgres runs `migrations/001_init.sql` automatically. The API listens on [http://localhost:8080](http://localhost:8080).

## Run locally (without Docker for the API)

1. Start PostgreSQL and apply the schema:

   ```bash
   psql "$DATABASE_URL" -f migrations/001_init.sql
   ```

2. Export the same variables as in `.env.example` (or use a tool that loads `.env`).

3. Build and run:

   ```bash
   go run ./cmd/server
   ```

## API

All routes are under **`/core/v1`**.

- `GET /core/v1/health` — liveness
- `POST /core/v1/auth/google` — body `{"idToken":"<Google credential>"}`; returns user JSON and a Helios JWT

## Regenerating database code (after SQL changes)

```bash
sqlc generate
```

If `sqlc` is not installed:

```bash
go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
```

## Project layout

- `cmd/server` — process entrypoint
- `internal/` — auth, config, handlers, middleware, user store
- `pkg/` — JWT helper, Google verification, sqlc-generated `db` package
- `migrations/` — PostgreSQL schema
- `docker/` — production-oriented multi-stage Dockerfile
