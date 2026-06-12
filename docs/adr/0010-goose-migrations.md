# ADR 0010: goose for schema migrations

Status: Accepted · Date: 2026-06-11

## Context

Schema-per-service (ADR 0006) means four independent migration histories that
must be runnable from scripts, CI, and the dev-up flow without an ORM.

## Decision

Plain-SQL goose migrations, one directory per schema under
`db/migrations/<schema>/`, each with its own `goose_db_version` table inside
that schema. `scripts/migrate-all.sh` applies all of them in dependency-free
order; `scripts/db-reset.sh` drops and re-applies. New migrations are created
with `task migrate-new -- <schema> <name>`.

## Consequences

- Migrations are reviewable SQL; no generated DDL surprises.
- Down-migrations are required for every up — `db-reset` depends on clean
  re-runs rather than rollbacks, but dev iteration uses `goose down`.
- Services never run migrations at boot; applying them is an explicit ops
  step (dev: `dev-up`; prod: release runbook).
