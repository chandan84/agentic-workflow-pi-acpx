# ADR 0007: Layered configuration loading

Status: Accepted · Date: 2026-06-11

## Context

Every service needs the same configurability story: sane defaults, a checked-in
example, machine-local overrides, environment injection for containers, and
flags for one-off runs — with typo-level validation at startup, not at first
use.

## Decision

`pkg/config` implements one loader used by all services, applying layers in
increasing precedence:

1. built-in defaults (Go struct tags)
2. repo `config.yaml` next to the binary
3. user file at `$<SVC>_CONFIG_FILE`
4. `<SVC>_`-prefixed environment variables
5. command-line flags

Config structs are strongly typed; validation runs at startup and reports the
exact key path of every error. Feature flags live in a dedicated `features:`
map. Each service ships a commented `config.example.yaml`.

## Consequences

- "How do I change X?" has one answer regardless of service.
- New keys must be added to the struct, the example file, and
  `docs/spec/config.md` — drift between them is reviewable.
- Hot-reload is stubbed (a hook on the loader); enabling it later requires no
  call-site changes.
