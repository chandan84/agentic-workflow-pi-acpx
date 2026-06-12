# acpx Authoring Guideline

Conventions for designing Flow IR documents that compile cleanly to acpx
TypeScript and execute reliably under the orchestrator.

## Naming
- Node ids: lower-kebab-case. Stable across versions; do not rename to "improve
  wording" — that breaks audit trails.
- Flow id: kebab-case, prefix optional (e.g. `release-`, `support-`).

## Granularity
- Keep nodes small. One node, one job.
- Group multiple atomic actions only when they are guaranteed atomic in acpx
  (e.g. a single `compute` step). Otherwise, separate nodes so failures map to
  the right place in the audit timeline.

## Checkpoints
- Place a checkpoint at every "human in the loop" decision and at every
  irreversible side effect (publish, send, delete).
- `timeoutSeconds` should reflect the actual expected human SLA, not infinity.
- Prefer `onTimeout: escalate` over `auto-approve`. `auto-approve` is only
  appropriate for low-risk decisions.

## Compensations
- For action nodes with external side effects, attach a `compensate` field
  pointing at the compensating action. The orchestrator runs it on segment
  failure inside the same Temporal Activity attempt.

## Fork / Join
- A `fork` must always be paired with a `join`. The join waits for all paths.
- Do not use `fork` for "fan-out and discard" patterns; emit work items instead.

## Decisions vs Actions
- `decision` is for branching the flow based on data — its branches map
  conditions to target node ids.
- `action` is for side effects. It must not change control flow.

## Work items
- Use `workItems[]` to declare human or queued work attached to a node. Each
  work item carries a schema validated against by the desktop UI.

## Versioning
- IR is versioned via the `schemaVersion` field. Bump only via an ADR and a
  migration function in `pkg/flowir`.
