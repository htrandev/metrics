package interceptors

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/htrandev/metrics/internal/proto"
	"github.com/htrandev/metrics/pkg/metadatautil"
	"github.com/htrandev/metrics/pkg/sign"
)

func Signature(signature string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if msg, ok := req.(*pb.UpdateMetricsRequest); ok {
			valid, err := checkHash(ctx, signature, msg)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "check hash: %s", err.Error())
			}
			if !valid {
				return nil, status.Error(codes.DataLoss, "request not valid")
			}
			return handler(ctx, req)
		}
		return nil, status.Error(codes.InvalidArgument, "request is wrong type")
	}
}

// checkHash check if request is valid.
func checkHash(ctx context.Context, signature string, req *pb.UpdateMetricsRequest) (bool, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return false, fmt.Errorf("unable to marshal req: %w", err)
	}
	s := sign.Signature(signature)
	si := s.Sign(b)
	gotHash := base64.RawURLEncoding.EncodeToString(si)
	receivedHash := metadatautil.GetHash256(ctx)
	if gotHash == receivedHash {
		return true, nil
	}
	return false, nil
}
