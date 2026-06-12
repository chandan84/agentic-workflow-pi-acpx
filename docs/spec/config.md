# Configuration

Layered configuration loader at `pkg/config/config.go`. Every service embeds
`config.Base` in its typed config struct and calls `config.Load[T](service,
&cfg)`.

## Layering

Later layers win:

1. **Defaults** — `config.Defaults(service)` returns a populated `Base`.
   The caller may pre-populate other fields before `Load`.
2. **Repo `config.yaml`** — read from the current working directory if
   present.
3. **User file** — path comes from `<PREFIX>_CONFIG` env. `PREFIX` is the
   uppercase-snake of the service name (e.g. `FLOW`, `ORCHESTRATOR`,
   `AGENTS`, `GATEWAY`).
4. **Environment** — `<PREFIX>_<KEY>` where the key is the flattened YAML
   path uppercased with `_` separators (e.g.
   `ORCHESTRATOR_TEMPORAL_HOST_PORT`).
5. **Flags** — applied by the service's `main` after `Load` returns.

```mermaid
flowchart LR
  D[Defaults] --> Y[config.yaml]
  Y --> F[$PREFIX_CONFIG file]
  F --> E[$PREFIX_KEY env]
  E --> X[CLI flags]
```

`config.Load` walks reflectively over string, bool, int, and struct fields.
Embedded structs use the parent prefix; non-embedded structs prepend their
YAML key. `,inline` and anonymous embedded structs share the parent prefix.

## `Base` struct

```go
type Base struct {
    Service   string            // service identity
    Env       string            // dev | staging | prod
    GRPCAddr  string            // ":7001" etc
    HTTPAddr  string            // ":7101" etc
    LogLevel  string            // debug|info|warn|error
    LogFormat string            // json|text
    Postgres  Postgres          // .dsn, .schema
    NATS      NATS              // .url, .stream
    Temporal  Temporal          // .host_port, .namespace, .task_queue
    Auth      Auth              // .mode (none|static-token), .token
    OTel      OTel              // .enabled, .endpoint
    Features  map[string]bool   // ad-hoc feature flags
    Stubs     map[string]string // port-name -> "real"|"stub"
}
```

Required-field check: `Base.Validate()` enforces non-empty `Service` and
`Postgres.DSN`.

## Defaults

`config.Defaults(service)` returns:

```yaml
service:   <provided>
env:       dev
grpc_addr: ":0"           # random for tests
http_addr: ":0"
log_level: info
log_format: json
postgres:
  dsn:    postgres://awpa:awpa@localhost:5432/awpa?sslmode=disable
  schema: <service>
nats:
  url:    nats://localhost:4222
  stream: AWPA
temporal:
  host_port: localhost:7233
  namespace: default
  task_queue: awpa-<service>
auth:     { mode: none }
otel:     { enabled: false }
features: {}
stubs:    {}
```

## Key reference by service

### agents-service (PREFIX=`AGENTS`)

| Key                       | Env                            | Default                  |
|---------------------------|--------------------------------|--------------------------|
| `service`                 | `AGENTS_SERVICE`               | `agents`                 |
| `grpc_addr`               | `AGENTS_GRPC_ADDR`             | `:7001`                  |
| `http_addr`               | `AGENTS_HTTP_ADDR`             | `:7101`                  |
| `postgres.dsn`            | `AGENTS_POSTGRES_DSN`          | local Postgres           |
| `postgres.schema`         | `AGENTS_POSTGRES_SCHEMA`       | `agents`                 |
| `nats.url`                | `AGENTS_NATS_URL`              | `nats://localhost:4222`  |
| `auth.mode`               | `AGENTS_AUTH_MODE`             | `none`                   |

### flow-service (PREFIX=`FLOW`)

| Key                       | Env                            | Default                  |
|---------------------------|--------------------------------|--------------------------|
| `service`                 | `FLOW_SERVICE`                 | `flow`                   |
| `grpc_addr`               | `FLOW_GRPC_ADDR`               | `:7002`                  |
| `postgres.dsn`            | `FLOW_POSTGRES_DSN`            | local Postgres           |
| `postgres.schema`         | `FLOW_POSTGRES_SCHEMA`         | `flows`                  |
| `stubs.author`            | `FLOW_STUBS_AUTHOR`            | `real` (or `stub`)       |

### orchestrator-service (PREFIX=`ORCHESTRATOR`)

| Key                       | Env                                   | Default                  |
|---------------------------|---------------------------------------|--------------------------|
| `grpc_addr`               | `ORCHESTRATOR_GRPC_ADDR`              | `:7003`                  |
| `postgres.dsn`            | `ORCHESTRATOR_POSTGRES_DSN`           | local Postgres           |
| `postgres.schema`         | `ORCHESTRATOR_POSTGRES_SCHEMA`        | `runs`                   |
| `temporal.host_port`      | `ORCHESTRATOR_TEMPORAL_HOST_PORT`     | `localhost:7233`         |
| `temporal.namespace`      | `ORCHESTRATOR_TEMPORAL_NAMESPACE`     | `default`                |
| `temporal.task_queue`     | `ORCHESTRATOR_TEMPORAL_TASK_QUEUE`    | `awpa-orchestrator`      |
| `stubs.runtime`           | `ORCHESTRATOR_STUBS_RUNTIME`          | `real`                   |

### gateway-service (PREFIX=`GATEWAY`)

| Key                       | Env                                | Default                  |
|---------------------------|------------------------------------|--------------------------|
| `grpc_addr`               | `GATEWAY_GRPC_ADDR`                | `:7000`                  |
| `http_addr`               | `GATEWAY_HTTP_ADDR`                | `:7100`                  |
| upstream `agents`         | `GATEWAY_AGENTS_ADDR`              | `localhost:7001`         |
| upstream `flow`           | `GATEWAY_FLOW_ADDR`                | `localhost:7002`         |
| upstream `orchestrator`   | `GATEWAY_ORCHESTRATOR_ADDR`        | `localhost:7003`         |
| `auth.mode`               | `GATEWAY_AUTH_MODE`                | `none` (boundary)        |
| `auth.token`              | `GATEWAY_AUTH_TOKEN`               | empty                    |

## Stubs map

Stubs let services run without a real downstream. Each service documents
its own stub ports under `stubs.<port>`. Today:

- `flow.stubs.author` — `real` calls the flow-author skill via pi;
  `stub` returns `examples/flow-ir/hello-review.json`.
- `orchestrator.stubs.runtime` — `real` shells to `acpx`; `stub` runs
  a canned-segment walker.

## Features map

Free-form `Features: map[string]bool`. Reserved for migration switches and
A/B rollouts. No keys are enabled by default.

## See also

- [security.md](./security.md) — `Auth` modes and secrets handling.
- [observability.md](./observability.md) — `LogLevel`, `LogFormat`, `OTel`.
