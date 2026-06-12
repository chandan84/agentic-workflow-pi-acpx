# ADR 0006: One Postgres, schema-per-service

Status: Accepted · Date: 2026-06-11

## Context

Bounded contexts need isolated persistence, but a desktop-first product
cannot ask users to run four database servers. Temporal dev-server brings its
own SQLite.

## Decision

A single Postgres instance hosts one schema per service (`agents`, `flows`,
`runs`, `audit`). Each service connects with credentials whose
`search_path` is its own schema and owns its migrations
(`db/migrations/<schema>/`). Cross-schema queries and foreign keys are
forbidden; cross-context references are plain ID columns validated at the
application layer.

## Consequences

- Service isolation is preserved while ops stays "one database".
- Any service can later be pointed at its own Postgres instance by changing
  only its DSN — no code change.
- Referential integrity across contexts is eventual and app-enforced; the
  audit trail records dangling-reference repairs rather than the DB
  preventing them.
