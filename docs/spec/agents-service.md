# agents-service

Owns agent groups, agents, and resource bundles. Backing schema: `agents.*`.
Service manifest: `services/agents/service.yaml`. gRPC contract:
`proto/agents/v1/agents.proto`.

## Responsibilities

- CRUD for `AgentGroup`, `Agent`, `ResourceBundle`.
- Attach resources (MCP servers, tools, filesystem mounts, secrets) to agents.
- Emit lifecycle events for downstream consumers (gateway aggregation,
  orchestrator agent lookups).

The service does not invoke agents — that is the orchestrator's job via
`pkg/runtime.AgentRuntime`. It is the system of record for *agent identity and
its bundled resources only*.

## Owned schema (excerpt)

From `db/migrations/agents/001_init.sql`:

| Table                     | Columns (PK first)                                                                                  |
|---------------------------|-----------------------------------------------------------------------------------------------------|
| `agents.agents_group`     | `id UUID`, `name TEXT UNIQUE`, `description TEXT`, `created_at TIMESTAMPTZ`                         |
| `agents.agent`            | `id UUID`, `group_id UUID FK`, `name TEXT`, `role TEXT`, `pi_workspace TEXT`, `env JSONB`, `created_at` |
| `agents.resource_bundle`  | `id UUID`, `name TEXT UNIQUE`, `kind TEXT`, `spec JSONB`                                             |
| `agents.agent_resource`   | `(agent_id, resource_id)` composite PK, `mount_path TEXT`                                           |

Index `idx_agent_group` on `agents.agent(group_id)`. `name` is unique inside
its group via `UNIQUE (group_id, name)`. Resource bundle `kind` is one of
`mcp | tool | fs | secret`.

## gRPC surface

All RPCs live in `agents.v1.AgentsService`:

| RPC               | Request → Response                              | Notes                                                                  |
|-------------------|-------------------------------------------------|------------------------------------------------------------------------|
| `CreateGroup`     | `CreateGroupRequest` → `CreateGroupResponse`    | Idempotent on `name` uniqueness; emits `agents.group.created`.         |
| `ListGroups`      | `ListGroupsRequest` → `ListGroupsResponse`      | Pagination via `common.v1.Page` / `PageInfo`.                          |
| `CreateAgent`     | `CreateAgentRequest` → `CreateAgentResponse`    | Reserves `pi_workspace` on disk; emits `agents.agent.created`.         |
| `ListAgents`      | `ListAgentsRequest` → `ListAgentsResponse`      | Filtered by `group_id`.                                                |
| `GetAgent`        | `GetAgentRequest` → `GetAgentResponse`          | Used by orchestrator to resolve an `acp` node's `agentId`.             |
| `CreateResource`  | `CreateResourceRequest` → `CreateResourceResponse` | `spec` is opaque JSON validated by the resource `kind` adapter.     |
| `ListResources`   | `ListResourcesRequest` → `ListResourcesResponse` | Pagination.                                                            |
| `AttachResource`  | `AttachResourceRequest` → `AttachResourceResponse` | Emits `agents.resource.linked`; idempotent on `(agent_id, resource_id)`. |

## Events produced

Subjects from `pkg/events/subjects.go`:

- `agents.agent.created` — `SubjectAgentCreated`
- `agents.group.created` — `SubjectGroupCreated`
- `agents.resource.linked` — `SubjectResourceLinked`

Payloads are the proto messages themselves (`Agent`, `AgentGroup`,
`AgentResource`) serialised as JSON. No subject is consumed by this service.

## Config keys

Loaded via `pkg/config.Load[T]` with prefix `AGENTS`. Required envs declared in
`service.yaml`: `AGENTS_POSTGRES_DSN`. Common keys (see
[config.md](./config.md) for the full Base struct):

```yaml
service: agents
grpc_addr: ":7001"
http_addr: ":7101"
postgres:
  dsn: "postgres://awpa:awpa@localhost:5432/awpa?sslmode=disable"
  schema: agents
nats:
  url: "nats://localhost:4222"
  stream: AWPA
auth:
  mode: none      # or static-token
```

Stub overrides live under `stubs.<port>`. The agents service exposes no
outbound ports today, so the map is unused.

## Failure modes

- **Duplicate name / group name**: returns `AlreadyExists`. Unique constraint
  enforces this in DB.
- **Resource attach to missing agent or bundle**: `NotFound`; FK cascades on
  delete take care of orphans.
- **`pi_workspace` directory cannot be created**: `Internal`; agent row is
  rolled back.
- **NATS unreachable**: lifecycle event is dropped, the synchronous gRPC call
  still succeeds — consumers must tolerate this until JetStream durable
  publishing is wired (see [roadmap.md](./roadmap.md)).

## Capacity notes

- Read-mostly. A single Postgres connection pool of 8 is enough for the
  desktop-first deployment.
- `ListAgents` paginates via `common.v1.Page.cursor`; default page size 50.
- No background jobs.
