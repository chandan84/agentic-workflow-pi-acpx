// Package grpcsrv adapts the application Service to the generated gRPC server.
package grpcsrv

import (
	"context"

	agentsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/agents/v1"
	commonv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/common/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/ports"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements agents.v1.AgentsServiceServer.
type Server struct {
	agentsv1.UnimplementedAgentsServiceServer
	Svc *app.Service
}

// New wires a Server.
func New(s *app.Service) *Server { return &Server{Svc: s} }

// CreateGroup creates a group.
func (s *Server) CreateGroup(ctx context.Context, in *agentsv1.CreateGroupRequest) (*agentsv1.CreateGroupResponse, error) {
	g, err := s.Svc.CreateGroup(ctx, in.GetName(), in.GetDescription())
	if err != nil {
		return nil, err
	}
	return &agentsv1.CreateGroupResponse{Group: toProtoGroup(g)}, nil
}

// ListGroups lists groups.
func (s *Server) ListGroups(ctx context.Context, in *agentsv1.ListGroupsRequest) (*agentsv1.ListGroupsResponse, error) {
	groups, next, err := s.Svc.Store.ListGroups(ctx, pageLimit(in.GetPage()), pageCursor(in.GetPage()))
	if err != nil {
		return nil, err
	}
	out := &agentsv1.ListGroupsResponse{PageInfo: &commonv1.PageInfo{NextCursor: next}}
	for _, g := range groups {
		out.Groups = append(out.Groups, toProtoGroup(g))
	}
	return out, nil
}

// CreateAgent creates an agent.
func (s *Server) CreateAgent(ctx context.Context, in *agentsv1.CreateAgentRequest) (*agentsv1.CreateAgentResponse, error) {
	a, err := s.Svc.CreateAgent(ctx, in.GetGroupId(), in.GetName(), in.GetRole(), in.GetEnv())
	if err != nil {
		return nil, err
	}
	return &agentsv1.CreateAgentResponse{Agent: toProtoAgent(a)}, nil
}

// ListAgents lists agents in a group.
func (s *Server) ListAgents(ctx context.Context, in *agentsv1.ListAgentsRequest) (*agentsv1.ListAgentsResponse, error) {
	rows, next, err := s.Svc.Store.ListAgents(ctx, in.GetGroupId(), pageLimit(in.GetPage()), pageCursor(in.GetPage()))
	if err != nil {
		return nil, err
	}
	out := &agentsv1.ListAgentsResponse{PageInfo: &commonv1.PageInfo{NextCursor: next}}
	for _, a := range rows {
		out.Agents = append(out.Agents, toProtoAgent(a))
	}
	return out, nil
}

// GetAgent fetches one agent.
func (s *Server) GetAgent(ctx context.Context, in *agentsv1.GetAgentRequest) (*agentsv1.GetAgentResponse, error) {
	a, err := s.Svc.Store.GetAgent(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	return &agentsv1.GetAgentResponse{Agent: toProtoAgent(a)}, nil
}

// CreateResource creates a resource bundle.
func (s *Server) CreateResource(ctx context.Context, in *agentsv1.CreateResourceRequest) (*agentsv1.CreateResourceResponse, error) {
	r, err := s.Svc.CreateResource(ctx, in.GetName(), in.GetKind(), in.GetSpec())
	if err != nil {
		return nil, err
	}
	return &agentsv1.CreateResourceResponse{Resource: toProtoResource(r)}, nil
}

// ListResources lists resource bundles.
func (s *Server) ListResources(ctx context.Context, in *agentsv1.ListResourcesRequest) (*agentsv1.ListResourcesResponse, error) {
	rows, next, err := s.Svc.Store.ListResources(ctx, pageLimit(in.GetPage()), pageCursor(in.GetPage()))
	if err != nil {
		return nil, err
	}
	out := &agentsv1.ListResourcesResponse{PageInfo: &commonv1.PageInfo{NextCursor: next}}
	for _, r := range rows {
		out.Resources = append(out.Resources, toProtoResource(r))
	}
	return out, nil
}

// AttachResource links an agent to a resource bundle.
func (s *Server) AttachResource(ctx context.Context, in *agentsv1.AttachResourceRequest) (*agentsv1.AttachResourceResponse, error) {
	l, err := s.Svc.AttachResource(ctx, in.GetAgentId(), in.GetResourceId(), in.GetMountPath())
	if err != nil {
		return nil, err
	}
	return &agentsv1.AttachResourceResponse{Link: &agentsv1.AgentResource{
		AgentId: l.AgentID, ResourceId: l.ResourceID, MountPath: l.MountPath,
	}}, nil
}

func toProtoGroup(g ports.Group) *agentsv1.AgentGroup {
	return &agentsv1.AgentGroup{
		Id: g.ID, Name: g.Name, Description: g.Description,
		CreatedAt: timestamppb.New(g.CreatedAt),
	}
}

func toProtoAgent(a ports.Agent) *agentsv1.Agent {
	return &agentsv1.Agent{
		Id: a.ID, GroupId: a.GroupID, Name: a.Name, Role: a.Role,
		PiWorkspace: a.PiWorkspace, Env: a.Env,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
}

func toProtoResource(r ports.Resource) *agentsv1.ResourceBundle {
	return &agentsv1.ResourceBundle{Id: r.ID, Name: r.Name, Kind: r.Kind, Spec: r.Spec}
}

func pageLimit(p *commonv1.Page) int {
	if p == nil {
		return 100
	}
	if p.Limit <= 0 {
		return 100
	}
	return int(p.Limit)
}

func pageCursor(p *commonv1.Page) string {
	if p == nil {
		return ""
	}
	return p.Cursor
}
