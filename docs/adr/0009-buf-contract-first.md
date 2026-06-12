# ADR 0009: Contract-first gRPC with buf; generated Go committed

Status: Accepted · Date: 2026-06-11

## Context

Siloed services may only integrate through contracts. Those contracts need
linting, breaking-change detection, and codegen for Go now and TypeScript
(Connect-Web) next.

## Decision

`/proto` is the contract root, managed by buf (`buf.yaml`, `buf.gen.yaml`).
CI runs `buf lint` and `buf breaking` against `main`. Generated Go lives in
the committed `pkg/protogen` module so a clean clone builds without buf
installed; `task generate` refreshes it and CI fails if the output drifts
from the committed code.

## Consequences

- `go build` works offline and without protoc/buf — important for the
  desktop-first contributor experience.
- Generated-code diffs appear in PRs, making contract changes visible at
  review time.
- The TS/Connect-Web generation target is added to `buf.gen.yaml` when the
  desktop client wiring lands; the web app currently ships a transport
  placeholder (`apps/desktop/web/src/client/transport.ts`).
