# Runbook: bring up the local dev stack

Audience: anyone developing against the backend. Time: ~2 minutes.

## Prerequisites

- Docker with the compose plugin
- Go 1.22+, [Task](https://taskfile.dev), and `goose` on PATH

## Steps

1. `task dev-up`
   - starts Postgres, NATS (JetStream enabled), and the Temporal dev-server
     from `deploy/compose.yaml`
   - waits for each dependency to accept connections
     (`scripts/wait-for-deps.sh`)
   - applies all goose migrations (`scripts/migrate-all.sh`)
2. Start the services you need, e.g.:
   ```sh
   (cd services/agents && go run ./cmd/agents)
   (cd services/gateway && go run ./cmd/gateway)
   ```
   Each reads its `config.example.yaml` defaults; override with
   `<SVC>_`-prefixed env vars (see `docs/spec/config.md`).
3. Verify: `curl -fsS http://localhost:<healthPort>/healthz` per
   `services/*/service.yaml`. `/readyz` turns 200 once Postgres/NATS are
   reachable from that service.
4. Frontend: `npm --prefix apps/desktop/web run dev` (browser) or
   `npm --prefix apps/desktop/web run tauri:dev` (shell).

## Teardown

`task dev-down` stops the compose stack; volumes persist. For a clean slate
see [db-reset.md](db-reset.md).

## Troubleshooting

- **Port collisions** — ports are declared in `deploy/compose.yaml` and
  `services/*/service.yaml`; adjust there, not in code.
- **`/readyz` stays 503** — that service can't reach a dependency; its log
  line `readiness check failed` names which one.
- **Migrations fail mid-way** — each schema migrates independently; re-running
  `scripts/migrate-all.sh` is safe (goose is idempotent per version).
