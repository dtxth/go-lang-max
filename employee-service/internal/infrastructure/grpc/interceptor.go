package grpc

import (
	"context"
	"time"

	"employee-service/internal/infrastructure/logger"
	"google.golang.org/grpc"
)

// LoggingInterceptor logs all gRPC calls with duration
func LoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Log request start
		log.Info(ctx, "gRPC request started", map[string]interface{}{
			"method": info.FullMethod,
		})

		// Call handler
		resp, err := handler(ctx, req)

		// Log request completion
		duration := time.Since(start)
		fields := map[string]interface{}{
			"method":      info.FullMethod,
			"duration_ms": duration.Milliseconds(),
			"duration":    duration.String(),
		}

		if err != nil {
			fields["error"] = err.Error()
			log.Error(ctx, "gRPC request failed", fields)
		} else {
			log.Info(ctx, "gRPC request completed", fields)
		}

		return resp, err
	}
}
