# Roadmap

From this skeleton to GA. The order favours pulling unknowns forward —
runtime durability, multi-tenant, packaging — over polish.

## Phase 0 — Skeleton (now)

Where the repo is today.

- All four services compile and pass tests with stubs (`stubs.author=stub`,
  `stubs.runtime=stub`).
- gRPC contracts are stable.
- IR v1 schema is the canonical format. `flowir.Parse` and
  `flowir.DAGCheck` are wired into `FlowService.Validate` and `PutVersion`.
- Temporal workflow `RunFlow` segments at checkpoints and routes signals.
- Desktop renders Groups, Flows, Runs, Audit, Settings against the
  gateway.
- docker compose brings up Postgres, NATS JetStream, and Temporal dev.

## Phase 1 — Internal alpha (working e2e with the real runtime)

- Replace `stubs.runtime=stub` with the real acpx invocation; ship the
  `flow run --from --to` CLI surface acpx needs.
- Replace `stubs.author=stub` with the real flow-author pi skill.
- End-to-end `examples/flow-ir/hello-review.json` demo runs from
  `task dev-up` (see [docs/runbook/demo.md](../runbook/demo.md)).
- Audit aggregator (gateway) merges Temporal history, acpx trace bundles,
  and `audit.audit_event` into one timeline.
- Metrics under `awpa_*` exported; Prometheus scrape target documented.

## Phase 2 — Hardening (multi-user, multi-host)

- AuthN: implement `OIDCAuthenticator` in `pkg/authn`. Wire to the gateway.
  Remove `auth.mode=none` from non-dev defaults.
- Multi-tenant: introduce a `tenant_id` column on every table; gateway
  stamps it from the principal. Migrations under
  `db/migrations/<schema>/002_tenant.sql`.
- mTLS between gateway and backend services for non-loopback deployments.
- Action `fn` registry: explicit whitelist + per-`fn` signature check at
  IR validation.
- Approver groups: resolve `approvers[]` strings against agents-service
  group membership.
- Backpressure on JetStream: durable consumers with `MaxAckPending` tuned;
  failed publishes surfaced through the audit layer.

## Phase 3 — Beta (closed-circle deploys)

- Packaging: signed Tauri builds for macOS and Windows; Linux AppImage.
- Backend distribution: single OCI image per service plus a compose
  manifest under `deploy/`.
- Migrations bundled and applied automatically on service startup
  (current behaviour: explicit `task db-migrate`).
- Resilience tests: kill a service mid-run; assert recovery from Temporal
  history and `runs.run.temporal_run_id`.
- Performance: 100 concurrent runs on a single orchestrator process,
  median segment under 30s.

## Phase 4 — GA

- SLOs: per-service availability targets and run-success error budgets in
  `docs/spec/observability.md`.
- Backup and restore runbooks for Postgres and Temporal SQLite (or
  Temporal Cluster).
- Upgrade path documented for IR `schemaVersion` bumps; v1↔v2 dual-read.
- Public docs site generated from `docs/`.
- Hardened auth: OIDC mandatory; bearer-token mode reserved for
  single-tenant on-prem.

## Out of scope (post-GA)

- Browser-only build of the desktop UI.
- Real-time collaborative editing of an IR document.
- Cross-region multi-orchestrator topology.
