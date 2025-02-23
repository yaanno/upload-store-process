package interceptor

import (
	"context"
	"time"

	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc"
)

func LoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		log.Info().
			Str("method", info.FullMethod).
			Interface("request", req).
			Msg("Received gRPC request")

		resp, err := handler(ctx, req)

		log.Info().
			Str("method", info.FullMethod).
			Dur("duration", time.Since(start)).
			Err(err).
			Msg("Completed gRPC request")

		return resp, err
	}
}
