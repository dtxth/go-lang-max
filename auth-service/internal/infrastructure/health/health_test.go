package health

import (
	"context"
	"testing"

	maxbotproto "maxbot-service/api/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// MockMaxBotClient is a mock implementation of MaxBotServiceClient for testing
type MockMaxBotClient struct {
	shouldFail bool
	response   *maxbotproto.SendNotificationResponse
}

func (m *MockMaxBotClient) SendNotification(ctx context.Context, req *maxbotproto.SendNotificationRequest, opts ...grpc.CallOption) (*maxbotproto.SendNotificationResponse, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	if m.response != nil {
		return m.response, nil
	}
	return &maxbotproto.SendNotificationResponse{
		Success: true,
	}, nil
}

// Implement other required methods (not used in health check)
func (m *MockMaxBotClient) GetMaxIDByPhone(ctx context.Context, req *maxbotproto.GetMaxIDByPhoneRequest, opts ...grpc.CallOption) (*maxbotproto.GetMaxIDByPhoneResponse, error) {
	return nil, nil
}

func (m *MockMaxBotClient) ValidatePhone(ctx context.Context, req *maxbotproto.ValidatePhoneRequest, opts ...grpc.CallOption) (*maxbotproto.ValidatePhoneResponse, error) {
	return nil, nil
}

func (m *MockMaxBotClient) SendMessage(ctx context.Context, req *maxbotproto.SendMessageRequest, opts ...grpc.CallOption) (*maxbotproto.SendMessageResponse, error) {
	return nil, nil
}

func (m *MockMaxBotClient) GetChatInfo(ctx context.Context, req *maxbotproto.GetChatInfoRequest, opts ...grpc.CallOption) (*maxbotproto.GetChatInfoResponse, error) {
	return nil, nil
}

func (m *MockMaxBotClient) GetChatMembers(ctx context.Context, req *maxbotproto.GetChatMembersRequest, opts ...grpc.CallOption) (*maxbotproto.GetChatMembersResponse, error) {
	return nil, nil
}

func (m *MockMaxBotClient) GetChatAdmins(ctx context.Context, req *maxbotproto.GetChatAdminsRequest, opts ...grpc.CallOption) (*maxbotproto.GetChatAdminsResponse, error) {
	return nil, nil
}

func (m *MockMaxBotClient) CheckPhoneNumbers(ctx context.Context, req *maxbotproto.CheckPhoneNumbersRequest, opts ...grpc.CallOption) (*maxbotproto.CheckPhoneNumbersResponse, error) {
	return nil, nil
}

func (m *MockMaxBotClient) NormalizePhone(ctx context.Context, req *maxbotproto.NormalizePhoneRequest, opts ...grpc.CallOption) (*maxbotproto.NormalizePhoneResponse, error) {
	return nil, nil
}

func (m *MockMaxBotClient) BatchGetUsersByPhone(ctx context.Context, req *maxbotproto.BatchGetUsersByPhoneRequest, opts ...grpc.CallOption) (*maxbotproto.BatchGetUsersByPhoneResponse, error) {
	return nil, nil
}

// TestHealthCheckerMaxBotHealthy tests health check when MaxBot is healthy
func TestHealthCheckerMaxBotHealthy(t *testing.T) {
	mockClient := &MockMaxBotClient{
		shouldFail: false,
		response: &maxbotproto.SendNotificationResponse{
			Success: true,
		},
	}

	checker := &HealthChecker{
		maxBotClient: mockClient,
	}

	healthy := checker.CheckMaxBotHealth(context.Background())
	assert.True(t, healthy, "MaxBot should be healthy")
}

// TestHealthCheckerMaxBotUnhealthy tests health check when MaxBot is unhealthy
func TestHealthCheckerMaxBotUnhealthy(t *testing.T) {
	mockClient := &MockMaxBotClient{
		shouldFail: true,
	}

	checker := &HealthChecker{
		maxBotClient: mockClient,
	}

	healthy := checker.CheckMaxBotHealth(context.Background())
	assert.False(t, healthy, "MaxBot should be unhealthy when request fails")
}

// TestHealthCheckerNilClient tests health check with nil client
func TestHealthCheckerNilClient(t *testing.T) {
	checker := &HealthChecker{
		maxBotClient: nil,
	}

	healthy := checker.CheckMaxBotHealth(context.Background())
	assert.False(t, healthy, "MaxBot should be unhealthy when client is nil")
}

// TestHealthCheckerMaxBotReturnsError tests health check when MaxBot returns error response
func TestHealthCheckerMaxBotReturnsError(t *testing.T) {
	mockClient := &MockMaxBotClient{
		shouldFail: false,
		response: &maxbotproto.SendNotificationResponse{
			Success: false,
			Error:   "Service unavailable",
		},
	}

	checker := &HealthChecker{
		maxBotClient: mockClient,
	}

	// Even with error response, if we got a response, service is reachable
	healthy := checker.CheckMaxBotHealth(context.Background())
	assert.True(t, healthy, "MaxBot should be considered healthy if reachable (even with error response)")
}
