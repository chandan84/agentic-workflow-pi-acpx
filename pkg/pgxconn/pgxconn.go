// Package pgxconn opens a database/sql connection to Postgres using the
// lib/pq driver. The name is historical — the actual driver under the hood
// is `pq`. Every service uses this helper so driver registration happens
// once and connection tuning is consistent.
package pgxconn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Options control the opened *sql.DB.
type Options struct {
	DSN             string
	Schema          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	PingTimeout     time.Duration
}

// Open opens a *sql.DB and sets search_path to Schema. The caller is
// responsible for closing the *sql.DB.
func Open(ctx context.Context, opt Options) (*sql.DB, error) {
	if opt.DSN == "" {
		return nil, errors.New("pgxconn: empty DSN")
	}
	db, err := sql.Open("postgres", opt.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgxconn: sql.Open: %w", err)
	}
	if opt.MaxOpenConns <= 0 {
		opt.MaxOpenConns = 25
	}
	if opt.MaxIdleConns <= 0 {
		opt.MaxIdleConns = 5
	}
	if opt.ConnMaxLifetime <= 0 {
		opt.ConnMaxLifetime = 30 * time.Minute
	}
	if opt.PingTimeout <= 0 {
		opt.PingTimeout = 3 * time.Second
	}
	db.SetMaxOpenConns(opt.MaxOpenConns)
	db.SetMaxIdleConns(opt.MaxIdleConns)
	db.SetConnMaxLifetime(opt.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, opt.PingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pgxconn: ping: %w", err)
	}
	if opt.Schema != "" {
		if _, err := db.ExecContext(pingCtx, fmt.Sprintf("SET search_path TO %q,public", opt.Schema)); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("pgxconn: set search_path: %w", err)
		}
	}
	return db, nil
}
