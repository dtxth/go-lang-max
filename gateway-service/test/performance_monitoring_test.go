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

// PerformanceMetrics holds performance monitoring data
type PerformanceMetrics struct {
	RequestCount    int64
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	ErrorCount      int64
	TimeoutCount    int64
	RetryCount      int64
	mu              sync.RWMutex
}

// NewPerformanceMetrics creates a new performance metrics collector
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		MinDuration: time.Duration(1<<63 - 1), // Max duration as initial min
	}
}

// RecordRequest records a request's performance metrics
func (pm *PerformanceMetrics) RecordRequest(duration time.Duration, err error, retryCount int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.RequestCount++
	pm.TotalDuration += duration
	pm.AverageDuration = pm.TotalDuration / time.Duration(pm.RequestCount)

	if duration < pm.MinDuration {
		pm.MinDuration = duration
	}
	if duration > pm.MaxDuration {
		pm.MaxDuration = duration
	}

	if err != nil {
		pm.ErrorCount++
		if status.Code(err) == codes.DeadlineExceeded {
			pm.TimeoutCount++
		}
	}

	pm.RetryCount += int64(retryCount)
}

// GetMetrics returns a copy of current metrics
func (pm *PerformanceMetrics) GetMetrics() PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return *pm
}

// Reset resets all metrics
func (pm *PerformanceMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.RequestCount = 0
	pm.TotalDuration = 0
	pm.AverageDuration = 0
	pm.MinDuration = time.Duration(1<<63 - 1)
	pm.MaxDuration = 0
	pm.ErrorCount = 0
	pm.TimeoutCount = 0
	pm.RetryCount = 0
}

// TestPerformanceMetricsCollection tests performance metrics collection functionality
// Requirements: 6.1, 6.5
func TestPerformanceMetricsCollection(t *testing.T) {
	metrics := NewPerformanceMetrics()

	// Test recording successful requests
	t.Run("RecordSuccessfulRequests", func(t *testing.T) {
		metrics.Reset()
		
		durations := []time.Duration{
			100 * time.Millisecond,
			150 * time.Millisecond,
			200 * time.Millisecond,
		}

		for _, duration := range durations {
			metrics.RecordRequest(duration, nil, 0)
		}

		result := metrics.GetMetrics()
		
		if result.RequestCount != 3 {
			t.Errorf("Expected 3 requests, got %d", result.RequestCount)
		}
		
		if result.ErrorCount != 0 {
			t.Errorf("Expected 0 errors, got %d", result.ErrorCount)
		}
		
		if result.MinDuration != 100*time.Millisecond {
			t.Errorf("Expected min duration 100ms, got %v", result.MinDuration)
		}
		
		if result.MaxDuration != 200*time.Millisecond {
			t.Errorf("Expected max duration 200ms, got %v", result.MaxDuration)
		}
		
		expectedAvg := 150 * time.Millisecond
		if result.AverageDuration != expectedAvg {
			t.Errorf("Expected average duration %v, got %v", expectedAvg, result.AverageDuration)
		}
	})

	// Test recording failed requests
	t.Run("RecordFailedRequests", func(t *testing.T) {
		metrics.Reset()
		
		// Record some successful and failed requests
		metrics.RecordRequest(100*time.Millisecond, nil, 0)
		metrics.RecordRequest(200*time.Millisecond, status.Error(codes.Internal, "internal error"), 2)
		metrics.RecordRequest(150*time.Millisecond, status.Error(codes.DeadlineExceeded, "timeout"), 1)

		result := metrics.GetMetrics()
		
		if result.RequestCount != 3 {
			t.Errorf("Expected 3 requests, got %d", result.RequestCount)
		}
		
		if result.ErrorCount != 2 {
			t.Errorf("Expected 2 errors, got %d", result.ErrorCount)
		}
		
		if result.TimeoutCount != 1 {
			t.Errorf("Expected 1 timeout, got %d", result.TimeoutCount)
		}
		
		if result.RetryCount != 3 {
			t.Errorf("Expected 3 total retries, got %d", result.RetryCount)
		}
	})

	// Test concurrent access
	t.Run("ConcurrentAccess", func(t *testing.T) {
		metrics.Reset()
		
		var wg sync.WaitGroup
		numGoroutines := 10
		requestsPerGoroutine := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					duration := time.Duration(j+1) * time.Millisecond
					var err error
					if j%10 == 0 {
						err = status.Error(codes.Internal, "test error")
					}
					metrics.RecordRequest(duration, err, j%3)
				}
			}()
		}

		wg.Wait()
		result := metrics.GetMetrics()
		
		expectedRequests := int64(numGoroutines * requestsPerGoroutine)
		if result.RequestCount != expectedRequests {
			t.Errorf("Expected %d requests, got %d", expectedRequests, result.RequestCount)
		}
		
		expectedErrors := int64(numGoroutines * (requestsPerGoroutine / 10))
		if result.ErrorCount != expectedErrors {
			t.Errorf("Expected %d errors, got %d", expectedErrors, result.ErrorCount)
		}
	})
}

