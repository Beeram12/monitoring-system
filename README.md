# Uptime Monitor

A lightweight, full-stack uptime monitor. Register URLs, the backend pings each one on its
own interval, and the dashboard shows live up/down status and response times.

- **Backend**: Go (Gin), PostgreSQL (via `pgx` + `sqlc`), in-process scheduler.
- **Frontend**: Next.js (App Router) + Tailwind + shadcn/ui, polls the API every 5s.
- **Database**: PostgreSQL, schema applied automatically via embedded migrations on boot.

## Architecture

```
backend
├── cmd/api/main.go              entrypoint: wiring, HTTP server, graceful shutdown
├── config/config.go             env-based configuration
├── internal
│   ├── domain/                  core types + interfaces (MonitorRepository, CheckRepository, Pinger)
│   ├── service/                 business logic, depends only on domain interfaces
│   ├── handler/                 Gin HTTP handlers, request/response DTOs
│   ├── routes/routes.go         single place every HTTP route is registered
│   ├── scheduler/scheduler.go   goroutine-per-monitor ticker loop that triggers pings
│   ├── repository/postgres/     sqlc-backed implementation of the domain repositories
│   └── db/                      embedded SQL migrations + sqlc-generated query code
├── pkg/httpclient/pinger.go     Pinger implementation (plain net/http)
└── Dockerfile

frontend
├── app/                Next.js App Router pages
├── components/         dashboard UI (monitor table, add-monitor dialog) + shadcn/ui primitives
├── lib/                 typed API client
└── Dockerfile
```

The backend is organized around interfaces in `internal/domain` (`MonitorRepository`,
`CheckRepository`, `Pinger`) so the service layer and scheduler depend on behavior, not on
Postgres or `net/http` directly — swapping storage or the ping mechanism doesn't touch business
logic.

## 1-Line Setup

```bash
docker compose up --build
```

This starts three containers: `postgres` (16-alpine), `backend` (Go API on `:8080`, runs DB
migrations on boot), and `frontend` (Next.js on `:3000`).

Once it's up:
- Dashboard: http://localhost:3000
- API: http://localhost:8080/api

## Testing Steps (verifying up/down detection)

1. Start the stack: `docker compose up --build`
2. Open http://localhost:3000 — the table is empty.
3. Click **Add URL**, register a healthy URL:
   - URL: `https://example.com`
   - Interval: `10` seconds
4. Click **Add URL** again, register a deliberately broken URL:
   - URL: `http://localhost:9999/this-does-not-exist`
   - Interval: `10` seconds
5. Within ~10 seconds the table updates (it polls every 5s):
   - `example.com` shows a green **Up** badge, HTTP `200`, and a response time in ms.
   - The broken URL shows a red **Down** badge with no successful response.

You can do the same against the raw API instead of the UI:

```bash
# Register a healthy URL
curl -X POST localhost:8080/api/monitors \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","name":"Example","interval_sec":10}'

# Register an unreachable URL
curl -X POST localhost:8080/api/monitors \
  -H "Content-Type: application/json" \
  -d '{"url":"http://localhost:9999/nope","name":"Broken","interval_sec":10}'

# After ~10s, check statuses
curl localhost:8080/api/statuses | python3 -m json.tool
```

You should see one entry with `"is_up": true, "status_code": 200` and another with
`"is_up": false`, an `error` message, and `"status_code": 0`.

## API Reference

| Method | Path                        | Description                                  |
|--------|-----------------------------|-----------------------------------------------|
| GET    | `/health`                   | Liveness check                                |
| POST   | `/api/monitors`              | Register a monitor `{url, name, interval_sec}` |
| GET    | `/api/monitors`              | List registered monitors                      |
| DELETE | `/api/monitors/:id`          | Remove a monitor                               |
| GET    | `/api/monitors/:id/checks`   | Recent checks for one monitor                  |
| GET    | `/api/statuses`              | Every monitor + its latest check (dashboard feed) |

## Configuration

The backend loads all configuration from `backend/.env` (via `env_file:` in
`docker-compose.yml`, or directly via `godotenv` when running `go run` locally). There are no
hardcoded defaults in code — every variable below is required at startup. Copy
`backend/.env.example` to `backend/.env` and adjust as needed; a working `.env` for
`docker compose up` is already checked in for convenience (not committed in a real project).

| Variable                  | Example                                                        |
|----------------------------|-----------------------------------------------------------------|
| `PORT`                     | `8080`                                                          |
| `DATABASE_URL`              | `postgres://postgres:postgres@postgres:5432/uptime?sslmode=disable` |
| `PING_TIMEOUT`              | `10s`                                                           |
| `SCHEDULER_POLL_INTERVAL`   | `10s` (how often the scheduler reconciles its monitor list)     |
| `ALLOWED_ORIGIN`            | `http://localhost:3000`                                         |

Frontend: `NEXT_PUBLIC_API_BASE_URL` (default `http://localhost:8080`, set via `docker-compose.yml` build args).

## Deployment Sketch (Light)

- For a real deployment, each container would map to a managed cloud service: the Postgres
container becomes a managed database (AWS RDS / GCP Cloud SQL), the backend becomes a container
running on a serverless container platform (AWS ECS Fargate / GCP Cloud Run), and the frontend
deploys to Vercel (or another Fargate/Cloud Run service behind a CDN). A load balancer would sit
in front of the backend with HTTPS, and the frontend's `NEXT_PUBLIC_API_BASE_URL` would point at
the backend's public URL instead of `localhost`.

- One thing to watch: the scheduler currently runs in-process inside the backend, so it's only
safe to run a single backend replica — running multiple would mean every replica pings the same
URLs and writes duplicate check rows. Scaling the API horizontally would require splitting the
scheduler into its own single-instance service first.

## AI Collaboration

See [AI_LOG.md](./AI_LOG.md) for the AI tooling used, representative prompts, and a documented
course-correction.
