package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"gateway-service/internal/config"
	"gateway-service/internal/infrastructure/errors"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries        int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	Jitter            bool
}

// Retrier implements retry logic with exponential backoff
type Retrier struct {
	config       RetryConfig
	errorHandler *errors.ErrorHandler
}

// NewRetrier creates a new retrier
func NewRetrier(cfg config.ServiceConfig, errorHandler *errors.ErrorHandler) *Retrier {
	return &Retrier{
		config: RetryConfig{
			MaxRetries:        cfg.MaxRetries,
			InitialDelay:      cfg.RetryDelay,
			MaxDelay:          cfg.MaxRetryDelay,
			BackoffMultiplier: cfg.BackoffMultiplier,
			Jitter:            true, // Add jitter to prevent thundering herd
		},
		errorHandler: errorHandler,
	}
}

// Execute executes a function with retry logic
func (r *Retrier) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error
	
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}
		
		lastErr = err
		
		// Don't retry on the last attempt
		if attempt == r.config.MaxRetries {
			break
		}
		
		// Check if error is retryable
		if !r.errorHandler.IsRetryableError(err) {
			return err // Don't retry non-retryable errors
		}
		
		// Calculate delay for next attempt
		delay := r.calculateDelay(attempt)
		
		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	return fmt.Errorf("max retries (%d) exceeded, last error: %w", r.config.MaxRetries, lastErr)
}

// calculateDelay calculates the delay for the given attempt using exponential backoff
func (r *Retrier) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff delay
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.BackoffMultiplier, float64(attempt))
	
	// Apply maximum delay limit
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}
	
	// Add jitter to prevent thundering herd problem
	if r.config.Jitter {
		jitter := rand.Float64() * 0.1 * delay // Up to 10% jitter
		delay += jitter
	}
	
	return time.Duration(delay)
}

// ExecuteWithCircuitBreaker executes a function with both retry and circuit breaker
func (r *Retrier) ExecuteWithCircuitBreaker(ctx context.Context, circuitBreaker CircuitBreaker, fn func(ctx context.Context) error) error {
	return r.Execute(ctx, func(ctx context.Context) error {
		return circuitBreaker.Execute(ctx, fn)
	})
}

// CircuitBreaker interface for circuit breaker integration
type CircuitBreaker interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}