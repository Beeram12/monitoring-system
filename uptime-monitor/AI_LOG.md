# AI Collaboration Log

## AI Tech Stack

- **Assistant**: Claude Code (Anthropic), model **Claude Sonnet 4.6**, used as an interactive
  CLI agent with direct filesystem/shell/Docker access 
- No other AI tools were used in this pass; the whole backend, frontend, and infra were
  generated and refined in one ongoing agent session.

## My Role vs. the AI's Role

- I designed the overall architecture going in — relational DB over NoSQL (the data is two fixed-shape, relationally-linked entities: a monitor and its checks, with no need for
schema flexibility at this scale)
- Go for the backend, an interface-driven layering
(domain → service → repository → handler) so storage and the ping mechanism stay swappable
behind interfaces, and a goroutine-per-monitor in-process scheduler instead of a separate
worker service for MVP simplicity. 
- I asked the AI to implement that design, lent it the two
provided frontend design themes and code to pull reusable shadcn/ui pieces from, and picked **which** scaffold
to base the dashboard's visual theme on (`tasko-modern-task-management-dashboard`'s green/white
card-and-table look) rather than letting the AI choose a theme on its own.

- Where I leaned on the AI more heavily was the internal mechanics of the scheduler — the
goroutine-per-monitor reconcile loop, the `context.WithCancel` per-monitor lifecycle, and the
`sync.Mutex`-guarded `running` map that lets monitors be started/stopped without restarting the
process. 
- I knew the shape I wanted (one independent ping loop per monitor, reconciled against
the DB on an interval) but asked the AI to work out the concrete goroutine/channel mechanics and
then walked through dry-run traces with it afterward to confirm I understood exactly how the
map and cancel functions interacted before accepting it.

## The Prompts that Shipped It

The project started from one instruction:

- > "Use Golang as the language and a relational DB for storage/endpoints. You can use the two
provided frontend scaffolds (`optimus-the-ai-platform...` and
`tasko-modern-task-management-dashboard`) for the frontend. Make sure all assignment details
and constraints are met, follow strict backend/frontend guidelines and directory structure,
keep the codebase clean, and use interfaces when designing it."

From there:
- The agent read the assignment PDF directly to extract the actual requirements (register
   URLs, ping periodically, store status code/response time/timestamp, dashboard, single
   `docker compose up`, deployment sketch, AI log) instead of working off a paraphrase.
- It asked clarifying questions rather than guessing silently, and I made the calls:
   - Repurpose the full `tasko` task-management template, or build a fresh minimal dashboard
     reusing only its shadcn/ui primitives? — **I chose the fresh minimal dashboard**, since
     carrying over unrelated pages (tasks/team/calendar) into a strict-MVP monitor contradicts
     the assignment's "beautifully simple" guidance. I kept `tasko`'s visual theme (color
     palette, card/table styling) as the dashboard's look.
   - Go framework — Gin vs. plain `net/http`+chi vs. Echo — **I picked Gin**.
   - DB access pattern — **I explicitly told it to use sqlc instead of an ORM** (rejecting GORM
     outright): sqlc keeps SQL explicit and hand-written instead of hidden behind ORM query
     generation, and its generated `Querier` interface plugs directly into an interface-driven
     repository layer, which matched the "use interfaces" requirement I'd set from the start.
   - Scheduling model — in-process goroutine+ticker vs. a separate worker container — **I chose
     in-process** for MVP simplicity, with the explicit condition that it had to be safe to
     add/remove monitors at runtime without restarting the process.
- It generated, in order: Postgres schema + migrations, sqlc query files, the `domain` package
   (interfaces: `MonitorRepository`, `CheckRepository`, `Pinger`), Postgres repository adapters,
   the `httpclient.Pinger`, the service layer, Gin handlers, the scheduler, `main.go` wiring,
   then the Next.js dashboard (API client, types, add-monitor dialog, status table, polling
   page), then Dockerfiles and `docker-compose.yml`.
- It verified the result by actually running `docker compose up --build`, registering a real
   healthy URL (`https://example.com`) and a deliberately broken one
   (`http://localhost:9999/nope`) against the live API, and confirming the JSON showed
   `is_up: true` / `is_up: false` respectively before calling it done.

