-- +goose Up
CREATE SCHEMA IF NOT EXISTS agents;

CREATE TABLE agents.agents_group (
  id          UUID PRIMARY KEY,
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (name)
);

CREATE TABLE agents.agent (
  id           UUID PRIMARY KEY,
  group_id     UUID NOT NULL REFERENCES agents.agents_group(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  role         TEXT NOT NULL,
  pi_workspace TEXT NOT NULL,
  env          JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (group_id, name)
);

CREATE TABLE agents.resource_bundle (
  id   UUID PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  kind TEXT NOT NULL,
  spec JSONB NOT NULL
);

CREATE TABLE agents.agent_resource (
  agent_id    UUID NOT NULL REFERENCES agents.agent(id) ON DELETE CASCADE,
  resource_id UUID NOT NULL REFERENCES agents.resource_bundle(id) ON DELETE CASCADE,
  mount_path  TEXT NOT NULL,
  PRIMARY KEY (agent_id, resource_id)
);

CREATE INDEX idx_agent_group ON agents.agent(group_id);

-- +goose Down
DROP SCHEMA IF EXISTS agents CASCADE;
