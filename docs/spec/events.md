# Event catalog

NATS JetStream subjects produced by the backend. Single source of truth:
`pkg/events/subjects.go`. The constants and the `AllSubjects()` helper are
referenced by every service; do not introduce subjects outside this file.

## Conventions

- Subject names follow `<context>.<entity>.<verb>` (e.g.
  `runs.checkpoint.awaiting`). The leading context segment matches the
  owning bounded context.
- Payloads are JSON-encoded proto messages from the corresponding service
  contract.
- Stream: `AWPA`. All subjects publish into the same JetStream stream.
- Producers are *exactly one* per subject. Multiple consumers per subject
  are allowed.
- Retention is `WorkQueue` with `MaxAge=72h` by default. `audit.event` is
  promoted to `Limits` retention with no `MaxAge` so audit history survives.

## Catalog

| Subject                          | Constant                          | Producer       | Consumers              | Payload                                | Retention |
|----------------------------------|-----------------------------------|----------------|------------------------|----------------------------------------|-----------|
| `agents.agent.created`           | `SubjectAgentCreated`             | agents-service | gateway                | `agents.v1.Agent`                      | 72h       |
| `agents.group.created`           | `SubjectGroupCreated`             | agents-service | gateway                | `agents.v1.AgentGroup`                 | 72h       |
| `agents.resource.linked`         | `SubjectResourceLinked`           | agents-service | gateway                | `agents.v1.AgentResource`              | 72h       |
| `flows.flow.created`             | `SubjectFlowCreated`              | flow-service   | gateway                | `flows.v1.Flow`                        | 72h       |
| `flows.flow.version.created`     | `SubjectFlowVersionCreated`       | flow-service   | gateway, orchestrator  | `flows.v1.FlowVersion`                 | 72h       |
| `flows.flow.generation.log`      | `SubjectFlowGenerationLog`        | flow-service   | gateway                | `{ flow_id, log_line }`                | 24h       |
| `flows.flow.generation.done`     | `SubjectFlowGenerationDone`       | flow-service   | gateway                | `flows.v1.FlowVersion`                 | 72h       |
| `flows.flow.generation.error`    | `SubjectFlowGenerationError`      | flow-service   | gateway                | `common.v1.Error`                      | 72h       |
| `runs.run.started`               | `SubjectRunStarted`               | orchestrator   | gateway                | `execution.v1.Run`                     | 72h       |
| `runs.run.updated`               | `SubjectRunUpdated`               | orchestrator   | gateway                | `execution.v1.Run`                     | 72h       |
| `runs.run.completed`             | `SubjectRunCompleted`             | orchestrator   | gateway                | `execution.v1.Run`                     | 72h       |
| `runs.run.failed`                | `SubjectRunFailed`                | orchestrator   | gateway                | `execution.v1.Run`                     | 72h       |
| `runs.checkpoint.awaiting`       | `SubjectRunCheckpointWait`        | orchestrator   | gateway                | `execution.v1.CheckpointRequest`       | 72h       |
| `runs.checkpoint.resolved`       | `SubjectRunCheckpointDone`        | orchestrator   | gateway                | `execution.v1.CheckpointRequest`       | 72h       |
| `runs.segment.started`           | `SubjectRunSegmentStart`          | orchestrator   | gateway                | `{ run_id, from_node, to_node }`       | 72h       |
| `runs.segment.finished`          | `SubjectRunSegmentEnd`            | orchestrator   | gateway                | `{ run_id, from_node, to_node, output }` | 72h     |
| `runs.workitem.emitted`          | `SubjectRunWorkItemEmitted`       | orchestrator   | gateway                | `execution.v1.WorkItem`                | 72h       |
| `audit.event`                    | `SubjectAuditEvent`               | all services   | gateway, audit reader  | `audit.v1.AuditEvent`                  | unlimited |

## Producer rules

- Producers are listed in each service's `service.yaml` under
  `streams.produced`. Anything not declared there is a violation of the
  integration rule and should fail CI.
- Synchronous gRPC handlers publish their lifecycle event in the same
  transaction context as the DB write where possible. If the publish fails,
  the gRPC call still returns success — the consumer-side reconciliation
  via `audit.event` is what makes this safe.

## Consumer rules

- All gateway consumers are durable JetStream consumers with
  `DeliverPolicy=All` for catch-up after restart.
- `audit.event` is never consumed by services that wrote the event — it is
  exclusively a sink for downstream readers (the audit reader and the
  gateway timeline aggregator).
- Stream filtering by `run_id` is done at the consumer with subject filters
  unavailable; payload inspection is acceptable because event volume is
  low.

## Adding a new subject

1. Add the constant in `pkg/events/subjects.go` and include it in
   `AllSubjects()`.
2. Declare the producer in that service's `service.yaml`.
3. Update this table.
4. Bump consumers as needed.
