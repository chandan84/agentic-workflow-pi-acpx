-- +goose Up
CREATE SCHEMA IF NOT EXISTS runs;

CREATE TABLE runs.run (
  id              UUID PRIMARY KEY,
  flow_id         UUID NOT NULL,
  flow_version_id UUID NOT NULL,
  status          TEXT NOT NULL,
  current_node    TEXT NOT NULL DEFAULT '',
  input_json      JSONB NOT NULL DEFAULT '{}'::jsonb,
  temporal_run_id TEXT NOT NULL DEFAULT '',
  started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE runs.work_item (
  id           UUID PRIMARY KEY,
  run_id       UUID NOT NULL REFERENCES runs.run(id) ON DELETE CASCADE,
  node_id      TEXT NOT NULL,
  kind         TEXT NOT NULL,
  status       TEXT NOT NULL,
  payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  assignee     TEXT NOT NULL DEFAULT '',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE runs.checkpoint_request (
  id          UUID PRIMARY KEY,
  run_id      UUID NOT NULL REFERENCES runs.run(id) ON DELETE CASCADE,
  node_id     TEXT NOT NULL,
  prompt      TEXT NOT NULL,
  status      TEXT NOT NULL,
  deadline    TIMESTAMPTZ NULL,
  approver    TEXT NOT NULL DEFAULT '',
  note        TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_run_flow             ON runs.run(flow_id);
CREATE INDEX idx_work_item_run        ON runs.work_item(run_id);
CREATE INDEX idx_checkpoint_request_run ON runs.checkpoint_request(run_id);

-- +goose Down
DROP SCHEMA IF EXISTS runs CASCADE;
