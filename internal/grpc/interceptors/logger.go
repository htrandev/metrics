package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Logger(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()

		res, err := handler(ctx, req)

		elapsed := time.Since(start)
		if err != nil {
			log.Error("got incoming HTTP request",
				zap.String("full method", info.FullMethod),
				zap.Duration("elapsed", elapsed),
				zap.Error(err),
			)
		} else {
			log.Info("got incoming HTTP request",
				zap.String("full method", info.FullMethod),
				zap.Duration("elapsed", elapsed),
			)
		}

		return res, err
	}
}
