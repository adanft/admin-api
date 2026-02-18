# ADMIN API

REST API in Go for user administration with JWT authentication (access token) and rotating refresh tokens stored in an `HttpOnly` cookie.

This project is designed as a practical internal backend: quick startup with Docker Compose, `users` and `auth` HTTP endpoints, and PostgreSQL persistence using SQL migrations.

## Features

- User CRUD: `GET/POST/PUT/DELETE /users`
- Authentication flow: `register`, `login`, `me`, `refresh`, `logout`
- Access token via `Authorization: Bearer <token>` header
- Refresh token via `HttpOnly` cookie with rotation on `POST /auth/refresh`
- Password hashing with `bcrypt` (through `golang.org/x/crypto`)
- Consistent API errors with business `code` and HTTP `status`
- Middleware for CORS, recovery, request logging, and request ID

## Tech Stack

- Go `1.24.6`
- PostgreSQL `16`
- `net/http` (`ServeMux` with method + route patterns)
- Bun ORM (`github.com/uptrace/bun`)
- Docker Compose for local development

## Quick Start (Docker)

Requirements:

- Docker + Docker Compose

Run:

```bash
cp .env.example .env
docker compose up -d
```

API will be available at `http://localhost:9090`.

Note: SQL files in `migrations/` run when PostgreSQL initializes its data volume. To reset DB from scratch:

```bash
docker compose down -v
docker compose up -d
```

## Local Run (without Docker for API)

You need PostgreSQL running and environment variables configured (start from `.env.example`).

```bash
go run ./cmd
```

## Environment Variables

Key variables (full list in `.env.example`):

- `SERVER_ADDRESS` (recommended default: `:9090`)
- `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASS`, `DATABASE_NAME`, `DATABASE_SSL_MODE`
- `AUTH_JWT_SECRET`, `AUTH_JWT_ISSUER`, `AUTH_JWT_AUDIENCE`
- `AUTH_ACCESS_TOKEN_TTL` (example: `15m`)
- `AUTH_REFRESH_TOKEN_TTL` (example: `168h`)
- `AUTH_REFRESH_COOKIE_NAME`, `AUTH_REFRESH_COOKIE_PATH`, `AUTH_REFRESH_COOKIE_SECURE`, `AUTH_REFRESH_COOKIE_SAMESITE`

## Endpoints

### Auth

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /auth/me`

### Users

- `GET /users`
- `GET /users/{id}`
- `POST /users`
- `PUT /users/{id}`
- `DELETE /users/{id}`

## Basic Usage

### 1) Register user

```bash
curl -X POST http://localhost:9090/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ada",
    "lastName": "Lovelace",
    "username": "ada",
    "email": "ada@example.com",
    "password": "StrongP@ss1",
    "avatar": ""
  }'
```

### 2) Login (store refresh cookie)

```bash
curl -i -X POST http://localhost:9090/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"identity":"ada@example.com","password":"StrongP@ss1"}'
```

`accessToken` is returned in `data.accessToken`, and the refresh token is stored in `cookies.txt`.

### 3) Get authenticated user

```bash
curl http://localhost:9090/auth/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

### 4) Refresh session

```bash
curl -X POST http://localhost:9090/auth/refresh \
  -b cookies.txt -c cookies.txt
```

## Response Format

Success:

```json
{
  "success": true,
  "data": {},
  "status": 200
}
```

Error:

```json
{
  "success": false,
  "code": "UNAUTHORIZED",
  "error": "unauthorized",
  "status": 401
}
```

## E2E Tests

Available scripts:

- `scripts/test_user_endpoints.sh`
- `scripts/test_auth_endpoints.sh`

Run:

```bash
bash scripts/test_user_endpoints.sh
bash scripts/test_auth_endpoints.sh
```

Both scripts can auto-start the stack (`AUTO_START=1` by default) and use `curl` + `jq`.

## Project Structure

```text
cmd/                    # entrypoint
config/                 # config loading and validation
internal/app/           # dependency wiring and HTTP server setup
internal/http/          # handlers, middleware, requests/responses
internal/usecase/       # application logic
internal/repository/    # PostgreSQL data access
internal/domain/        # domain rules
migrations/             # initial SQL schema
scripts/                # e2e tests
```

## Security Notes

- Use a strong `AUTH_JWT_SECRET` in real environments
- Set `AUTH_REFRESH_COOKIE_SECURE=true` in production
- Configure CORS (`CORS_ALLOW_ORIGIN`, etc.) for your frontend
- Avoid `CORS_ALLOW_ORIGIN=*` outside development

## Current Status

Implemented and functional for core users + auth flows. A natural next step is stricter role-based authorization on admin endpoints and stronger security-event observability/auditing.
