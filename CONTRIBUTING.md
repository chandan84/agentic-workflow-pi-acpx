# Contributing

## Branch model
- `main` is protected. Feature work on `claude/<topic>-<slug>` branches.
- Open a PR; CI must pass; one reviewer.

## Code style
- Go: `gofmt -s` + `goimports -local github.com/chandan84/agentic-workflow-pi-acpx`. `golangci-lint run` clean.
- TypeScript: Prettier + ESLint defaults.
- Protos: `buf lint` + `buf breaking` against `main`.

## Adding a service
1. Create `services/<name>/` with `service.yaml`, `config.example.yaml`, `cmd/<name>/main.go`, `internal/...`.
2. Add the module to `go.work`.
3. Define protos under `proto/<domain>/v1/`.
4. Add migrations under `db/migrations/<schema>/`.
5. Document in `docs/spec/<name>-service.md` and update `docs/spec/overview.md`.
6. Add an ADR if this introduces an architectural choice.

## ADRs
One ADR per material decision. Use the template at `docs/adr/0000-template.md`.

## Tests
Every package ships at least one test. End-to-end demo lives in `tests/e2e/`.
