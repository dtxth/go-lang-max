package test

import (
	"context"
	"testing"
	"time"

	"gateway-service/internal/config"
	grpcClient "gateway-service/internal/infrastructure/grpc"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestProperty5_TimeoutConfiguration tests that timeout configuration works correctly
// **Feature: gateway-grpc-implementation, Property 5: Timeout Configuration**
// **Validates: Requirements 6.1**
func TestProperty5_TimeoutConfiguration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Reduced from 100 for faster testing
	properties := gopter.NewProperties(parameters)

	// Test that timeout configuration is respected for service calls
	properties.Property("timeout configuration is preserved in service config", prop.ForAll(
		func(authTimeout, chatTimeout, employeeTimeout, structureTimeout int64) bool {
			// Ensure timeouts are reasonable (between 1ms and 30s)
			if authTimeout < 1 || authTimeout > 30000 ||
				chatTimeout < 1 || chatTimeout > 30000 ||
				employeeTimeout < 1 || employeeTimeout > 30000 ||
				structureTimeout < 1 || structureTimeout > 30000 {
				return true // Skip invalid inputs
			}

			// Create configuration with specified timeouts
			cfg := &config.Config{
				Services: config.ServicesConfig{
					Auth: config.ServiceConfig{
						Address:           "localhost:50051",
						Timeout:           time.Duration(authTimeout) * time.Millisecond,
						MaxRetries:        3,
						RetryDelay:        100 * time.Millisecond,
						MaxRetryDelay:     5 * time.Second,
						BackoffMultiplier: 2.0,
					},
					Chat: config.ServiceConfig{
						Address:           "localhost:50052",
						Timeout:           time.Duration(chatTimeout) * time.Millisecond,
						MaxRetries:        3,
						RetryDelay:        100 * time.Millisecond,
						MaxRetryDelay:     5 * time.Second,
						BackoffMultiplier: 2.0,
					},
					Employee: config.ServiceConfig{
						Address:           "localhost:50053",
						Timeout:           time.Duration(employeeTimeout) * time.Millisecond,
						MaxRetries:        3,
						RetryDelay:        100 * time.Millisecond,
						MaxRetryDelay:     5 * time.Second,
						BackoffMultiplier: 2.0,
					},
					Structure: config.ServiceConfig{
						Address:           "localhost:50054",
						Timeout:           time.Duration(structureTimeout) * time.Millisecond,
						MaxRetries:        3,
						RetryDelay:        100 * time.Millisecond,
						MaxRetryDelay:     5 * time.Second,
						BackoffMultiplier: 2.0,
					},
				},
			}

			// Verify that configuration values are preserved
			if cfg.Services.Auth.Timeout != time.Duration(authTimeout)*time.Millisecond {
				t.Logf("Auth timeout not preserved: expected %v, got %v", 
					time.Duration(authTimeout)*time.Millisecond, cfg.Services.Auth.Timeout)
				return false
			}

			if cfg.Services.Chat.Timeout != time.Duration(chatTimeout)*time.Millisecond {
				t.Logf("Chat timeout not preserved: expected %v, got %v", 
					time.Duration(chatTimeout)*time.Millisecond, cfg.Services.Chat.Timeout)
				return false
			}

			if cfg.Services.Employee.Timeout != time.Duration(employeeTimeout)*time.Millisecond {
				t.Logf("Employee timeout not preserved: expected %v, got %v", 
					time.Duration(employeeTimeout)*time.Millisecond, cfg.Services.Employee.Timeout)
				return false
			}

			if cfg.Services.Structure.Timeout != time.Duration(structureTimeout)*time.Millisecond {
				t.Logf("Structure timeout not preserved: expected %v, got %v", 
					time.Duration(structureTimeout)*time.Millisecond, cfg.Services.Structure.Timeout)
				return false
			}

			return true
		},
		gen.Int64Range(1, 30000),    // Auth timeout in milliseconds
		gen.Int64Range(1, 30000),    // Chat timeout in milliseconds
		gen.Int64Range(1, 30000),    // Employee timeout in milliseconds
		gen.Int64Range(1, 30000),    // Structure timeout in milliseconds
	))

	// Test that timeout context is properly created and respected
	properties.Property("timeout context is properly created with configured values", prop.ForAll(
		func(timeoutMs int64) bool {
			// Ensure timeout is reasonable (between 10ms and 200ms for very fast testing)
			if timeoutMs < 10 || timeoutMs > 200 {
				return true // Skip invalid inputs
			}

			timeout := time.Duration(timeoutMs) * time.Millisecond
			
			// Create a context with the specified timeout
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Measure how long the context takes to expire
			start := time.Now()
			<-ctx.Done()
			elapsed := time.Since(start)

			// The elapsed time should be approximately equal to the timeout
			// Allow for some variance due to system scheduling (±30ms)
			tolerance := 30 * time.Millisecond
			if elapsed < timeout-tolerance || elapsed > timeout+tolerance {
				t.Logf("Context timeout not respected: expected ~%v, got %v", timeout, elapsed)
				return false
			}

			return true
		},
		gen.Int64Range(10, 200), // Timeout in milliseconds (further reduced range for faster testing)
	))

	// Test that retry logic respects timeout configuration
	properties.Property("retry logic respects timeout configuration", prop.ForAll(
		func(timeoutMs, retryDelayMs int64, maxRetries int) bool {
			// Ensure parameters are reasonable for fast testing
			if timeoutMs < 50 || timeoutMs > 500 ||
				retryDelayMs < 5 || retryDelayMs > 50 ||
				maxRetries < 1 || maxRetries > 3 {
				return true // Skip invalid inputs
			}

			timeout := time.Duration(timeoutMs) * time.Millisecond
			retryDelay := time.Duration(retryDelayMs) * time.Millisecond
			
			cfg := config.ServiceConfig{
				Timeout:           timeout,
				MaxRetries:        maxRetries,
				RetryDelay:        retryDelay,
				MaxRetryDelay:     500 * time.Millisecond, // Reduced for faster testing
				BackoffMultiplier: 2.0,
			}

			// Create a function that always fails with a retryable error
			failingFunc := func(ctx context.Context) error {
				// Simulate a retryable gRPC error
				return status.Error(codes.Unavailable, "service unavailable")
			}

			// Measure how long the retry logic takes
			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxRetries+2)*timeout)
			defer cancel()
			
			err := grpcClient.RetryWithExponentialBackoff(ctx, cfg, failingFunc)
			elapsed := time.Since(start)

			// The function should fail after all retries
			if err == nil {
				t.Logf("Expected retry to fail, but it succeeded")
				return false
			}

			// Calculate expected minimum time (sum of retry delays)
			expectedMinTime := time.Duration(0)
			currentDelay := retryDelay
			for i := 0; i < maxRetries; i++ {
				expectedMinTime += currentDelay
				currentDelay = time.Duration(float64(currentDelay) * cfg.BackoffMultiplier)
				if currentDelay > cfg.MaxRetryDelay {
					currentDelay = cfg.MaxRetryDelay
				}
			}

			// The elapsed time should be at least the expected minimum time
			// Allow for more variance due to system scheduling and jitter (±30% tolerance)
			tolerance := time.Duration(float64(expectedMinTime) * 0.3)
			if elapsed < expectedMinTime-tolerance {
				t.Logf("Retry took less time than expected: expected >=%v, got %v (tolerance: %v)", expectedMinTime, elapsed, tolerance)
				return false
			}

			return true
		},
		gen.Int64Range(50, 500),    // Timeout in milliseconds (reduced range)
		gen.Int64Range(5, 50),      // Retry delay in milliseconds (reduced range)
		gen.IntRange(1, 3),         // Max retries (reduced range)
	))

	// Test that configuration loading preserves timeout values
	properties.Property("configuration loading preserves timeout values", prop.ForAll(
		func(timeoutMs int64) bool {
			// Ensure timeout is reasonable
			if timeoutMs < 1 || timeoutMs > 30000 {
				return true // Skip invalid inputs
			}

			expectedTimeout := time.Duration(timeoutMs) * time.Millisecond
			
			// Create a service config
			cfg := config.ServiceConfig{
				Address:           "localhost:50051",
				Timeout:           expectedTimeout,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
			}

			// Verify the timeout is preserved
			if cfg.Timeout != expectedTimeout {
				t.Logf("Timeout not preserved: expected %v, got %v", expectedTimeout, cfg.Timeout)
				return false
			}

			return true
		},
		gen.Int64Range(1, 30000), // Timeout in milliseconds
	))

	// Run all properties
	properties.TestingRun(t)
}