# orchestrator-service

Temporal client + worker driving acpx flow runs segmented at `checkpoint` nodes.
Owns the `runs` Postgres schema; publishes `runs.*` events.

See `docs/spec/orchestrator-service.md`.
