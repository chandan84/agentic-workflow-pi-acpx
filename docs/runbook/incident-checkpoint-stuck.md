# Runbook: run stuck at a checkpoint

Symptom: a run sits in `awaiting_checkpoint` although an approver claims to
have approved it (or no approval prompt ever appeared).

## Diagnose

1. **Find the checkpoint row**: in the `runs` schema,
   `select * from checkpoint_request where run_id = $1 order by created_at desc;`
   - `status = awaiting` → the approval never reached the orchestrator
   - `status = approved` but run not progressing → Signal delivery problem
2. **Check Temporal**: `temporal workflow show --workflow-id run-<runId>`
   (dev-server UI at http://localhost:8233). Look for the pending
   `CheckpointApproved` Signal await and whether the timeout Timer fired.
3. **Check the event path**: did `runs.checkpoint.requested` reach the
   gateway? `nats stream view RUNS` (subjects in `docs/spec/events.md`).
4. **Check workers**: orchestrator log for `worker started`; a worker that
   can't reach Temporal leaves Workflows un-progressed with no error visible
   to users.

## Remediate

- **Lost approval**: re-send through the gateway
  (`ResolveCheckpoint`), or directly:
  `temporal workflow signal --workflow-id run-<runId> --name CheckpointApproved --input '{"checkpointId":"...","decision":"approved","approver":"ops"}'`
- **Dead worker**: restart orchestrator-service; Temporal redelivers
  automatically — no run state is lost (that is the point of ADR 0004).
- **Timer already fired with `onTimeout: abort`**: the run is terminally
  failed; start a new run. Record the cause in the audit trail.

## Afterwards

File the gap: every occurrence of this runbook should produce either a fix or
a new alert (see `docs/spec/observability.md`, metric
`checkpoint_await_age_seconds`).
