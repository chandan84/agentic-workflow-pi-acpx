// Package store is the Postgres-backed Store adapter.
//
// This iteration ships a real adapter that opens a *sql.DB and implements
// Ping; CRUD methods are wired to SQL but kept short — the operations are
// intentionally straight-line so the patterns are obvious.
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/ports"
)

// Postgres is the database/sql-backed Store.
type Postgres struct {
	DB *sql.DB
}

// New opens a *sql.DB for the given DSN. The driver must be registered by the caller.
func New(db *sql.DB) *Postgres { return &Postgres{DB: db} }

// CreateGroup inserts an agents.agents_group row.
func (p *Postgres) CreateGroup(ctx context.Context, g ports.Group) (ports.Group, error) {
	if p.DB == nil {
		return ports.Group{}, errors.New("agents.store: nil DB")
	}
	g.CreatedAt = time.Now()
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO agents.agents_group(id,name,description,created_at) VALUES($1,$2,$3,$4)`,
		g.ID, g.Name, g.Description, g.CreatedAt)
	return g, err
}

// ListGroups reads agents.agents_group rows.
func (p *Postgres) ListGroups(ctx context.Context, limit int, _ string) ([]ports.Group, string, error) {
	if limit == 0 {
		limit = 100
	}
	rows, err := p.DB.QueryContext(ctx, `SELECT id,name,description,created_at FROM agents.agents_group ORDER BY name LIMIT $1`, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var out []ports.Group
	for rows.Next() {
		var g ports.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt); err != nil {
			return nil, "", err
		}
		out = append(out, g)
	}
	return out, "", rows.Err()
}

// CreateAgent inserts an agents.agent row.
func (p *Postgres) CreateAgent(ctx context.Context, a ports.Agent) (ports.Agent, error) {
	a.CreatedAt = time.Now()
	envJSON, _ := json.Marshal(a.Env)
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO agents.agent(id,group_id,name,role,pi_workspace,env,created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`,
		a.ID, a.GroupID, a.Name, a.Role, a.PiWorkspace, envJSON, a.CreatedAt)
	return a, err
}

// GetAgent fetches one agent.
func (p *Postgres) GetAgent(ctx context.Context, id string) (ports.Agent, error) {
	row := p.DB.QueryRowContext(ctx,
		`SELECT id,group_id,name,role,pi_workspace,env,created_at FROM agents.agent WHERE id=$1`, id)
	var a ports.Agent
	var envJSON []byte
	if err := row.Scan(&a.ID, &a.GroupID, &a.Name, &a.Role, &a.PiWorkspace, &envJSON, &a.CreatedAt); err != nil {
		return ports.Agent{}, err
	}
	_ = json.Unmarshal(envJSON, &a.Env)
	return a, nil
}

// ListAgents reads agents for a group.
func (p *Postgres) ListAgents(ctx context.Context, group string, limit int, _ string) ([]ports.Agent, string, error) {
	if limit == 0 {
		limit = 100
	}
	rows, err := p.DB.QueryContext(ctx,
		`SELECT id,group_id,name,role,pi_workspace,env,created_at FROM agents.agent WHERE ($1='' OR group_id::text=$1) ORDER BY name LIMIT $2`,
		group, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var out []ports.Agent
	for rows.Next() {
		var a ports.Agent
		var envJSON []byte
		if err := rows.Scan(&a.ID, &a.GroupID, &a.Name, &a.Role, &a.PiWorkspace, &envJSON, &a.CreatedAt); err != nil {
			return nil, "", err
		}
		_ = json.Unmarshal(envJSON, &a.Env)
		out = append(out, a)
	}
	return out, "", rows.Err()
}

// CreateResource inserts a resource bundle.
func (p *Postgres) CreateResource(ctx context.Context, r ports.Resource) (ports.Resource, error) {
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO agents.resource_bundle(id,name,kind,spec) VALUES($1,$2,$3,$4)`,
		r.ID, r.Name, r.Kind, r.Spec)
	return r, err
}

// ListResources reads resource bundles.
func (p *Postgres) ListResources(ctx context.Context, limit int, _ string) ([]ports.Resource, string, error) {
	if limit == 0 {
		limit = 100
	}
	rows, err := p.DB.QueryContext(ctx, `SELECT id,name,kind,spec::text FROM agents.resource_bundle ORDER BY name LIMIT $1`, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var out []ports.Resource
	for rows.Next() {
		var r ports.Resource
		if err := rows.Scan(&r.ID, &r.Name, &r.Kind, &r.Spec); err != nil {
			return nil, "", err
		}
		out = append(out, r)
	}
	return out, "", rows.Err()
}

// AttachResource links an agent to a resource bundle.
func (p *Postgres) AttachResource(ctx context.Context, l ports.AgentResource) (ports.AgentResource, error) {
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO agents.agent_resource(agent_id,resource_id,mount_path) VALUES($1,$2,$3)
		 ON CONFLICT (agent_id,resource_id) DO UPDATE SET mount_path=EXCLUDED.mount_path`,
		l.AgentID, l.ResourceID, l.MountPath)
	return l, err
}

// Ping is the readiness probe.
func (p *Postgres) Ping(ctx context.Context) error { return p.DB.PingContext(ctx) }
