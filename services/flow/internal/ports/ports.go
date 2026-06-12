// Package ports declares outbound dependencies of the flow service.
package ports

import (
	"context"
	"time"
)

// Flow is the flows.flow row.
type Flow struct {
	ID, Name, Description string
	CreatedAt             time.Time
}

// FlowVersion is the flows.flow_version row.
type FlowVersion struct {
	ID, FlowID  string
	Version     int
	IRJSON      string
	FlowTS      string
	CreatedAt   time.Time
}

// Store is the persistence port.
type Store interface {
	CreateFlow(ctx context.Context, f Flow) (Flow, error)
	GetFlow(ctx context.Context, id string) (Flow, []FlowVersion, error)
	ListFlows(ctx context.Context, limit int, cursor string) ([]Flow, string, error)
	PutVersion(ctx context.Context, v FlowVersion) (FlowVersion, error)
	Ping(ctx context.Context) error
}
