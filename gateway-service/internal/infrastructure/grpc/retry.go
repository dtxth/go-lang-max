package grpc

import (
	"context"
	"fmt"
	"math"
	"time"

	"gateway-service/internal/config"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryableFunc represents a function that can be retried
type RetryableFunc func(ctx context.Context) error

// RetryWithExponentialBackoff executes a function with retry logic and exponential backoff
func RetryWithExponentialBackoff(ctx context.Context, cfg config.ServiceConfig, fn RetryableFunc) error {
	var lastErr error
	
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !isRetryableError(err) {
			return err // Non-retryable error, fail immediately
		}
		
		// Don't sleep after the last attempt
		if attempt == cfg.MaxRetries {
			break
		}
		
		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt, cfg)
		
		// Create timer for delay
		timer := time.NewTimer(delay)
		defer timer.Stop()
		
		// Wait for delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			// Continue to next attempt
		}
	}
	
	return fmt.Errorf("max retries (%d) exceeded, last error: %w", cfg.MaxRetries, lastErr)
}

// isRetryableError determines if an error should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check gRPC status codes
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable,
			codes.DeadlineExceeded,
			codes.ResourceExhausted,
			codes.Aborted,
			codes.Internal:
			return true
		default:
			return false
		}
	}
	
	// For non-gRPC errors, we can add additional logic here
	// For now, we'll be conservative and not retry
	return false
}

// calculateDelay calculates the delay for the given attempt using exponential backoff
func calculateDelay(attempt int, cfg config.ServiceConfig) time.Duration {
	// Calculate exponential backoff: delay = base_delay * (multiplier ^ attempt)
	delay := float64(cfg.RetryDelay) * math.Pow(cfg.BackoffMultiplier, float64(attempt))
	
	// Apply jitter (±10% randomization to avoid thundering herd)
	jitter := 0.1 * delay * float64(2*time.Now().UnixNano()%2 - 1) // Simple jitter
	delay += jitter
	
	// Ensure delay doesn't exceed maximum
	if time.Duration(delay) > cfg.MaxRetryDelay {
		delay = float64(cfg.MaxRetryDelay)
	}
	
	// Ensure delay is not negative
	if delay < 0 {
		delay = float64(cfg.RetryDelay)
	}
	
	return time.Duration(delay)
}

// WithRetry wraps a gRPC call with retry logic
func WithRetry[T any](ctx context.Context, cfg config.ServiceConfig, call func(ctx context.Context) (T, error)) (T, error) {
	var result T
	
	retryFunc := func(ctx context.Context) error {
		var err error
		result, err = call(ctx)
		return err
	}
	
	err := RetryWithExponentialBackoff(ctx, cfg, retryFunc)
	if err != nil {
		return result, err
	}
	
	return result, nil
}

// WithRetryNoResult wraps a gRPC call that returns no result with retry logic
func WithRetryNoResult(ctx context.Context, cfg config.ServiceConfig, call func(ctx context.Context) error) error {
	return RetryWithExponentialBackoff(ctx, cfg, call)
}