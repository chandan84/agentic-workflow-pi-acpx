// Package store is the Postgres adapter.
package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/ports"
)

// Postgres backs the Store with database/sql.
type Postgres struct{ DB *sql.DB }

// New wraps a *sql.DB.
func New(db *sql.DB) *Postgres { return &Postgres{DB: db} }

// CreateFlow inserts a flow.
func (p *Postgres) CreateFlow(ctx context.Context, f ports.Flow) (ports.Flow, error) {
	if p.DB == nil {
		return ports.Flow{}, errors.New("nil DB")
	}
	f.CreatedAt = time.Now()
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO flows.flow(id,name,description,created_at) VALUES($1,$2,$3,$4)`,
		f.ID, f.Name, f.Description, f.CreatedAt)
	return f, err
}

// GetFlow loads a flow and its versions.
func (p *Postgres) GetFlow(ctx context.Context, id string) (ports.Flow, []ports.FlowVersion, error) {
	var f ports.Flow
	row := p.DB.QueryRowContext(ctx, `SELECT id,name,description,created_at FROM flows.flow WHERE id=$1`, id)
	if err := row.Scan(&f.ID, &f.Name, &f.Description, &f.CreatedAt); err != nil {
		return f, nil, err
	}
	rows, err := p.DB.QueryContext(ctx,
		`SELECT id,flow_id,version,ir_json::text,flow_ts,created_at FROM flows.flow_version WHERE flow_id=$1 ORDER BY version`, id)
	if err != nil {
		return f, nil, err
	}
	defer rows.Close()
	var out []ports.FlowVersion
	for rows.Next() {
		var v ports.FlowVersion
		if err := rows.Scan(&v.ID, &v.FlowID, &v.Version, &v.IRJSON, &v.FlowTS, &v.CreatedAt); err != nil {
			return f, nil, err
		}
		out = append(out, v)
	}
	return f, out, rows.Err()
}

// ListFlows returns flows.
func (p *Postgres) ListFlows(ctx context.Context, limit int, _ string) ([]ports.Flow, string, error) {
	if limit == 0 {
		limit = 100
	}
	rows, err := p.DB.QueryContext(ctx, `SELECT id,name,description,created_at FROM flows.flow ORDER BY name LIMIT $1`, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var out []ports.Flow
	for rows.Next() {
		var f ports.Flow
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.CreatedAt); err != nil {
			return nil, "", err
		}
		out = append(out, f)
	}
	return out, "", rows.Err()
}

// PutVersion appends a flow version.
func (p *Postgres) PutVersion(ctx context.Context, v ports.FlowVersion) (ports.FlowVersion, error) {
	v.CreatedAt = time.Now()
	row := p.DB.QueryRowContext(ctx,
		`INSERT INTO flows.flow_version(id,flow_id,version,ir_json,flow_ts,created_at)
		 VALUES($1,$2, COALESCE((SELECT MAX(version)+1 FROM flows.flow_version WHERE flow_id=$2),1), $3::jsonb, $4, $5)
		 RETURNING version`,
		v.ID, v.FlowID, v.IRJSON, v.FlowTS, v.CreatedAt)
	if err := row.Scan(&v.Version); err != nil {
		return v, err
	}
	return v, nil
}

// Ping is the readiness probe.
func (p *Postgres) Ping(ctx context.Context) error { return p.DB.PingContext(ctx) }
