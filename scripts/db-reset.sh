#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"
docker exec -i awpa-postgres psql -U awpa -d awpa <<'SQL'
DROP SCHEMA IF EXISTS agents CASCADE;
DROP SCHEMA IF EXISTS flows  CASCADE;
DROP SCHEMA IF EXISTS runs   CASCADE;
DROP SCHEMA IF EXISTS audit  CASCADE;
SQL
"$HERE/migrate-all.sh"
