package client

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/htrandev/metrics/internal/agent"
	"github.com/htrandev/metrics/internal/model"
	pb "github.com/htrandev/metrics/internal/proto"
	"github.com/htrandev/metrics/pkg/metadatautil"
	"github.com/htrandev/metrics/pkg/sign"
	"google.golang.org/protobuf/proto"
)

var _ agent.Client = (*GRPCClient)(nil)

type GRPCClient struct {
	client pb.MetricsClient
	opts   CommonOptions
}

func NewGRPC(client pb.MetricsClient, opts ...Option) *GRPCClient {
	c := &GRPCClient{
		client: client,
	}
	for _, opt := range opts {
		opt(&c.opts)
	}
	return c
}

func (c *GRPCClient) Send(ctx context.Context, metrics []model.MetricDto) error {
	req := buildGRPCRequest(metrics)

	ctx, err := c.setMetadata(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc client: set metadata: %w", err)
	}

	if _, err := c.client.UpdateMetrics(ctx, req); err != nil {
		return fmt.Errorf("grpc client: update metrics: %w", err)
	}
	return nil
}

func (c *GRPCClient) setMetadata(ctx context.Context, req *pb.UpdateMetricsRequest) (context.Context, error) {
	ctx = metadatautil.SetRealIP(ctx, c.opts.ip)
	if c.opts.signature == "" {
		b, err := proto.Marshal(req)
		if err != nil {
			return ctx, fmt.Errorf("unable to marshal req: %w", err)
		}
		si := sign.Signature(c.opts.signature)
		signature := si.Sign(b)
		hash := base64.RawURLEncoding.EncodeToString(signature)
		ctx = metadatautil.SetHash256(ctx, hash)
	}
	return ctx, nil
}

func buildGRPCRequest(metrics []model.MetricDto) *pb.UpdateMetricsRequest {
	pbMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, metric := range metrics {
		pbMetrics = append(pbMetrics, model.ToProto(metric))
	}

	builder := pb.UpdateMetricsRequest_builder{
		Metrics: pbMetrics,
	}
	return builder.Build()
}
