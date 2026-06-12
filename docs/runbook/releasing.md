# Runbook: cutting a release

## Versioning

One repo-wide semver tag (`vX.Y.Z`) covers services and the desktop app;
individual service images are tagged with the same version. Contracts in
`/proto` are additionally guarded by `buf breaking` — a breaking proto change
requires a new `vN` package directory, never an in-place edit.

## Steps

1. **Green main.** CI (`.github/workflows/ci.yml`) must be green: Go
   build/test/lint, buf lint + breaking, web build, tauri check.
2. **Migration review.** Diff `db/migrations/` since the last tag. Confirm
   every new migration is backward-compatible with the previous service
   version (expand-then-contract: additive first, destructive only one
   release after consumers stopped using the old shape).
3. **Tag.** `git tag vX.Y.Z && git push origin vX.Y.Z`.
4. **Build artifacts.**
   - service images: `docker build` per `services/*/Dockerfile`, tag and push
   - desktop bundles: `npm --prefix apps/desktop/web run tauri:build`
5. **Apply migrations** to the target environment with
   `scripts/migrate-all.sh` (set `PG_DSN`), **before** rolling service
   binaries.
6. **Roll services**, orchestrator last — its Temporal workers must be the
   newest code so in-flight Workflow histories replay deterministically.
7. **Smoke check**: `/healthz` and `/readyz` per service, then one end-to-end
   run of `examples/flow-ir/hello-review.json` through the gateway.

## Rollback

Re-deploy the previous image tag. Do not roll back migrations; the
expand-then-contract rule above is what makes binary rollback safe.
