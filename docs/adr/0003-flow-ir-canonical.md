# ADR 0003: Versioned Flow IR as the canonical flow format

Status: Accepted · Date: 2026-06-11

## Context

A flow exists in three shapes: the visual graph the user edits (React Flow),
the executable program acpx runs (`.flow.ts`), and the stored/versioned
artifact. Letting any of those be authoritative couples the others to it.

## Decision

A versioned JSON document — the Flow IR (`schemaVersion` field, JSON Schema at
`pkg/flowir/schema/flow-ir.v1.json`) — is the single source of truth. The
React Flow graph is a projection (`apps/desktop/web/src/lib/ir.ts`,
round-trippable). The `.flow.ts` is generated output (`pkg/flowir/codegen.go`),
never hand-edited. Validation (JSON Schema + DAG rules) happens on the IR.

## Consequences

- Flow versioning, diffing, and audit operate on one stable format.
- Schema evolution goes through a migration table keyed by `schemaVersion`;
  v1 documents stay loadable forever.
- Codegen changes can be verified with golden tests
  (`pkg/flowir/golden_test.go`, `examples/`).
- The editor must preserve unknown IR fields it does not render, or
  round-tripping breaks forward compatibility.