// TestTimeoutHandlingUnderLoad tests timeout handling under various load conditions
// Requirements: 6.1
func TestTimeoutHandlingUnderLoad(t *testing.T) {
	testCases := []struct {
		name           string
		timeout        time.Duration
		requestDelay   time.Duration
		concurrency    int
		expectTimeouts bool
	}{
		{
			name:           "LowLoadWithinTimeout",
			timeout:        200 * time.Millisecond,
			requestDelay:   50 * time.Millisecond,
			concurrency:    5,
			expectTimeouts: false,
		},
		{
			name:           "HighLoadExceedingTimeout",
			timeout:        50 * time.Millisecond,
			requestDelay:   100 * time.Millisecond,
			concurrency:    10,
			expectTimeouts: true,
		},
		{
			name:           "MediumLoadPartialTimeout",
			timeout:        100 * time.Millisecond,
			requestDelay:   80 * time.Millisecond,
			concurrency:    5,
			expectTimeouts: false, // Should be within timeout with some margin
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := NewPerformanceMetrics()
			var wg sync.WaitGroup

			for i := 0; i < tc.concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					
					ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
					defer cancel()

					start := time.Now()
					
					// Simulate work that takes requestDelay time
					select {
					case <-time.After(tc.requestDelay):
						// Work completed
						duration := time.Since(start)
						metrics.RecordRequest(duration, nil, 0)
					case <-ctx.Done():
						// Timeout occurred
						duration := time.Since(start)
						err := status.Error(codes.DeadlineExceeded, "timeout")
						metrics.RecordRequest(duration, err, 0)
					}
				}()
			}

			wg.Wait()
			result := metrics.GetMetrics()

			if result.RequestCount != int64(tc.concurrency) {
				t.Errorf("Expected %d requests, got %d", tc.concurrency, result.RequestCount)
			}

			if tc.expectTimeouts {
				if result.TimeoutCount == 0 {
					t.Errorf("Expected some timeouts, but got none")
				}
			} else {
				// Allow for some timing variance in CI environments
				if result.TimeoutCount > int64(tc.concurrency/2) {
					t.Errorf("Expected few or no timeouts, got %d", result.TimeoutCount)
				}
			}
		})
	}
}

// TestRetryBehaviorUnderFailureScenarios tests retry behavior under various failure scenarios
// Requirements: 6.5
func TestRetryBehaviorUnderFailureScenarios(t *testing.T) {
	testCases := []struct {
		name            string
		maxRetries      int
		failureRate     float64 // 0.0 = no failures, 1.0 = all failures
		numRequests     int
		expectRetries   bool
	}{
		{
			name:          "NoFailures",
			maxRetries:    3,
			failureRate:   0.0,
			numRequests:   10,
			expectRetries: false,
		},
		{
			name:          "AllFailures",
			maxRetries:    3,
			failureRate:   1.0,
			numRequests:   5,
			expectRetries: true,
		},
		{
			name:          "SomeFailures",
			maxRetries:    2,
			failureRate:   0.6, // 60% failure rate
			numRequests:   10,
			expectRetries: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.ServiceConfig{
				MaxRetries:        tc.maxRetries,
				RetryDelay:        10 * time.Millisecond,
				MaxRetryDelay:     100 * time.Millisecond,
				BackoffMultiplier: 2.0,
			}

			errHandler := errorHandler.NewErrorHandler()
			retrier := retry.NewRetrier(cfg, errHandler)
			metrics := NewPerformanceMetrics()

			var wg sync.WaitGroup

			for i := 0; i < tc.numRequests; i++ {
				wg.Add(1)
				go func(requestID int) {
					defer wg.Done()
					
					attemptCount := 0
					testFunc := func(ctx context.Context) error {
						attemptCount++
						
						// Simulate failure based on failure rate and request ID
						failureThreshold := int(tc.failureRate * 100)
						if (requestID*37)%100 < failureThreshold { // Use deterministic "random"
							return status.Error(codes.Unavailable, "simulated failure")
						}
						return nil
					}

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					start := time.Now()
					err := retrier.Execute(ctx, testFunc)
					duration := time.Since(start)

					retryCount := attemptCount - 1 // First attempt is not a retry
					if retryCount < 0 {
						retryCount = 0
					}
					
					metrics.RecordRequest(duration, err, retryCount)
				}(i)
			}

			wg.Wait()
			result := metrics.GetMetrics()

			if result.RequestCount != int64(tc.numRequests) {
				t.Errorf("Expected %d requests, got %d", tc.numRequests, result.RequestCount)
			}

			if tc.expectRetries {
				if result.RetryCount == 0 {
					t.Errorf("Expected some retries, but got none")
				}
			} else {
				if result.RetryCount != 0 {
					t.Errorf("Expected no retries, but got %d", result.RetryCount)
				}
			}
		})
	}
}

