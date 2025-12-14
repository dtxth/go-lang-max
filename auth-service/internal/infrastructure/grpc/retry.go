package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxRetries int
	Backoff    []time.Duration
}

// DefaultRetryConfig returns the default retry configuration
// Retry up to 3 times with exponential backoff (1s, 2s, 4s)
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		Backoff:    []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second},
	}
}

// IsRetryableError determines if a gRPC error should be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		// Non-gRPC errors are retryable
		return true
	}

	// Retry on transient errors
	switch st.Code() {
	case codes.Unavailable,
		codes.DeadlineExceeded,
		codes.ResourceExhausted,
		codes.Aborted,
		codes.Internal,
		codes.Unknown:
		return true
	default:
		return false
	}
}

// WithRetry wraps a gRPC call with retry logic
// It retries up to 3 times with exponential backoff (1s, 2s, 4s)
// Logs each retry attempt
func WithRetry(ctx context.Context, operation string, fn func() error) error {
	config := DefaultRetryConfig()
	return WithRetryConfig(ctx, operation, fn, config)
}

// WithRetryConfig wraps a gRPC call with custom retry configuration
func WithRetryConfig(ctx context.Context, operation string, fn func() error, config RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff duration
			backoffDuration := config.Backoff[attempt-1]
			
			log.Printf("[gRPC Retry] Attempt %d/%d for %s after %v backoff", 
				attempt+1, config.MaxRetries+1, operation, backoffDuration)

			// Wait with context cancellation support
			select {
			case <-time.After(backoffDuration):
				// Continue with retry
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			}
		}

		// Execute the operation
		err := fn()
		if err == nil {
			if attempt > 0 {
				log.Printf("[gRPC Retry] Success for %s after %d retries", operation, attempt)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryableError(err) {
			log.Printf("[gRPC Retry] Non-retryable error for %s: %v", operation, err)
			return err
		}

		// Log retry attempt
		if attempt < config.MaxRetries {
			log.Printf("[gRPC Retry] Retryable error for %s (attempt %d/%d): %v", 
				operation, attempt+1, config.MaxRetries+1, err)
		}
	}

	// All retries exhausted
	log.Printf("[gRPC Retry] All retries exhausted for %s: %v", operation, lastErr)
	return fmt.Errorf("gRPC call failed after %d retries: %w", config.MaxRetries, lastErr)
}

// UnaryClientInterceptor returns a gRPC unary client interceptor with retry logic
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return WithRetry(ctx, method, func() error {
			return invoker(ctx, method, req, reply, cc, opts...)
		})
	}
}
