// Package ports declares the outbound dependencies of the agents service.
// Real adapters live under internal/store; stub adapters live under internal/stub.
package ports

import (
	"context"
	"time"
)

// Group is the agents.agents_group row.
type Group struct {
	ID, Name, Description string
	CreatedAt             time.Time
}

// Agent is the agents.agent row.
type Agent struct {
	ID, GroupID, Name, Role, PiWorkspace string
	Env                                  map[string]string
	CreatedAt                            time.Time
}

// Resource is the agents.resource_bundle row.
type Resource struct {
	ID, Name, Kind, Spec string
}

// AgentResource links agents to resource bundles with a mount path.
type AgentResource struct {
	AgentID, ResourceID, MountPath string
}

// Store is the persistence port.
type Store interface {
	CreateGroup(ctx context.Context, g Group) (Group, error)
	ListGroups(ctx context.Context, limit int, cursor string) ([]Group, string, error)

	CreateAgent(ctx context.Context, a Agent) (Agent, error)
	GetAgent(ctx context.Context, id string) (Agent, error)
	ListAgents(ctx context.Context, groupID string, limit int, cursor string) ([]Agent, string, error)

	CreateResource(ctx context.Context, r Resource) (Resource, error)
	ListResources(ctx context.Context, limit int, cursor string) ([]Resource, string, error)
	AttachResource(ctx context.Context, link AgentResource) (AgentResource, error)

	Ping(ctx context.Context) error
}