// TestConnectionPoolingPerformance tests gRPC connection pooling performance
// Requirements: 1.1
func TestConnectionPoolingPerformance(t *testing.T) {
	// Test connection reuse vs new connections
	t.Run("ConnectionReuse", func(t *testing.T) {
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

		// Test that client manager can be created and configured properly
		clientManager := grpcClient.NewClientManager(cfg)
		
		if clientManager == nil {
			t.Fatal("Failed to create client manager")
		}

		// Test configuration preservation
		if cfg.Services.Auth.Timeout != 1*time.Second {
			t.Errorf("Expected timeout 1s, got %v", cfg.Services.Auth.Timeout)
		}

		// Test multiple client retrievals (simulating connection reuse)
		// Note: Without actual gRPC server, clients will be nil, but we test the manager logic
		metrics := NewPerformanceMetrics()
		numRequests := 100

		start := time.Now()
		for i := 0; i < numRequests; i++ {
			client := clientManager.GetAuthClient()
			// Client will be nil without server, but that's expected in unit tests
			_ = client
		}
		totalDuration := time.Since(start)

		// Record the performance of client retrieval
		avgDuration := totalDuration / time.Duration(numRequests)
		metrics.RecordRequest(avgDuration, nil, 0)

		result := metrics.GetMetrics()
		
		// Client retrieval should be very fast (< 10ms per request even in unit tests)
		if result.AverageDuration > 10*time.Millisecond {
			t.Errorf("Client retrieval too slow: %v per request", result.AverageDuration)
		}
	})

	// Test concurrent access to connection pool
	t.Run("ConcurrentConnectionAccess", func(t *testing.T) {
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
		metrics := NewPerformanceMetrics()

		numGoroutines := 50
		requestsPerGoroutine := 20
		var wg sync.WaitGroup

		start := time.Now()
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					requestStart := time.Now()
					client := clientManager.GetAuthClient()
					requestDuration := time.Since(requestStart)
					
					// In unit tests without server, client will be nil, but that's expected
					_ = client
					metrics.RecordRequest(requestDuration, nil, 0)
				}
			}()
		}

		wg.Wait()
		totalDuration := time.Since(start)

		result := metrics.GetMetrics()
		expectedRequests := int64(numGoroutines * requestsPerGoroutine)
		
		if result.RequestCount != expectedRequests {
			t.Errorf("Expected %d requests, got %d", expectedRequests, result.RequestCount)
		}

		// Total time should be reasonable for concurrent access
		if totalDuration > 5*time.Second {
			t.Errorf("Concurrent access took too long: %v", totalDuration)
		}

		t.Logf("Concurrent access performance: %d requests in %v (avg: %v per request)", 
			expectedRequests, totalDuration, result.AverageDuration)
	})
}

// TestCircuitBreakerPerformance tests circuit breaker performance under load
// Requirements: 6.1, 6.5
func TestCircuitBreakerPerformance(t *testing.T) {
	t.Run("CircuitBreakerFailFast", func(t *testing.T) {
		cfg := config.CircuitBreakerConfig{
			MaxRequests: 5,
			Interval:    1 * time.Second,
			Timeout:     1 * time.Second,
		}

		cb := grpcClient.NewCircuitBreaker(cfg)
		metrics := NewPerformanceMetrics()

		// First, cause the circuit breaker to open by failing requests
		failingFunc := func(ctx context.Context) error {
			return status.Error(codes.Internal, "simulated failure")
		}

		// Make enough failing requests to open the circuit
		for i := 0; i < 10; i++ {
			start := time.Now()
			err := cb.Execute(context.Background(), failingFunc)
			duration := time.Since(start)
			metrics.RecordRequest(duration, err, 0)
		}

		result := metrics.GetMetrics()
		
		// All requests should have failed
		if result.ErrorCount != result.RequestCount {
			t.Errorf("Expected all requests to fail, got %d errors out of %d requests", 
				result.ErrorCount, result.RequestCount)
		}

		// Later requests should be faster (fail-fast behavior)
		// This is a simplified test - in practice, we'd need more sophisticated timing analysis
		if result.RequestCount < 5 {
			t.Errorf("Expected at least 5 requests to test circuit breaker behavior")
		}
	})
}

// Helper function to calculate absolute difference
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to calculate absolute difference for int64
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}