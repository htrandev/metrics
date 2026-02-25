package interceptors

import (
	"context"
	"net"

	"github.com/htrandev/metrics/pkg/metadatautil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Subnet(subnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ip := net.ParseIP(metadatautil.GetRealIP(ctx))
		if len(ip) != 0 && !subnet.Contains(ip) {
			return nil, status.Errorf(codes.PermissionDenied, "not trusted ip: %s", ip.String())
		}

		return handler(ctx, req)
	}
}
