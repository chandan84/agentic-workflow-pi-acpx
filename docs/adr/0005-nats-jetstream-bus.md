# ADR 0005: NATS JetStream as the event bus

Status: Accepted · Date: 2026-06-11

## Context

Services are siloed (no cross-service Go imports, no shared tables), but the
UI needs live updates and services need to react to each other's state
changes. Candidates: Postgres LISTEN/NOTIFY (no durability, couples services
to one DB), Kafka (operationally heavy for a desktop-first product), NATS.

## Decision

NATS is the bus. Durable subjects (run lifecycle, checkpoint requests, audit)
use JetStream streams; live UI fan-out uses ephemeral core-NATS
subscriptions. Every subject is declared as a constant in `pkg/events` and
catalogued in `docs/spec/events.md` with payload, producer, consumers, and
retention — adding a subject without updating both is a CI-reviewable smell.

## Consequences

- Gateway streaming RPCs are thin: subscribe, filter, forward.
- Consumers must tolerate at-least-once delivery; payloads carry event IDs
  for dedup.
- One more process in the dev stack (`deploy/compose.yaml`), negligible
  footprint.
- Events are notifications, not state transfer: the source of truth stays in
  each service's Postgres schema, and consumers re-fetch via gRPC when they
  need authoritative data.
