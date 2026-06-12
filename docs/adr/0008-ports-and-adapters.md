# ADR 0008: Ports-and-adapters per service

Status: Accepted · Date: 2026-06-11

## Context

Each service depends on heavyweight externals — Postgres, NATS, Temporal, and
the acpx/pi binaries. Tests and the compiling-skeleton phase must run without
any of them.

## Decision

Every outbound dependency is an interface ("port") declared in the service's
`internal/ports` package: `Store`, `EventBus`, `WorkflowClient`,
`FlowEngine`, `AgentRuntime`. Two implementations ship side by side: the real
adapter (`internal/store`, etc.) and an in-memory stub (`internal/stub`).
Selection is config-driven (`adapters: { store: postgres | stub }`).

## Consequences

- Every service boots and serves health checks with zero infrastructure
  (`stub` everything) — used by smoke tests and CI.
- Real adapters can be filled in incrementally behind stable interfaces.
- The cost is interface upkeep: a new dependency must come in through a port,
  never as a direct client call from `internal/app`.
