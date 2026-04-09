You are building a production-ready backend service in **Go** for a modular, self-hosted platform called **Helios**.

## 🧠 Project Overview

Create a backend service named **Helios Core**.

Helios Core is the **central shared service** responsible for:

* Authentication
* User management
* JWT token issuance
* Acting as the identity provider for all future Helios services

The system must be:

* Lightweight and high-performance
* Minimal dependencies
* Clean and modular (not overengineered)
* Designed for self-hosting via Docker

---

## 🏗️ Architecture Principles

1. Helios Core is a standalone service.
2. All future services (todo, fitness, media, etc.) will:

   * Trust Helios Core for authentication
   * Validate JWT tokens issued by Helios Core
3. There is a **single global user system** shared across all services.
4. No other service will manage users.

---

## ⚙️ Tech Stack (STRICT)

* Language: Go
* HTTP framework: Fiber (preferred for performance) OR Gin (if simpler)
* Database: PostgreSQL
* Query approach: Prefer SQLC or lightweight queries (avoid heavy ORM)
* JWT: golang-jwt
* Google Auth: official Google ID token verification library

---

## 📦 Project Structure (Clean Architecture)

```id="go-struct-01"
helios-core/
  cmd/
    server/
      main.go
  internal/
    auth/
    user/
    config/
    handler/
    middleware/
  pkg/
    jwt/
    google/
    db/
  migrations/
  docker/
  docker-compose.yml
  go.mod
```

Follow separation of concerns:

* `internal/` → business logic
* `pkg/` → reusable utilities
* `handler/` → HTTP layer

---

## 🔐 Authentication Flow (MVP)

ONLY implement **Google Login**.

### Flow:

1. Frontend gets Google ID token
2. Sends it to backend:

### POST /auth/google

Request:

```json id="go-req-01"
{
  "idToken": "google_id_token_here"
}
```

---

### Backend Responsibilities:

1. Verify Google ID token:

   * Use Google public keys
   * Validate audience if possible

2. Extract:

   * email
   * name
   * picture
   * sub (Google user ID)

3. Check PostgreSQL:

   * If user exists → return user
   * Else → create new user

4. Generate JWT:

   * Include user_id
   * email
   * issued_at
   * expiry (configurable)

5. Return:

```json id="go-res-01"
{
  "user": {
    "id": "uuid",
    "email": "user@email.com",
    "name": "User Name",
    "avatar": "url"
  },
  "token": "jwt_token_here"
}
```

---

## 🗄️ Database Schema (PostgreSQL)

Table: users

Fields:

* id (UUID, primary key)
* email (unique, indexed)
* name
* avatar_url
* google_id (unique, indexed)
* created_at
* updated_at

---

## 🔐 Security Requirements

* Validate all inputs strictly
* Do NOT store Google tokens
* JWT secret via environment variable
* Configurable token expiration
* Use HTTPS-ready setup
* Prevent duplicate users

---

## ⚡ Performance Guidelines

* Keep allocations minimal
* Avoid unnecessary abstractions
* Use connection pooling for DB
* Prefer simple SQL over ORM
* Keep middleware lightweight

---

## 🐳 Docker Requirements

Provide:

* Multi-stage Dockerfile (small final image)
* docker-compose with:

  * helios-core service
  * postgres service

---

## ⚙️ Config

Use environment variables:

Example:

```env id="go-env-01"
PORT=8080
DATABASE_URL=postgres://user:pass@db:5432/helios
JWT_SECRET=supersecret
JWT_EXPIRY=3600
GOOGLE_CLIENT_ID=your_client_id
```

---

## 🔄 Future Extensibility (IMPORTANT)

Design code so we can later:

* Add Apple/email login
* Add refresh tokens
* Add roles/permissions
* Add service-to-service auth

---

## 🎯 Output Requirements

Generate:

1. Full working Go project
2. Clean folder structure
3. Working `/auth/google` endpoint
4. DB connection + queries
5. JWT implementation
6. Docker setup
7. Migration file
8. Instructions to run locally

---

## 🚫 Do NOT include

* No microservices yet
* No event bus
* No Redis
* No overengineering

---

Focus on:
✔ Clean code
✔ Performance
✔ Simplicity
✔ Production-ready foundation
