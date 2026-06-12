// Package app holds the application-layer service object that wires ports
// together. Business logic lives here; the gRPC layer is a thin translator.
package app

import (
	"context"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/ports"
	"github.com/google/uuid"
)

// Service is the application object.
type Service struct {
	Store ports.Store
	Bus   events.Bus
}

// New builds a Service from its ports.
func New(s ports.Store, b events.Bus) *Service { return &Service{Store: s, Bus: b} }

// CreateGroup creates a group and publishes an event.
func (s *Service) CreateGroup(ctx context.Context, name, description string) (ports.Group, error) {
	g, err := s.Store.CreateGroup(ctx, ports.Group{
		ID: uuid.NewString(), Name: name, Description: description,
	})
	if err != nil {
		return g, err
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectGroupCreated, g)
	}
	return g, nil
}

// CreateAgent creates an agent and publishes an event.
func (s *Service) CreateAgent(ctx context.Context, groupID, name, role string, env map[string]string) (ports.Agent, error) {
	a := ports.Agent{
		ID: uuid.NewString(), GroupID: groupID, Name: name, Role: role,
		PiWorkspace: "/var/lib/awpa/agents/" + name + "/.pi",
		Env:         env,
	}
	a, err := s.Store.CreateAgent(ctx, a)
	if err != nil {
		return a, err
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectAgentCreated, a)
	}
	return a, nil
}

// CreateResource creates a resource bundle.
func (s *Service) CreateResource(ctx context.Context, name, kind, spec string) (ports.Resource, error) {
	r := ports.Resource{ID: uuid.NewString(), Name: name, Kind: kind, Spec: spec}
	return s.Store.CreateResource(ctx, r)
}

// AttachResource links a resource to an agent and publishes an event.
func (s *Service) AttachResource(ctx context.Context, agentID, resourceID, mount string) (ports.AgentResource, error) {
	l, err := s.Store.AttachResource(ctx, ports.AgentResource{
		AgentID: agentID, ResourceID: resourceID, MountPath: mount,
	})
	if err != nil {
		return l, err
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectResourceLinked, l)
	}
	return l, nil
}
