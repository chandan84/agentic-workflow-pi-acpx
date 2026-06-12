// Package upstreams holds gRPC client dials and proxy registration for the
// gateway's aggregated RPCs.
//
// The gateway registers the same service interfaces it exposes (agents, flow,
// execution) and forwards each call to the upstream service. This keeps the
// desktop a single-target client.
package upstreams

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial returns a gRPC ClientConn with insecure transport (dev mode).
func Dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
