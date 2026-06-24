# Simple Calorie App

Go + Gin + GORM + PostgreSQL API for tracking food entries, daily calorie summaries, monthly price limits, and admin reports.

## Project Layout

- `backend/` - Go API server.
- `fontend/` - Frontend app workspace.
- `backend/compose.yml` - Local PostgreSQL database for development and integration tests.
- `backend/cmd/api/api_test.go` - Automated integration tests that exercise routes, middleware, services, repositories, and PostgreSQL.
- `backend/internal/food/service_test.go` - Unit tests for service logic using spy mocks.

## Local Setup

Create a local env file:

```bash
cp backend/.env.example backend/.env
```

Start PostgreSQL:

```bash
cd backend
docker compose up -d
```

Run the API:

```bash
cd backend
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default.

## Environment

Default local values match `backend/compose.yml`.

```bash
MODE=debug
PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=myuser
DB_PASSWORD=mysecretpassword
DB_NAME=mydatabase
```

`.env` is optional when environment variables are supplied directly, which is useful for Docker and CI.

`CORS_ALLOWED_ORIGINS` is a comma-separated allowlist for browser frontends. For local development the defaults cover common frontend ports `3000` and `5173`.

## Auth Tokens

This API uses predefined bearer tokens, not JWT login.

- User John: `Authorization: Bearer user-token-123`
- User Jane: `Authorization: Bearer user-token-456`
- Admin: `Authorization: Bearer admin-token-789`

## Common API Routes

- `GET /ping` - basic app ping.
- `GET /api/food-entries` - list current user's entries, optional `date_from=YYYY-MM-DD&date_to=YYYY-MM-DD`.
- `POST /api/food-entries` - create current user's entry.
- `PUT /api/food-entries/:id` - full update current user's entry.
- `PATCH /api/food-entries/:id` - partial update current user's entry.
- `DELETE /api/food-entries/:id` - delete current user's entry.
- `GET /api/daily-summary` - current user's summary, optional `date=YYYY-MM-DD`.
- `GET /api/daily-summaries` - current user's summary range, requires `date_from` and `date_to`.
- `GET /api/admin/food-entries` - admin list all entries.
- `POST /api/admin/food-entries` - admin create entry for any user.
- `GET /api/admin/reports` - admin report.

## Tests

Run regular unit tests:

```bash
cd backend
go test ./...
```

Run integration tests:

```bash
cd backend
docker compose up -d
go test -tags integration -v ./cmd/api/... -timeout 60s
```

The integration tests truncate and reseed `food_entries`, then hit the real Gin router and PostgreSQL through `httptest`.

## Docker

Build the backend image:

```bash
cd backend
docker build -t calorie-api .
```

Run against the local Compose database from macOS/Windows:

```bash
docker run --rm -p 8080:8080 \
  -e MODE=release \
  -e PORT=8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_USERNAME=myuser \
  -e DB_PASSWORD=mysecretpassword \
  -e DB_NAME=mydatabase \
  calorie-api
```

For Linux, use a Docker network or set `DB_HOST` to a reachable PostgreSQL host.

## CI

GitHub Actions workflow lives at `.github/workflows/tests-and-publish-image.yml`.

It runs:

- `go test ./...`
- `go test -tags integration -v ./cmd/api/...`

The integration job starts a PostgreSQL service with the same local credentials.

## Branching

Use a small trunk-based flow until the project needs release branches.

- `main` is always deployable.
- Feature work uses `feature/<short-name>` or `codex/<short-name>`.
- Open PRs into `main`; merge only after CI passes.
- Use protected `main` when the repo moves to shared team work.
- Add `release/<version>` branches only when production hotfixes need to diverge from active feature work.
