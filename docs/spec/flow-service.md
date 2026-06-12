# flow-service

Owns Flow IR documents and the prompt-to-IR generation pipeline. Backing
schema: `flows.*`. Service manifest: `services/flow/service.yaml`. gRPC
contract: `proto/flows/v1/flows.proto`.

## Responsibilities

- CRUD on `Flow` and `FlowVersion`.
- `Validate` IR JSON against `pkg/flowir/schema/flow-ir.v1.json` plus the DAG
  checks in `pkg/flowir.DAGCheck`.
- `Codegen` an IR to a `.flow.ts` consumable by `acpx`.
- `Generate` (server-streaming) drives the `flow-author` pi agent to produce a
  new IR from a natural-language prompt.

The service does not execute flows. Execution is owned by the orchestrator.

## Owned schema

From `db/migrations/flows/001_init.sql`:

| Table                | Columns                                                                                              |
|----------------------|------------------------------------------------------------------------------------------------------|
| `flows.flow`         | `id UUID PK`, `name TEXT UNIQUE`, `description TEXT`, `created_at TIMESTAMPTZ`                       |
| `flows.flow_version` | `id UUID PK`, `flow_id UUID FK`, `version INT`, `ir_json JSONB`, `flow_ts TEXT`, `created_at`        |

Unique constraint `(flow_id, version)`. Index `idx_flow_version_flow` on
`flow_id`. `version` is monotonically incremented on each `PutVersion`.

## gRPC surface

`flows.v1.FlowService`:

| RPC          | Request → Response                          | Notes                                                                 |
|--------------|---------------------------------------------|-----------------------------------------------------------------------|
| `CreateFlow` | `CreateFlowRequest` → `CreateFlowResponse`  | Emits `flows.flow.created`.                                           |
| `ListFlows`  | `ListFlowsRequest` → `ListFlowsResponse`    | Cursor-paginated.                                                     |
| `GetFlow`    | `GetFlowRequest` → `GetFlowResponse`        | Returns the flow plus all versions, newest first.                     |
| `PutVersion` | `PutVersionRequest` → `PutVersionResponse`  | Validates IR (`flowir.Parse` + `flowir.DAGCheck`), runs `Codegen` to populate `flow_ts`, then writes. Emits `flows.flow.version.created`. |
| `Validate`   | `ValidateRequest` → `ValidateResponse`      | Pure; never touches the DB. Returns `errors[]`.                       |
| `Codegen`    | `CodegenRequest` → `CodegenResponse`        | Pure; emits the `.flow.ts` template.                                  |
| `Generate`   | `GenerateRequest` → stream `GenerateChunk`  | See below.                                                            |

### `Generate` streaming RPC

`GenerateChunk` is a `oneof`:

- `partial_ir` — stringly partial JSON as the author agent builds the document.
- `log_line` — model thinking / tool calls, mirrored to
  `flows.flow.generation.log`.
- `completed FlowVersion` — terminal success; the row has already been
  written and `flows.flow.version.created` emitted.
- `error common.v1.Error` — terminal failure; mirrored to
  `flows.flow.generation.error`.

Default author is the `flow-author` skill under `agents/`. The agent id may
be overridden via `GenerateRequest.author_agent_id` to A/B different skills.

Backpressure: the stream uses the gRPC HTTP/2 window. The service buffers at
most 64 chunks before blocking the author.

## Events produced

Subjects from `pkg/events/subjects.go`:

- `flows.flow.created` — `SubjectFlowCreated`
- `flows.flow.version.created` — `SubjectFlowVersionCreated`
- `flows.flow.generation.log` — `SubjectFlowGenerationLog`
- `flows.flow.generation.done` — `SubjectFlowGenerationDone`
- `flows.flow.generation.error` — `SubjectFlowGenerationError`

The generation `.log` subject is best-effort. The terminal `.done` /
`.error` subjects are required and durable.

## Config keys

Prefix `FLOW`. Required env from `service.yaml`: `FLOW_POSTGRES_DSN`.

```yaml
service: flow
grpc_addr: ":7002"
http_addr: ":7102"
postgres:
  dsn: "postgres://awpa:awpa@localhost:5432/awpa?sslmode=disable"
  schema: flows
nats:
  url: "nats://localhost:4222"
  stream: AWPA
stubs:
  author: real         # "stub" returns examples/flow-ir/hello-review.json
```

The `stubs.author` switch is honoured by the author-agent adapter and lets the
service run e2e without a working pi runtime.

## Failure modes

- **IR fails schema**: `Validate` returns errors; `PutVersion` returns
  `InvalidArgument` with the errors flattened to the gRPC status detail.
- **IR fails DAG checks**: same path; `DAGCheck` covers duplicate ids,
  unreachable nodes, unknown edge endpoints, decision branches, fork/join
  balance, end reachability.
- **Author agent unreachable**: `Generate` emits a terminal `error` chunk and
  publishes `flows.flow.generation.error`.
- **Codegen failure**: `PutVersion` does not write the row; the client may
  retry after fixing the IR.

## Capacity notes

- IR documents are small (KB range). Postgres JSONB is appropriate.
- `Generate` is the only long-lived RPC; expect minutes per call. Tune client
  timeouts accordingly.
- `flow_ts` is stored as TEXT to keep the desktop's editor view round-trippable
  without re-running codegen.

## See also

- [flow-ir.md](./flow-ir.md) — canonical IR schema.
- [orchestrator-service.md](./orchestrator-service.md) — consumer of `flow_ts`.
- [events.md](./events.md) — full subject catalog.
