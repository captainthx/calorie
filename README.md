# Simple Calorie App

Calorie tracking app with:

- `backend/` - Go + Gin + GORM + PostgreSQL API
- `frontend/` - Vite + React UI
- `backend/compose.yml` - PostgreSQL only, for local backend work
- `compose.yml` - PostgreSQL + backend + frontend, for running the full stack

## Quick Start

### Backend only

Create local env:

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

API base URL: `http://localhost:8080`

### Full stack

Run everything from the project root:

```bash
docker compose up -d --build
```

Open:

- Frontend: `http://localhost:3000`
- Backend: `http://localhost:8080`
- Swagger: `http://localhost:8080/docs`

Stop everything:

```bash
docker compose down
```

Wipe database volume too:

```bash
docker compose down -v
```

## Environment

Default backend values match `backend/compose.yml`:

```bash
GIN_MODE=debug
PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=myuser
DB_PASSWORD=mysecretpassword
DB_NAME=mydatabase
```

`.env` is optional if you pass env vars directly.

## Frontend Runtime Config

The frontend no longer bakes the API URL into the JS bundle.

Resolution order:

1. `window._env_.API_BASE_URL`
2. `VITE_API_BASE_URL`
3. `http://localhost:8080/api`

Relevant files:

- `frontend/public/env.js` - empty stub for local dev
- `frontend/index.html` - loads `env.js` before the React bundle
- `frontend/src/config.ts` - runtime fallback chain
- `frontend/src/services/api.ts` - reads `config.API_BASE_URL`
- `frontend/Dockerfile` - writes `env.js` when the container starts

Result: build the image once, then switch backend targets at runtime with `-e API_BASE_URL=...`.

## Auth Tokens

The API uses predefined bearer tokens.

- User John: `Authorization: Bearer user-token-123`
- User Jane: `Authorization: Bearer user-token-456`
- Admin: `Authorization: Bearer admin-token-789`

## API Routes

Public routes:

- `GET /ping` - simple liveness response
- `GET /health` - readiness check; returns `{"status":"ok"}` or `{"status":"unhealthy"}`
- `GET /docs` - Swagger UI redirect

Authenticated routes:

- `GET /api/food-entries`
- `POST /api/food-entries`
- `PUT /api/food-entries/:id`
- `PATCH /api/food-entries/:id`
- `DELETE /api/food-entries/:id`
- `GET /api/daily-summary`
- `GET /api/admin/food-entries`
- `POST /api/admin/food-entries`
- `GET /api/admin/reports`

## Response Shape

Business API success responses use:

```json
{
  "success": true,
  "data": {}
}
```

Error responses use:

```json
{
  "success": false,
  "error": "forbidden"
}
```

`GET /health` is the one exception and returns plain JSON so Docker and load balancers can probe it easily.

## Docker

### Backend image

Build:

```bash
docker build -t calorie-api ./backend
```

Run against a PostgreSQL host:

```bash
docker run -d -p 8080:8080 \
  -e GIN_MODE=debug \
  -e PORT=8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_USERNAME=myuser \
  -e DB_PASSWORD=mysecretpassword \
  -e DB_NAME=mydatabase \
  --name calorie-api \
  calorie-api
```

Stop and start:

```bash
docker stop calorie-api
docker start calorie-api
```

### Frontend image

Build once:

```bash
docker build -t calorie-frontend ./frontend
```

Run with the default backend:

```bash
docker run -d -p 3000:8080 --name calorie-ui calorie-frontend
```

Run with a custom backend URL:

```bash
docker run -d -p 3000:8080 \
  -e API_BASE_URL=http://your-backend:8080/api \
  --name calorie-ui \
  calorie-frontend
```

Stop and start:

```bash
docker stop calorie-ui
docker start calorie-ui
```

No rebuild is needed when only the backend URL changes.

### Compose logs

```bash
docker compose logs -f
docker compose logs -f backend
docker compose logs -f frontend
```

## Swagger

Swagger UI:

```bash
http://localhost:8080/docs
```

Regenerate docs after changing annotations:

```bash
cd backend
go run github.com/swaggo/swag/cmd/swag@latest init -g ./cmd/api/main.go -o ./docs
```

## Tests

Unit tests:

```bash
cd backend
go test ./...
```

Integration tests:

```bash
cd backend
docker compose up -d
go test -tags integration -v ./cmd/api/... -timeout 60s
```

Frontend build check:

```bash
cd frontend
npm run build
```

## CI

GitHub Actions workflows live in `.github/workflows/`.

Current checks include:

- `go test ./...`
- `go test -tags integration -v ./cmd/api/...`
- frontend build and lint checks
