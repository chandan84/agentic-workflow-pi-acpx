// Package grpcsrv exposes the audit + (proxied) execution surface to the desktop.
//
// The gateway is a thin aggregator: business-domain RPCs proxy to the upstream
// services through their generated gRPC clients (wired in main.go). Only audit
// has a service-owned implementation here.
package grpcsrv

import (
	"context"

	commonv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/common/v1"

	auditv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/audit/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/gateway/internal/audit"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuditServer implements audit.v1.AuditServiceServer using a Source.
type AuditServer struct {
	auditv1.UnimplementedAuditServiceServer
	Src audit.Source
}

// NewAudit wires an AuditServer.
func NewAudit(src audit.Source) *AuditServer { return &AuditServer{Src: src} }

// ListAudit returns historical events.
func (s *AuditServer) ListAudit(ctx context.Context, in *auditv1.ListAuditRequest) (*auditv1.ListAuditResponse, error) {
	limit := 200
	if p := in.GetPage(); p != nil && p.Limit > 0 {
		limit = int(p.Limit)
	}
	rows, err := s.Src.List(ctx, in.GetRunId(), limit)
	if err != nil {
		return nil, err
	}
	out := &auditv1.ListAuditResponse{PageInfo: &commonv1.PageInfo{}}
	for _, e := range rows {
		out.Events = append(out.Events, toProto(e))
	}
	return out, nil
}

// TailAudit streams live events.
func (s *AuditServer) TailAudit(in *auditv1.TailAuditRequest, stream auditv1.AuditService_TailAuditServer) error {
	done := make(chan struct{})
	cancel, err := s.Src.Subscribe(stream.Context(), in.GetRunId(), func(e audit.Event) {
		_ = stream.Send(toProto(e))
	})
	if err != nil {
		return err
	}
	defer cancel()
	<-stream.Context().Done()
	close(done)
	return nil
}

func toProto(e audit.Event) *auditv1.AuditEvent {
	return &auditv1.AuditEvent{
		Id: e.ID, RunId: e.RunID, NodeId: e.NodeID,
		Layer: e.Layer, Kind: e.Kind, PayloadJson: e.PayloadJSON,
		At: timestamppb.New(e.At),
	}
}