## Cross-Questions and Follow-Up Changes I Asked For

After the first working version, I went through the generated backend file by file and pushed
back on several structural decisions rather than accepting the first pass as final:

- **"Why are env vars hardcoded with defaults in `config.go`? Load them from a `.env` file
  instead."** The original `config.Load()` returned sane defaults baked into the Go source if
  an env var was unset. I had it rewritten to use `godotenv`, with every variable now
  **required** (the app fails fast with a clear error if one is missing) instead of silently
  falling back to a hardcoded value — config should live in `.env`/`.env.example`, not in code.
- **"Separate the monitor handler from the checks handler — keep it clean."** I asked why
  monitor CRUD and check/status reads were in one `monitor_handler.go` file and had it split
  into `monitor_handler.go` (monitor CRUD only) and `check_handler.go` (check history +
  combined status feed), with the corresponding service-layer method (`ListChecks`) moved off
  `MonitorService` and onto `CheckService` so it lives next to the check-recording logic it's
  paired with.
- **"Keep all routes in one file under a `routes` directory — don't scatter route registration
  across handler files."** I pushed back on each handler owning its own `RegisterRoutes`
  method, which meant the HTTP surface of the app was defined in two different places.
- **Repeated cross-questioning on the data layer**: I asked it to justify, file by file, why two
  tables instead of one, what each column was for, why SQL over NoSQL, and whether write
  throughput favored either — making it defend the schema design rather than taking it at face
  value, and only after that walked through whether all repository inputs were actually
  validated end to end (see below).

## The Course Corrections

**Everything was mixed into single files instead of properly separated — I had to explicitly
ask for the split.** The first pass put all HTTP route registration *inside* each handler
(`MonitorHandler.RegisterRoutes`, `CheckHandler.RegisterRoutes`), and originally both monitor
and check/status endpoints lived in one undifferentiated `monitor_handler.go`. This isn't broken
code, but it's exactly the kind of structural sloppiness that compounds as a codebase grows: the
HTTP surface of the app was defined in N different places instead of one, and a file named
"monitor" silently also owned check-history and status-aggregation endpoints. I explicitly
called this out and asked for:
1. `monitor_handler.go` and `check_handler.go` split by actual responsibility, with handler
   methods exported (`Create`/`List`/`Delete`, `ListChecks`/`ListStatuses`) instead of private,
   and the corresponding service method (`ListChecks`) relocated from `MonitorService` to
   `CheckService` to match.
2. A dedicated `internal/routes/routes.go` as the single place every route is registered,
   with `main.go` simply calling `routes.Register(r, monitorHandler, checkHandler)` instead of
   each handler wiring its own paths.
Both changes were mechanical refactors with no behavior change — verified by rebuilding
(`go build ./...`, `go vet ./...`), re-running `docker compose up --build`, and re-confirming the
up/down detection test still passed against the live API after the split.

3. **Input validation was missing and had to be pointed out, not just file organization.** When I
asked directly whether all input checks in `check_repository.go` were actually handled, the
honest answer was: the repository layer itself does no validation (correctly — it trusts its
caller and only translates `pgx.ErrNoRows` into a nil result), but tracing every parameter back
to its origin surfaced a real gap: an HTTP request to `GET /api/monitors/:id/checks` with an
`id` that parses as a valid int64 but doesn't correspond to any real monitor silently returns
`200 OK` with an empty list instead of a `404`, because Postgres just returns zero rows rather
than erroring. The only enforcement that *does* exist is at the schema level (`checks.monitor_id
REFERENCES monitors(id) ON DELETE CASCADE`), not in application code. This was flagged explicitly
as a known, currently-unaddressed gap rather than silently left in — a real instance of the
agent not catching its own omission until directly cross-examined on it, not something it
volunteered up front.

No libraries were hallucinated and no architecture had to be thrown out wholesale; the
corrections above were about forcing better separation of concerns and surfacing a validation
gap that the initial implementation didn't flag on its own.
