# Runbook: reset the local database

Destroys **all local data** in every service schema. Dev-only — there is
deliberately no production equivalent of this script.

## Steps

1. Ensure the compose stack is running (`task dev-up` if not).
2. `task db-reset`
   - drops the `agents`, `flows`, `runs`, and `audit` schemas
   - re-applies every goose migration from `db/migrations/<schema>/`
3. Restart any running services so in-memory caches and prepared statements
   don't reference dropped objects.

## Notes

- Temporal dev-server state (SQLite inside the container) is **not** touched;
  to clear run histories too: `task dev-down && docker volume prune` then
  `task dev-up`.
- NATS JetStream streams are recreated by services on boot; resetting the DB
  without restarting services can leave events referencing rows that no
  longer exist — restart, don't mix.
