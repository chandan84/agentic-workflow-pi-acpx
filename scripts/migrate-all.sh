#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"
DB_URL="${DB_URL:-postgres://awpa:awpa@localhost:5432/awpa?sslmode=disable}"

for schema in agents flows runs audit; do
  echo "migrating schema: $schema"
  if command -v goose >/dev/null 2>&1; then
    goose -dir "$ROOT/db/migrations/$schema" postgres "$DB_URL" up
  else
    echo "  goose not installed; applying via psql"
    docker exec -i awpa-postgres psql -U awpa -d awpa <<SQL
CREATE SCHEMA IF NOT EXISTS $schema;
SQL
    for f in "$ROOT/db/migrations/$schema"/*.sql; do
      [ -f "$f" ] || continue
      docker exec -i awpa-postgres psql -U awpa -d awpa < "$f"
    done
  fi
done
