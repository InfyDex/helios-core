# Helios Core

Central identity service for the Helios platform: Google Sign-In, PostgreSQL users, and JWT issuance for other services.

**Machine / agent context:** see [`AGENTS.md`](./AGENTS.md) for a concise project map, API list, env vars, and where to change code—useful for new chats or Cursor rules.

**Flutter host app (separate repo):** use [`HELIOS_FLUTTER_HOST_PROMPT.md`](./HELIOS_FLUTTER_HOST_PROMPT.md) as a copy-paste prompt to scaffold the Helios modular client (Google login, plugins, `AGENTS.md` + Cursor rule for Flutter).

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
| `GOOGLE_CLIENT_ID` | One or more **OAuth 2.0 Client IDs** (comma-separated, no spaces required). Each platform’s ID token has an `aud` claim that must match **one** of these values. |

Helios verifies the Google ID token signature and expiry, then checks that `aud` is in this list. Use the **Web**, **Android**, and **iOS** client IDs from the same Google Cloud project (see below).

### Google Cloud: create Web, Android, and iOS OAuth clients

Do this once per project in [Google Cloud Console](https://console.cloud.google.com/).

1. **Pick or create a project** (same project for all three clients).

2. **OAuth consent screen** — **APIs & Services** → **OAuth consent screen**. Choose *External* or *Internal* (Workspace), fill app name, support email, and developer contact. Add scopes your apps need (at minimum the defaults for sign-in: `openid`, `email`, `profile`; add `.../auth/user.phonenumber.read` only if you use phone). Save.

3. **Web client** — **APIs & Services** → **Credentials** → **Create credentials** → **OAuth client ID** → Application type **Web application**.  
   - **Authorized JavaScript origins**: your real web origins (e.g. `https://app.example.com`, `http://localhost:3000` for dev).  
   - **Authorized redirect URIs**: only if your web flow uses redirects (e.g. `http://localhost:3000` callback).  
   Create → copy the **Client ID** (ends with `.apps.googleusercontent.com`).

4. **Android client** — **Create credentials** → **OAuth client ID** → type **Android**.  
   - **Package name**: same as `applicationId` in `build.gradle` (e.g. `com.example.helios`).  
   - **SHA-1 certificate fingerprint**: use your **debug** keystore for local builds and add **release** SHA-1 for production (you can create multiple Android clients or add multiple fingerprints to one client, depending on Console UI).  
   Debug SHA-1 (typical):

   ```powershell
   keytool -list -v -keystore "$env:USERPROFILE\.android\debug.keystore" -alias androiddebugkey -storepass android -keypass android
   ```

   Copy **Client ID**.

5. **iOS client** — **Create credentials** → **OAuth client ID** → type **iOS**.  
   - **Bundle ID**: same as Xcode (e.g. `com.example.helios`).  
   Create → copy **Client ID**.

6. **Configure Helios** — set all three IDs (order does not matter):

   ```env
   GOOGLE_CLIENT_ID=WEB_ID.apps.googleusercontent.com,ANDROID_ID.apps.googleusercontent.com,IOS_ID.apps.googleusercontent.com
   ```

7. **Mobile apps** — use each platform’s Google Sign-In SDK and ensure the returned **ID token** is sent to `POST /core/v1/auth/google`. The token’s `aud` will be that platform’s OAuth client ID, which must appear in `GOOGLE_CLIENT_ID` above.

**Optional shortcut:** some teams use only the **Web** client ID in `requestIdToken(webClientId)` on Android/iOS so every token has the same `aud`. Then Helios only needs that single web ID. If you prefer **native** Android/iOS clients (separate `aud` per platform), use the comma-separated list as in step 6.

## Run with Docker Compose

From the repository root, set `GOOGLE_CLIENT_ID` (one ID or comma-separated web/Android/iOS IDs), then build and start:

```powershell
$env:GOOGLE_CLIENT_ID = "web-id.apps.googleusercontent.com,android-id.apps.googleusercontent.com,ios-id.apps.googleusercontent.com"
docker compose up --build
```

Alternatively, put `GOOGLE_CLIENT_ID=...` in a `.env` file beside `docker-compose.yml`; Compose substitutes it into the service environment. Use **straight double quotes** in `.env` if your tooling requires wrapping values that contain commas.

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
- `POST /core/v1/auth/google` — body `{"idToken":"<Google credential>"}`; returns user JSON (including `phone` when known) and a Helios JWT

To receive a phone number in the ID token, the frontend must request the scope `https://www.googleapis.com/auth/user.phonenumber.read` (and the user must approve). The backend reads the OIDC `phone_number` claim; without that scope, `phone` is stored and returned as empty until a later login includes it.

If you already had a database from an older `001_init.sql` without `phone`, apply `migrations/002_user_phone.sql` once (e.g. `psql "$DATABASE_URL" -f migrations/002_user_phone.sql` or run the `ALTER` inside your DB). New Docker volumes run all `migrations/*.sql` in order on first Postgres start.

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
