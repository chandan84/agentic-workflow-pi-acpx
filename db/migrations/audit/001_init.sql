-- +goose Up
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE audit.audit_event (
  id           UUID PRIMARY KEY,
  run_id       UUID NOT NULL,
  node_id      TEXT NOT NULL DEFAULT '',
  layer        TEXT NOT NULL,    -- temporal|acpx|app
  kind         TEXT NOT NULL,
  payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_run_at ON audit.audit_event(run_id, at);

-- +goose Down
DROP SCHEMA IF EXISTS audit CASCADE;
