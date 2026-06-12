# Security

The platform runs on a single operator's machine in dev. Production targets a
small-team, self-hosted deployment behind a private network. The security
model reflects that: the gateway is the auth boundary; backend services
trust the loopback / private network they listen on.

## Authentication modes

`pkg/config.Auth.Mode` selects the AuthN scheme. Today:

| Mode             | Where used   | Notes                                                                              |
|------------------|--------------|------------------------------------------------------------------------------------|
| `none`           | dev default  | All RPCs accepted without credentials. Suitable only for loopback.                 |
| `static-token`   | small teams  | Single shared bearer token configured via `<PREFIX>_AUTH_TOKEN`.                   |
| `oidc` (planned) | beta+        | Standard OIDC code flow at the gateway, claims forwarded as gRPC metadata.         |

`auth.mode` is read from the embedded `config.Base.Auth`. The gateway is the
canonical place to enforce auth; backend services may also enable
`static-token` for defense in depth on multi-host deployments.

## Authenticator interface

A single `Authenticator` interface lives in `pkg/authn` and is wired into
each service's gRPC server as an interceptor. Implementations:

- `NoneAuthenticator` — always returns an empty principal.
- `StaticTokenAuthenticator` — compares the `authorization: Bearer ...`
  header against the configured token; returns `{ subject: "static" }` on
  match, `Unauthenticated` otherwise.
- `OIDCAuthenticator` (planned) — verifies a JWT against a discovered JWKS,
  returns the standard claims as principal.

The interceptor populates a `context.Context` value the handlers read. Role
checks against `IR.Permissions.Roles` are the handler's responsibility.

## Secrets handling

- Secrets only enter the process via environment variables or file paths.
  Examples: `AGENTS_POSTGRES_DSN`, `GATEWAY_AUTH_TOKEN`, OIDC client secret
  (planned). Never via gRPC requests.
- `agents.resource_bundle` rows with `kind=secret` store a *reference*
  (filesystem path or env-var name) in `spec`, never the secret value
  itself. The runtime resolver opens the reference when mounting.
- The desktop app uses Tauri's secure storage for the user's bearer token;
  it is sent only as a gRPC-Web `authorization` header.
- Postgres DSNs may contain credentials; they are not logged. The slog
  field encoder is configured to redact any value whose key matches
  `(?i)(dsn|password|token|secret)`.

## Log redaction

`pkg/obs` configures slog with a field redactor. Keys that case-insensitively
contain `password`, `secret`, `token`, `authorization`, or `dsn` have their
value replaced with `***`. The redactor is applied before the JSON encoder so
it works for both structured and message fields.

Stack traces are emitted at `error` only; the formatter strips file paths
under `/home/<user>/` to a `~` prefix.

## Threat surface

| Surface                | Mitigation today                                              | Open items                                              |
|------------------------|---------------------------------------------------------------|---------------------------------------------------------|
| Gateway HTTP/gRPC      | Bind to `127.0.0.1` by default; auth-mode `none` only on loopback. | TLS termination, OIDC, rate limiting.               |
| Inter-service gRPC     | Loopback only.                                                | mTLS for multi-host deployments.                        |
| Postgres / NATS / Temporal | Local docker compose with default creds.                  | Production hardening per-service.                       |
| `pi_workspace` dirs    | Each agent gets an isolated dir under `agents.agent.pi_workspace`. | Per-workspace seccomp / cgroups.                     |
| `action` node `fn`     | Whitelisted host functions; unknown `fn` rejected at IR validation (planned). | Today the registry is open — see roadmap.            |
| Checkpoint approvers   | Match by `approvers[]` string against principal subject.      | Group / role expansion via agents-service planned.      |

## Audit and non-repudiation

Every state-changing RPC writes an `audit.audit_event` row (layer `app`)
with the caller subject in `payload_json.actor`. Combined with the
orchestrator's Temporal history and the acpx trace bundles, the three
audit layers (see [observability.md](./observability.md)) form the
non-repudiation log.

## See also

- [config.md](./config.md) — auth keys and env mapping.
- [gateway-service.md](./gateway-service.md) — the AuthN boundary.
