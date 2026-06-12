// Package ports declares orchestrator-service ports.
package ports

import (
	"context"
	"time"
)

// Run is the runs.run row.
type Run struct {
	ID, FlowID, FlowVersionID, Status, CurrentNode string
	StartedAt, UpdatedAt                           time.Time
}

// CheckpointRequest is the runs.checkpoint_request row.
type CheckpointRequest struct {
	ID, RunID, NodeID, Prompt, Status, Approver, Note string
	Deadline                                          time.Time
}

// WorkItem is the runs.work_item row.
type WorkItem struct {
	ID, RunID, NodeID, Kind, Status, PayloadJSON, Assignee string
}

// Store persists runs, work items, checkpoint requests.
type Store interface {
	CreateRun(ctx context.Context, r Run) (Run, error)
	UpdateRun(ctx context.Context, id, status, node string) error
	GetRun(ctx context.Context, id string) (Run, error)
	ListRuns(ctx context.Context, flowID string, limit int, cursor string) ([]Run, string, error)

	UpsertCheckpoint(ctx context.Context, c CheckpointRequest) (CheckpointRequest, error)
	GetCheckpoint(ctx context.Context, id string) (CheckpointRequest, error)

	InsertWorkItem(ctx context.Context, w WorkItem) error
	Ping(ctx context.Context) error
}
