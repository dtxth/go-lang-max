package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWithRetry_Success(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}

	err := WithRetry(context.Background(), "test-operation", operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 3 {
			return status.Error(codes.Unavailable, "service unavailable")
		}
		return nil
	}

	config := RetryConfig{
		MaxRetries: 3,
		Backoff:    []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 40 * time.Millisecond},
	}

	err := WithRetryConfig(context.Background(), "test-operation", operation, config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return status.Error(codes.InvalidArgument, "invalid argument")
	}

	err := WithRetry(context.Background(), "test-operation", operation)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retries), got %d", callCount)
	}
}

func TestWithRetry_AllRetriesExhausted(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return status.Error(codes.Unavailable, "service unavailable")
	}

	config := RetryConfig{
		MaxRetries: 2,
		Backoff:    []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}

	err := WithRetryConfig(context.Background(), "test-operation", operation, config)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	// Should be called 3 times: initial + 2 retries
	if callCount != 3 {
		t.Errorf("Expected 3 calls (initial + 2 retries), got %d", callCount)
	}
}

func TestWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	
	callCount := 0
	operation := func() error {
		callCount++
		if callCount == 1 {
			// Cancel context after first call
			cancel()
			return status.Error(codes.Unavailable, "service unavailable")
		}
		return nil
	}

	config := RetryConfig{
		MaxRetries: 3,
		Backoff:    []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond},
	}

	err := WithRetryConfig(ctx, "test-operation", operation, config)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	// Should only be called once before context cancellation
	if callCount != 1 {
		t.Errorf("Expected 1 call before cancellation, got %d", callCount)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "unavailable error",
			err:      status.Error(codes.Unavailable, "unavailable"),
			expected: true,
		},
		{
			name:     "deadline exceeded error",
			err:      status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			expected: true,
		},
		{
			name:     "invalid argument error",
			err:      status.Error(codes.InvalidArgument, "invalid argument"),
			expected: false,
		},
		{
			name:     "not found error",
			err:      status.Error(codes.NotFound, "not found"),
			expected: false,
		},
		{
			name:     "non-gRPC error",
			err:      errors.New("generic error"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
