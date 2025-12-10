package health

import (
	"context"
	"time"

	// maxbotproto "maxbot-service/api/proto"
	"google.golang.org/grpc"
)

// HealthChecker performs health checks on external services
type HealthChecker struct {
	maxBotClient maxbotproto.MaxBotServiceClient
	timeout      time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(conn *grpc.ClientConn) *HealthChecker {
	return &HealthChecker{
		maxBotClient: maxbotproto.NewMaxBotServiceClient(conn),
		timeout:      5 * time.Second,
	}
}

// CheckMaxBotHealth checks if MaxBot service is healthy
func (h *HealthChecker) CheckMaxBotHealth(ctx context.Context) bool {
	if h.maxBotClient == nil {
		return false
	}
	
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	
	// Try to send a test notification to a dummy phone number
	// We use a special format that MaxBot should recognize as a health check
	resp, err := h.maxBotClient.SendNotification(timeoutCtx, &maxbotproto.SendNotificationRequest{
		Phone: "+70000000000", // Health check phone number
		Text:  "HEALTH_CHECK",
	})
	
	if err != nil {
		return false
	}
	
	// Consider service healthy if we got a response (even if it's an error response)
	// The important thing is that the service is reachable
	return resp != nil
}
