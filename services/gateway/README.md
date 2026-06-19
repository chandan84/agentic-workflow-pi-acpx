# gateway-service

Single endpoint exposed to the desktop app. Proxies agents, flow, and
execution domain RPCs to the upstream services and serves the unified audit
timeline directly. Subscribes to NATS for live UI fan-out.

See `docs/spec/gateway-service.md`.

## Wire protocol

The gateway listens for **gRPC** today. The desktop app uses Connect-Web in
the browser shell — to make those clients reach the gateway you have two
production options:

1. Add Connect-Go handlers alongside the existing gRPC handlers. Run
   `buf generate` with `protoc-gen-connect-go` (not yet on PATH in CI), then
   register Connect handlers on a shared HTTP listener. The web app's
   generated TS clients (`apps/desktop/web/src/gen/`) speak Connect natively
   — no code change there.
2. Bundle a gRPC-Web proxy (e.g. `improbable-eng/grpc-web/go/grpcweb`) that
   wraps the existing `*grpc.Server`. Single dependency, no extra codegen.

This iteration ships the **gRPC backbone end-to-end** (integration test in
`tests/e2e/integration_test.go` drives the full create-flow → run lifecycle
through the gateway). The web app's Groups page demonstrates the
gateway-client wiring; until option 1 or 2 lands it falls back to its
mock data when the Connect call fails.
