# Architecture Decision Records

Short, immutable records of the decisions that shape this codebase. New
decisions get a new numbered file; superseding an old decision links both ways.

| # | Title | Status |
|---|---|---|
| [0001](0001-acpx-as-engine.md) | acpx as the flow execution engine | Accepted |
| [0002](0002-go-backend-rust-shell.md) | Go backend, Rust only for the Tauri shell | Accepted |
| [0003](0003-flow-ir-canonical.md) | Versioned Flow IR as the canonical flow format | Accepted |
| [0004](0004-checkpoint-segmentation.md) | Segment runs at checkpoint nodes under Temporal | Accepted |
| [0005](0005-nats-jetstream-bus.md) | NATS JetStream as the event bus | Accepted |
| [0006](0006-schema-per-service.md) | One Postgres, schema-per-service | Accepted |
| [0007](0007-config-layering.md) | Layered configuration loading | Accepted |
| [0008](0008-ports-and-adapters.md) | Ports-and-adapters per service | Accepted |
| [0009](0009-buf-contract-first.md) | Contract-first gRPC with buf; generated Go committed | Accepted |
| [0010](0010-goose-migrations.md) | goose for schema migrations | Accepted |
