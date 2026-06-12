# Data model

Four Postgres schemas, one per bounded context. All migrations live under
`db/migrations/<schema>/` and are applied by goose. Each schema is owned by
exactly one service.

## Schema ownership

| Schema   | Owning service       | Migrations directory             |
|----------|----------------------|----------------------------------|
| `agents` | agents-service       | `db/migrations/agents/`          |
| `flows`  | flow-service         | `db/migrations/flows/`           |
| `runs`   | orchestrator-service | `db/migrations/runs/`            |
| `audit`  | gateway (writers from all services) | `db/migrations/audit/` |

Cross-schema foreign keys are forbidden. References (e.g.
`runs.run.flow_id`) are by id only, with no DB-level constraint. Joins live
in the gateway aggregator, not in SQL.

## ER diagram

```mermaid
erDiagram
  AGENTS_GROUP ||--o{ AGENT : "group_id"
  AGENT ||--o{ AGENT_RESOURCE : "agent_id"
  RESOURCE_BUNDLE ||--o{ AGENT_RESOURCE : "resource_id"

  FLOW ||--o{ FLOW_VERSION : "flow_id"

  RUN ||--o{ WORK_ITEM : "run_id"
  RUN ||--o{ CHECKPOINT_REQUEST : "run_id"

  AGENTS_GROUP {
    UUID id PK
    TEXT name
    TEXT description
    TIMESTAMPTZ created_at
  }
  AGENT {
    UUID id PK
    UUID group_id FK
    TEXT name
    TEXT role
    TEXT pi_workspace
    JSONB env
    TIMESTAMPTZ created_at
  }
  RESOURCE_BUNDLE {
    UUID id PK
    TEXT name
    TEXT kind "mcp tool fs secret"
    JSONB spec
  }
  AGENT_RESOURCE {
    UUID agent_id FK
    UUID resource_id FK
    TEXT mount_path
  }

  FLOW {
    UUID id PK
    TEXT name
    TEXT description
    TIMESTAMPTZ created_at
  }
  FLOW_VERSION {
    UUID id PK
    UUID flow_id FK
    INT version
    JSONB ir_json
    TEXT flow_ts
    TIMESTAMPTZ created_at
  }

  RUN {
    UUID id PK
    UUID flow_id
    UUID flow_version_id
    TEXT status
    TEXT current_node
    JSONB input_json
    TEXT temporal_run_id
    TIMESTAMPTZ started_at
    TIMESTAMPTZ updated_at
  }
  WORK_ITEM {
    UUID id PK
    UUID run_id FK
    TEXT node_id
    TEXT kind "task approval review"
    TEXT status "open claimed done skipped"
    JSONB payload_json
    TEXT assignee
    TIMESTAMPTZ created_at
  }
  CHECKPOINT_REQUEST {
    UUID id PK
    UUID run_id FK
    TEXT node_id
    TEXT prompt
    TEXT status "awaiting approved rejected timeout"
    TIMESTAMPTZ deadline
    TEXT approver
    TEXT note
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  AUDIT_EVENT {
    UUID id PK
    UUID run_id
    TEXT node_id
    TEXT layer "temporal acpx app"
    TEXT kind
    JSONB payload_json
    TIMESTAMPTZ at
  }
```

## Notes

- `flows.flow_version.flow_ts` mirrors `flowir.Codegen`'s output. It is
  stored to make the desktop editor view round-trippable without re-running
  codegen.
- `runs.run.temporal_run_id` links a row to its Temporal workflow execution
  for layer-1 audit lookups.
- `runs.checkpoint_request.id` is created by the orchestrator as
  `"<run_id>:<node_id>"` (see `workflow.go`); the column stores that exact
  string when the checkpoint is materialised.
- `audit.audit_event` is intentionally denormalised; indexed by `(run_id,
  at)`. It survives indefinitely (no `MaxAge`).

## Cross-references

- [agents-service.md](./agents-service.md#owned-schema)
- [flow-service.md](./flow-service.md#owned-schema)
- [orchestrator-service.md](./orchestrator-service.md#owned-schema)
- [observability.md](./observability.md#the-three-audit-layers)
