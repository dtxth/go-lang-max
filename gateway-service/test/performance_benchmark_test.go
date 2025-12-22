package test

import (
	"context"
	"sync"
	"testing"
	"time"

	"gateway-service/internal/config"
	errorHandler "gateway-service/internal/infrastructure/errors"
	grpcClient "gateway-service/internal/infrastructure/grpc"
	"gateway-service/internal/infrastructure/retry"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BenchmarkGRPCClientCreation benchmarks gRPC client creation performance
func BenchmarkGRPCClientCreation(b *testing.B) {
	cfg := &config.Config{
		Services: config.ServicesConfig{
			Auth: config.ServiceConfig{
				Address:           "localhost:50051",
				Timeout:           1 * time.Second,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clientManager := grpcClient.NewClientManager(cfg)
		_ = clientManager
	}
}

// BenchmarkClientRetrieval benchmarks client retrieval from connection pool
func BenchmarkClientRetrieval(b *testing.B) {
	cfg := &config.Config{
		Services: config.ServicesConfig{
			Auth: config.ServiceConfig{
				Address:           "localhost:50051",
				Timeout:           1 * time.Second,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
			},
		},
	}

	clientManager := grpcClient.NewClientManager(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := clientManager.GetAuthClient()
		_ = client
	}
}

// BenchmarkRetryLogicExecution benchmarks retry logic execution
func BenchmarkRetryLogicExecution(b *testing.B) {
	cfg := config.ServiceConfig{
		MaxRetries:        3,
		RetryDelay:        1 * time.Millisecond, // Very fast for benchmarking
		MaxRetryDelay:     10 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	errHandler := errorHandler.NewErrorHandler()
	retrier := retry.NewRetrier(cfg, errHandler)

	// Successful function (no retries)
	successFunc := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := retrier.Execute(ctx, successFunc)
		cancel()
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkRetryLogicWithFailures benchmarks retry logic with failures
func BenchmarkRetryLogicWithFailures(b *testing.B) {
	cfg := config.ServiceConfig{
		MaxRetries:        2, // Reduced for faster benchmarking
		RetryDelay:        1 * time.Millisecond,
		MaxRetryDelay:     5 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	errHandler := errorHandler.NewErrorHandler()
	retrier := retry.NewRetrier(cfg, errHandler)

	// Always failing function
	failFunc := func(ctx context.Context) error {
		return status.Error(codes.Unavailable, "service unavailable")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := retrier.Execute(ctx, failFunc)
		cancel()
		if err == nil {
			b.Fatalf("Expected error but got none")
		}
	}
}

// BenchmarkCircuitBreakerExecution benchmarks circuit breaker execution
func BenchmarkCircuitBreakerExecution(b *testing.B) {
	cfg := config.CircuitBreakerConfig{
		MaxRequests: 100,
		Interval:    1 * time.Second,
		Timeout:     1 * time.Second,
	}

	cb := grpcClient.NewCircuitBreaker(cfg)

	successFunc := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cb.Execute(context.Background(), successFunc)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkConfigurationLoading benchmarks configuration loading
func BenchmarkConfigurationLoading(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := config.Load()
		_ = cfg
	}
}

// BenchmarkTimeoutContextCreation benchmarks timeout context creation
func BenchmarkTimeoutContextCreation(b *testing.B) {
	timeout := 1 * time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		cancel()
		_ = ctx
	}
}

// BenchmarkConcurrentClientAccess benchmarks concurrent access to client manager
func BenchmarkConcurrentClientAccess(b *testing.B) {
	cfg := &config.Config{
		Services: config.ServicesConfig{
			Auth: config.ServiceConfig{
				Address:           "localhost:50051",
				Timeout:           1 * time.Second,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
			},
		},
	}

	clientManager := grpcClient.NewClientManager(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client := clientManager.GetAuthClient()
			_ = client
		}
	})
}

// BenchmarkErrorHandling benchmarks error handling performance
func BenchmarkErrorHandling(b *testing.B) {
	errHandler := errorHandler.NewErrorHandler()
	testErr := status.Error(codes.Internal, "test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isRetryable := errHandler.IsRetryableError(testErr)
		_ = isRetryable
	}
}

// BenchmarkMetricsCollection benchmarks performance metrics collection
func BenchmarkMetricsCollection(b *testing.B) {
	// Simple metrics collection without external dependencies
	type simpleMetrics struct {
		count    int64
		duration time.Duration
		mu       sync.Mutex
	}
	
	metrics := &simpleMetrics{}
	duration := 100 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.mu.Lock()
		metrics.count++
		metrics.duration += duration
		metrics.mu.Unlock()
	}
}

// BenchmarkConcurrentMetricsCollection benchmarks concurrent metrics collection
func BenchmarkConcurrentMetricsCollection(b *testing.B) {
	// Simple metrics collection without external dependencies
	type simpleMetrics struct {
		count    int64
		duration time.Duration
		mu       sync.Mutex
	}
	
	metrics := &simpleMetrics{}
	duration := 100 * time.Millisecond

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metrics.mu.Lock()
			metrics.count++
			metrics.duration += duration
			metrics.mu.Unlock()
		}
	})
}