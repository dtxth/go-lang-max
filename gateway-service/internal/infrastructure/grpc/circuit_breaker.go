package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway-service/internal/config"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateHalfOpen
	StateOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for gRPC calls
type CircuitBreaker struct {
	config config.CircuitBreakerConfig
	
	mu           sync.RWMutex
	state        CircuitBreakerState
	counts       map[string]uint64
	expiry       time.Time
	generation   uint64
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(cfg config.CircuitBreakerConfig) *CircuitBreaker {
	cb := &CircuitBreaker{
		config: cfg,
		state:  StateClosed,
		counts: make(map[string]uint64),
	}
	
	// Set default ReadyToTrip function if not provided
	if cb.config.ReadyToTrip == nil {
		cb.config.ReadyToTrip = func(counts map[string]uint64) bool {
			// Default: trip if failure rate > 50% and at least 5 requests
			total := counts["total"]
			failures := counts["failures"]
			
			if total < 5 {
				return false
			}
			
			failureRate := float64(failures) / float64(total)
			return failureRate > 0.5
		}
	}
	
	return cb
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}
	
	defer func() {
		if r := recover(); r != nil {
			cb.afterRequest(generation, false)
			panic(r)
		}
	}()
	
	err = fn(ctx)
	cb.afterRequest(generation, err == nil)
	return err
}

// beforeRequest is called before executing a request
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	state, generation := cb.currentState(now)
	
	if state == StateOpen {
		return generation, fmt.Errorf("circuit breaker is open")
	} else if state == StateHalfOpen {
		if cb.counts["requests"] >= uint64(cb.config.MaxRequests) {
			return generation, fmt.Errorf("circuit breaker is half-open and max requests exceeded")
		}
	}
	
	cb.counts["requests"]++
	cb.counts["total"]++
	
	return generation, nil
}

// afterRequest is called after executing a request
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	state, generation := cb.currentState(now)
	
	if generation != before {
		return // Different generation, ignore
	}
	
	if success {
		cb.onSuccess(state)
	} else {
		cb.onFailure(state)
	}
}

// onSuccess handles successful requests
func (cb *CircuitBreaker) onSuccess(state CircuitBreakerState) {
	cb.counts["successes"]++
	
	if state == StateHalfOpen {
		// If we're half-open and got enough successes, close the circuit
		if cb.counts["successes"] >= uint64(cb.config.MaxRequests) {
			cb.setState(StateClosed, time.Now())
		}
	}
}

// onFailure handles failed requests
func (cb *CircuitBreaker) onFailure(state CircuitBreakerState) {
	cb.counts["failures"]++
	
	switch state {
	case StateClosed:
		if cb.config.ReadyToTrip(cb.counts) {
			cb.setState(StateOpen, time.Now())
		}
	case StateHalfOpen:
		cb.setState(StateOpen, time.Now())
	}
}

// currentState returns the current state and generation
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitBreakerState, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	
	return cb.state, cb.generation
}

// setState changes the state of the circuit breaker
func (cb *CircuitBreaker) setState(state CircuitBreakerState, now time.Time) {
	if cb.state == state {
		return
	}
	
	prev := cb.state
	cb.state = state
	cb.toNewGeneration(now)
	
	switch state {
	case StateClosed:
		// No expiry for closed state
		cb.expiry = time.Time{}
	case StateOpen:
		// Set expiry for open state
		cb.expiry = now.Add(cb.config.Timeout)
	case StateHalfOpen:
		// Set expiry for half-open state
		cb.expiry = now.Add(cb.config.Timeout)
	}
	
	// Log state change
	fmt.Printf("Circuit breaker state changed from %s to %s\n", prev.String(), state.String())
}

// toNewGeneration resets counters and increments generation
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = make(map[string]uint64)
	
	var expiry time.Time
	switch cb.state {
	case StateClosed:
		expiry = now.Add(cb.config.Interval)
	case StateOpen:
		expiry = now.Add(cb.config.Timeout)
	default:
		expiry = now.Add(cb.config.Interval)
	}
	cb.expiry = expiry
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	state, _ := cb.currentState(time.Now())
	return state
}

// Counts returns a copy of the current counts
func (cb *CircuitBreaker) Counts() map[string]uint64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	result := make(map[string]uint64)
	for k, v := range cb.counts {
		result[k] = v
	}
	return result
}