# agentic-workflow-pi-acpx

Agentic workflow platform that orchestrates **acpx** flow runs over **pi** agents,
exposed through a Tauri desktop app. Coarse-grained durability via Temporal,
segment-level execution via acpx, fine-grained agent loop via pi.

- **Engine:** acpx (zigflow dropped). Runs are segmented at `checkpoint` nodes.
- **Backend:** Go (modular monorepo of siloed services).
- **Desktop:** Tauri v2 (Rust shell) + React + TypeScript + React Flow.
- **Persistence:** Postgres (schema-per-service) + Temporal dev-server (SQLite).
- **Bus:** NATS JetStream.
- **Canonical flow format:** versioned Flow IR (JSON), codegen'd to `.flow.ts`.

## Layout

```
proto/        gRPC contracts (buf)
pkg/          shared libs (config, obs, events, flowir, authn, health, runtime, testutil)
services/     siloed bounded contexts: agents, flow, orchestrator, gateway
apps/desktop/ Tauri v2 + React + React Flow
agents/       common pi agent + flow-author skill (prompts, guidelines)
db/migrations/<schema>/  goose migrations per service schema
deploy/       docker compose for Postgres + NATS + Temporal dev
scripts/      dev-up, db-reset, er-gen, etc.
docs/         spec/, adr/, runbook/, plan/
examples/     sample Flow IR + codegen golden output
tests/e2e/    end-to-end demo run
```

## Quick start

```sh
task dev-up        # postgres + nats + temporal dev + migrations
task generate      # buf generate
task build         # build every module
task test          # test every module
cd apps/desktop/web && npm install && npm run dev
```

See `docs/spec/overview.md` and `docs/plan/` for design.
