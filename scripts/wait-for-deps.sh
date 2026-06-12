#!/usr/bin/env bash
set -euo pipefail
echo "waiting for postgres..."
until docker exec awpa-postgres pg_isready -U awpa -d awpa >/dev/null 2>&1; do sleep 1; done
echo "waiting for nats..."
until curl -fsS http://localhost:8222/healthz >/dev/null 2>&1; do sleep 1; done
echo "waiting for temporal..."
until curl -fsS http://localhost:7233/ >/dev/null 2>&1 || nc -z localhost 7233 2>/dev/null; do sleep 1; done
echo "deps ready."
