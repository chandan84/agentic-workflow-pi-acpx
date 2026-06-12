# gateway-service

Single gRPC aggregator exposed to the desktop app. Proxies agents, flow, and
execution domain RPCs to the upstream services and serves the unified audit
timeline. Subscribes to NATS for live UI fan-out.

See `docs/spec/gateway-service.md`.
