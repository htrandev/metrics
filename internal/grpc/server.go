package grpc

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/htrandev/metrics/internal/contracts"
	"github.com/htrandev/metrics/internal/model"
	pb "github.com/htrandev/metrics/internal/proto"
	"github.com/htrandev/metrics/pkg/metadatautil"
	"github.com/htrandev/metrics/pkg/sign"
)

type MetricServerOptions struct {
	Service   contracts.Service
	Signature string
}

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	opts *MetricServerOptions
}

func New(opts *MetricServerOptions) *MetricsServer {
	return &MetricsServer{opts: opts}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil {
		return &pb.UpdateMetricsResponse{}, nil
	}

	if s.opts.Signature != "" {
		valid, err := s.checkHash(ctx, req)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "check hash: %s", err.Error())
		}
		if !valid {
			return nil, status.Error(codes.DataLoss, "request not valid")
		}
	}

	metrics := buildMetrics(req.GetMetrics())

	if err := s.opts.Service.StoreManyWithRetry(ctx, metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "store many with retry: %s", err.Error())
	}
	return &pb.UpdateMetricsResponse{}, nil
}

func buildMetrics(metrics []*pb.Metric) []model.MetricDto {
	result := make([]model.MetricDto, 0, len(metrics))

	for _, metric := range metrics {
		result = append(result, model.FromProto(metric))
	}

	return result
}

// checkHash check if request is valid.
func (s *MetricsServer) checkHash(ctx context.Context, req *pb.UpdateMetricsRequest) (bool, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return false, fmt.Errorf("unable to marshal req: %w", err)
	}
	si := sign.Signature(s.opts.Signature)
	signature := si.Sign(b)
	gotHash := base64.RawURLEncoding.EncodeToString(signature)
	receivedHash := metadatautil.GetHash256(ctx)
	if gotHash == receivedHash {
		return true, nil
	}
	return false, nil
}
