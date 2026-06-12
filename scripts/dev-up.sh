#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"
docker compose -f "$ROOT/deploy/compose.yaml" up -d
"$HERE/wait-for-deps.sh"
"$HERE/migrate-all.sh"
echo "dev stack ready."
