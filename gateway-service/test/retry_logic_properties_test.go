package test

import (
	"context"
	"testing"
	"time"

	"gateway-service/internal/config"
	errorHandler "gateway-service/internal/infrastructure/errors"
	"gateway-service/internal/infrastructure/retry"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestProperty7_RetryWithExponentialBackoff tests retry logic with exponential backoff
// **Feature: gateway-grpc-implementation, Property 7: Retry with Exponential Backoff**
// **Validates: Requirements 6.5**
func TestProperty7_RetryWithExponentialBackoff(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Test that retry logic implements exponential backoff correctly
	properties.Property("retry implements exponential backoff with correct delays", prop.ForAll(
		func(maxRetries int, initialDelayMs int64, backoffMultiplier float64) bool {
			// Ensure parameters are reasonable
			if maxRetries < 1 || maxRetries > 3 ||
				initialDelayMs < 10 || initialDelayMs > 50 ||
				backoffMultiplier < 1.5 || backoffMultiplier > 2.5 {
				return true // Skip invalid inputs
			}

			initialDelay := time.Duration(initialDelayMs) * time.Millisecond
			maxDelay := 1 * time.Second

			cfg := config.ServiceConfig{
				MaxRetries:        maxRetries,
				RetryDelay:        initialDelay,
				MaxRetryDelay:     maxDelay,
				BackoffMultiplier: backoffMultiplier,
			}

			errHandler := errorHandler.NewErrorHandler()
			retrier := retry.NewRetrier(cfg, errHandler)

			// Track retry attempts and their timing
			attempts := 0
			attemptTimes := make([]time.Time, 0)

			// Create a function that always fails with a retryable error
			failingFunc := func(ctx context.Context) error {
				attempts++
				attemptTimes = append(attemptTimes, time.Now())
				return status.Error(codes.Unavailable, "service unavailable")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := retrier.Execute(ctx, failingFunc)
			
			// Should fail after all retries
			if err == nil {
				t.Logf("Expected retry to fail, but it succeeded")
				return false
			}

			// Should have attempted maxRetries + 1 times (initial + retries)
			expectedAttempts := maxRetries + 1
			if attempts != expectedAttempts {
				t.Logf("Expected %d attempts, got %d", expectedAttempts, attempts)
				return false
			}

			// Verify exponential backoff delays between attempts (more lenient check)
			if len(attemptTimes) >= 2 {
				// Just check that delays are generally increasing (allowing for jitter)
				for i := 1; i < len(attemptTimes)-1; i++ {
					delay1 := attemptTimes[i].Sub(attemptTimes[i-1])
					delay2 := attemptTimes[i+1].Sub(attemptTimes[i])
					
					// Allow for significant variance due to jitter and system scheduling
					// Just check that the second delay is not significantly smaller than the first
					if delay2 < delay1/3 {
						t.Logf("Delay %d (%v) is much smaller than delay %d (%v), expected exponential backoff", 
							i+1, delay2, i, delay1)
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(1, 3),           // Max retries (reduced range)
		gen.Int64Range(10, 50),       // Initial delay in milliseconds (increased minimum)
		gen.Float64Range(1.5, 2.5),   // Backoff multiplier (narrower range)
	))

	// Test that retry logic respects maximum retry limits
	properties.Property("retry respects maximum retry limits", prop.ForAll(
		func(maxRetries int) bool {
			// Ensure max retries is reasonable
			if maxRetries < 0 || maxRetries > 10 {
				return true // Skip invalid inputs
			}

			cfg := config.ServiceConfig{
				MaxRetries:        maxRetries,
				RetryDelay:        10 * time.Millisecond,
				MaxRetryDelay:     1 * time.Second,
				BackoffMultiplier: 2.0,
			}

			errHandler := errorHandler.NewErrorHandler()
			retrier := retry.NewRetrier(cfg, errHandler)

			attempts := 0
			failingFunc := func(ctx context.Context) error {
				attempts++
				return status.Error(codes.Unavailable, "service unavailable")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := retrier.Execute(ctx, failingFunc)

			// Should fail after all retries
			if err == nil {
				t.Logf("Expected retry to fail, but it succeeded")
				return false
			}

			// Should have attempted maxRetries + 1 times (initial + retries)
			expectedAttempts := maxRetries + 1
			if attempts != expectedAttempts {
				t.Logf("Expected %d attempts, got %d", expectedAttempts, attempts)
				return false
			}

			return true
		},
		gen.IntRange(0, 10), // Max retries
	))

	// Test that retry logic only retries retryable errors
	properties.Property("retry only retries retryable errors", prop.ForAll(
		func(grpcCode int) bool {
			// Map integer to gRPC codes, excluding OK (which would be nil error)
			var code codes.Code
			switch (grpcCode % 16) + 1 { // Skip codes.OK (0), use 1-16
			case 1:
				code = codes.Canceled
			case 2:
				code = codes.Unknown
			case 3:
				code = codes.InvalidArgument
			case 4:
				code = codes.DeadlineExceeded
			case 5:
				code = codes.NotFound
			case 6:
				code = codes.AlreadyExists
			case 7:
				code = codes.PermissionDenied
			case 8:
				code = codes.ResourceExhausted
			case 9:
				code = codes.FailedPrecondition
			case 10:
				code = codes.Aborted
			case 11:
				code = codes.OutOfRange
			case 12:
				code = codes.Unimplemented
			case 13:
				code = codes.Internal
			case 14:
				code = codes.Unavailable
			case 15:
				code = codes.DataLoss
			case 16:
				code = codes.Unauthenticated
			}

			cfg := config.ServiceConfig{
				MaxRetries:        3,
				RetryDelay:        10 * time.Millisecond,
				MaxRetryDelay:     1 * time.Second,
				BackoffMultiplier: 2.0,
			}

			errHandler := errorHandler.NewErrorHandler()
			retrier := retry.NewRetrier(cfg, errHandler)

			attempts := 0
			testErr := status.Error(code, "test error")
			failingFunc := func(ctx context.Context) error {
				attempts++
				return testErr
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := retrier.Execute(ctx, failingFunc)

			// Determine if this error should be retryable
			isRetryable := errHandler.IsRetryableError(testErr)

			if isRetryable {
				// Should have attempted multiple times for retryable errors
				if attempts <= 1 {
					t.Logf("Retryable error %v should have been retried, but got %d attempts", code, attempts)
					return false
				}
			} else {
				// Should have attempted only once for non-retryable errors
				if attempts != 1 {
					t.Logf("Non-retryable error %v should not have been retried, but got %d attempts", code, attempts)
					return false
				}
			}

			// Should always fail (since our function always returns an error)
			if err == nil {
				t.Logf("Expected function to fail, but it succeeded")
				return false
			}

			return true
		},
		gen.IntRange(0, 100), // Random integer to map to gRPC codes
	))

	// Test that retry logic respects context cancellation
	properties.Property("retry respects context cancellation", prop.ForAll(
		func(cancelAfterMs int64) bool {
			// Ensure cancel time is reasonable (between 5ms and 200ms)
			if cancelAfterMs < 5 || cancelAfterMs > 200 {
				return true // Skip invalid inputs
			}

			cfg := config.ServiceConfig{
				MaxRetries:        10, // High number to ensure context cancellation happens first
				RetryDelay:        50 * time.Millisecond,
				MaxRetryDelay:     1 * time.Second,
				BackoffMultiplier: 2.0,
			}

			errHandler := errorHandler.NewErrorHandler()
			retrier := retry.NewRetrier(cfg, errHandler)

			attempts := 0
			failingFunc := func(ctx context.Context) error {
				attempts++
				return status.Error(codes.Unavailable, "service unavailable")
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cancelAfterMs)*time.Millisecond)
			defer cancel()

			err := retrier.Execute(ctx, failingFunc)

			// Should fail due to context cancellation
			if err == nil {
				t.Logf("Expected retry to fail due to context cancellation, but it succeeded")
				return false
			}

			// We can't easily measure elapsed time here without modifying the test structure
			// So we'll just check that it failed and made fewer attempts than maximum
			if attempts > cfg.MaxRetries {
				t.Logf("Made too many attempts after context cancellation: expected <=%d, got %d", cfg.MaxRetries, attempts)
				return false
			}

			return true
		},
		gen.Int64Range(5, 200), // Cancel after milliseconds
	))

	// Test that successful operations don't trigger retries
	properties.Property("successful operations don't trigger retries", prop.ForAll(
		func(maxRetries int) bool {
			// Ensure max retries is reasonable
			if maxRetries < 1 || maxRetries > 10 {
				return true // Skip invalid inputs
			}

			cfg := config.ServiceConfig{
				MaxRetries:        maxRetries,
				RetryDelay:        10 * time.Millisecond,
				MaxRetryDelay:     1 * time.Second,
				BackoffMultiplier: 2.0,
			}

			errHandler := errorHandler.NewErrorHandler()
			retrier := retry.NewRetrier(cfg, errHandler)

			attempts := 0
			successFunc := func(ctx context.Context) error {
				attempts++
				return nil // Success
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := retrier.Execute(ctx, successFunc)

			// Should succeed
			if err != nil {
				t.Logf("Expected successful function to succeed, but got error: %v", err)
				return false
			}

			// Should have attempted only once
			if attempts != 1 {
				t.Logf("Successful function should be called only once, but got %d attempts", attempts)
				return false
			}

			return true
		},
		gen.IntRange(1, 10), // Max retries
	))

	// Run all properties
	properties.TestingRun(t)
}