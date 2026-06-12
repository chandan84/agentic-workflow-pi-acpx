# ADR 0002: Go backend, Rust only for the Tauri shell

Status: Accepted · Date: 2026-06-11

## Context

The backend needs first-class gRPC, Temporal, NATS, and Postgres clients plus
cheap static binaries for service deployment. The desktop app needs a native
shell; Tauri v2 is Rust-based.

## Decision

All backend services are Go. Rust appears only in `apps/desktop/src-tauri`,
kept deliberately thin (window, IPC for shell config and secure storage —
no domain logic). The frontend is React + TypeScript.

## Consequences

- One backend toolchain: `go.work` workspace, golangci-lint, single CI matrix.
- The Tauri shell can be maintained without deep Rust expertise because it
  contains no business logic; everything testable lives in Go or TypeScript.
- Domain logic must never leak into IPC commands — the web layer talks to the
  gateway-service directly over Connect-Web.
