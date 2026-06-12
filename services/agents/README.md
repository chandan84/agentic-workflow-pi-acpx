# agents-service

Bounded context for Agent Groups, Agents, and Resource Bundles. Owns the
`agents` Postgres schema. Publishes `agents.*` events.

See `docs/spec/agents-service.md`.

## Run locally
```sh
AGENTS_STUBS_BUS=stub AGENTS_STUBS_STORE=stub go run ./cmd/agents
```
