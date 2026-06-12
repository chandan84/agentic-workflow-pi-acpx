# ADR 0004: Segment runs at checkpoint nodes under Temporal

Status: Accepted · Date: 2026-06-11

## Context

acpx executes a flow as one process; it has no durable pause for human
approval. Human gates can be open for hours or days — far longer than a
process or activity should live.

## Decision

The flow-service splits an IR into segments at `checkpoint` nodes. The
orchestrator-service runs one generic Temporal Workflow per run: each segment
is a Temporal Activity that shells out to `acpx flow run` for that segment
(heartbeating while it runs); between segments the Workflow awaits a
`CheckpointApproved` Signal with a durable Timer enforcing the checkpoint's
`timeoutSeconds`/`onTimeout` policy (escalate, abort, or auto-approve).

## Consequences

- Human gates survive worker restarts and deploys; Temporal history records
  every approval with the approver identity.
- Segment boundaries are the unit of retry: a failed segment re-runs from its
  start, so node authors must keep segment-internal side effects idempotent or
  acceptable to repeat.
- acpx trace bundles are per-segment; the audit view stitches them per run.
- Temporal never sees individual nodes — keeping its event history small and
  the Workflow code generic across all flows.
