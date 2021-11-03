package metrics

import (
	"context"

	"github.com/uber-go/tally/v4"
	"google.golang.org/grpc/stats"
)

type grpcStatsHandler struct {
	requestLatency tally.Timer
}

// NewGRPCStatsHandler creates a gRPC stats handler for the given scope that
// records gRPC stats.
func NewGRPCStatsHandler(scope tally.Scope) stats.Handler {
	return &grpcStatsHandler{requestLatency: scope.Timer("grpc_request_latency")}
}

func (*grpcStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

func (*grpcStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

func (*grpcStatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {}

func (g *grpcStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	// Record time difference
	if end, _ := s.(*stats.End); end != nil {
		g.requestLatency.Record(end.EndTime.Sub(end.BeginTime))
	}
}
