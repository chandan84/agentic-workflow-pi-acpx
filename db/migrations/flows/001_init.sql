-- +goose Up
CREATE SCHEMA IF NOT EXISTS flows;

CREATE TABLE flows.flow (
  id          UUID PRIMARY KEY,
  name        TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE flows.flow_version (
  id          UUID PRIMARY KEY,
  flow_id     UUID NOT NULL REFERENCES flows.flow(id) ON DELETE CASCADE,
  version     INT  NOT NULL,
  ir_json     JSONB NOT NULL,
  flow_ts     TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (flow_id, version)
);

CREATE INDEX idx_flow_version_flow ON flows.flow_version(flow_id);

-- +goose Down
DROP SCHEMA IF EXISTS flows CASCADE;
