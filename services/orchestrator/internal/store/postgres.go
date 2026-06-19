// Package store is the Postgres adapter for the orchestrator's runs schema.
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/ports"
)

// Postgres is the database/sql-backed Store.
type Postgres struct{ DB *sql.DB }

// New wraps a *sql.DB.
func New(db *sql.DB) *Postgres { return &Postgres{DB: db} }

// CreateRun inserts a runs.run row.
func (p *Postgres) CreateRun(ctx context.Context, r ports.Run) (ports.Run, error) {
	if p.DB == nil {
		return ports.Run{}, errors.New("orchestrator.store: nil DB")
	}
	now := time.Now()
	r.StartedAt = now
	r.UpdatedAt = now
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO runs.run(id,flow_id,flow_version_id,status,current_node,started_at,updated_at)
		 VALUES($1,$2,$3,$4,$5,$6,$7)`,
		r.ID, r.FlowID, r.FlowVersionID, r.Status, r.CurrentNode, r.StartedAt, r.UpdatedAt)
	return r, err
}

// UpdateRun mutates status and/or current_node.
func (p *Postgres) UpdateRun(ctx context.Context, id, status, node string) error {
	_, err := p.DB.ExecContext(ctx,
		`UPDATE runs.run
		   SET status = COALESCE(NULLIF($2,''), status),
		       current_node = COALESCE(NULLIF($3,''), current_node),
		       updated_at = now()
		 WHERE id=$1`, id, status, node)
	return err
}

// GetRun fetches a run.
func (p *Postgres) GetRun(ctx context.Context, id string) (ports.Run, error) {
	row := p.DB.QueryRowContext(ctx,
		`SELECT id,flow_id,flow_version_id,status,current_node,started_at,updated_at
		   FROM runs.run WHERE id=$1`, id)
	var r ports.Run
	err := row.Scan(&r.ID, &r.FlowID, &r.FlowVersionID, &r.Status, &r.CurrentNode, &r.StartedAt, &r.UpdatedAt)
	return r, err
}

// ListRuns returns runs, optionally filtered by flow id.
func (p *Postgres) ListRuns(ctx context.Context, flowID string, limit int, _ string) ([]ports.Run, string, error) {
	if limit == 0 {
		limit = 100
	}
	rows, err := p.DB.QueryContext(ctx,
		`SELECT id,flow_id,flow_version_id,status,current_node,started_at,updated_at
		   FROM runs.run
		  WHERE ($1='' OR flow_id::text=$1)
		  ORDER BY started_at DESC
		  LIMIT $2`, flowID, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var out []ports.Run
	for rows.Next() {
		var r ports.Run
		if err := rows.Scan(&r.ID, &r.FlowID, &r.FlowVersionID, &r.Status, &r.CurrentNode, &r.StartedAt, &r.UpdatedAt); err != nil {
			return nil, "", err
		}
		out = append(out, r)
	}
	return out, "", rows.Err()
}

// UpsertCheckpoint inserts or updates a runs.checkpoint_request row.
func (p *Postgres) UpsertCheckpoint(ctx context.Context, c ports.CheckpointRequest) (ports.CheckpointRequest, error) {
	if c.Deadline.IsZero() {
		c.Deadline = time.Now().Add(24 * time.Hour)
	}
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO runs.checkpoint_request(id,run_id,node_id,prompt,status,deadline,approver,note,updated_at)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,now())
		 ON CONFLICT (id) DO UPDATE SET
		   status   = EXCLUDED.status,
		   approver = EXCLUDED.approver,
		   note     = EXCLUDED.note,
		   updated_at = now()`,
		c.ID, c.RunID, c.NodeID, c.Prompt, c.Status, c.Deadline, c.Approver, c.Note)
	return c, err
}

// GetCheckpoint reads a checkpoint request.
func (p *Postgres) GetCheckpoint(ctx context.Context, id string) (ports.CheckpointRequest, error) {
	row := p.DB.QueryRowContext(ctx,
		`SELECT id,run_id,node_id,prompt,status,deadline,approver,note
		   FROM runs.checkpoint_request WHERE id=$1`, id)
	var c ports.CheckpointRequest
	var deadline sql.NullTime
	if err := row.Scan(&c.ID, &c.RunID, &c.NodeID, &c.Prompt, &c.Status, &deadline, &c.Approver, &c.Note); err != nil {
		return c, err
	}
	if deadline.Valid {
		c.Deadline = deadline.Time
	}
	return c, nil
}

// InsertWorkItem records a runs.work_item row.
func (p *Postgres) InsertWorkItem(ctx context.Context, w ports.WorkItem) error {
	payload := w.PayloadJSON
	if payload == "" {
		payload = "{}"
	}
	// Sanity check that payload is valid JSON before INSERT.
	var any any
	if err := json.Unmarshal([]byte(payload), &any); err != nil {
		return err
	}
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO runs.work_item(id,run_id,node_id,kind,status,payload_json,assignee)
		 VALUES($1,$2,$3,$4,$5,$6::jsonb,$7)`,
		w.ID, w.RunID, w.NodeID, w.Kind, w.Status, payload, w.Assignee)
	return err
}

// Ping is the readiness probe.
func (p *Postgres) Ping(ctx context.Context) error { return p.DB.PingContext(ctx) }
