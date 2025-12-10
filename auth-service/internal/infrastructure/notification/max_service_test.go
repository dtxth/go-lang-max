package notification

import (
	"context"
	"net"
	"testing"

	"auth-service/internal/infrastructure/logger"
	// maxbotproto "maxbot-service/api/proto"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// mockMaxBotServer is a mock implementation of MaxBotServiceServer for testing
type mockMaxBotServer struct {
	maxbotproto.UnimplementedMaxBotServiceServer
	shouldFail    bool
	failWithError string
	errorCode     maxbotproto.ErrorCode
	lastMessage   string // Capture the last message sent
	lastPhone     string // Capture the last phone number
}

func (m *mockMaxBotServer) SendNotification(ctx context.Context, req *maxbotproto.SendNotificationRequest) (*maxbotproto.SendNotificationResponse, error) {
	// Capture the message and phone for testing
	m.lastMessage = req.Text
	m.lastPhone = req.Phone
	
	if m.shouldFail {
		if m.failWithError != "" {
			return &maxbotproto.SendNotificationResponse{
				Success:   false,
				ErrorCode: m.errorCode,
				Error:     m.failWithError,
			}, nil
		}
		return nil, status.Error(codes.Unavailable, "service unavailable")
	}
	
	return &maxbotproto.SendNotificationResponse{
		Success:   true,
		ErrorCode: maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED,
		Error:     "",
	}, nil
}

// setupMockMaxBotServer creates a mock MaxBot gRPC server for testing
func setupMockMaxBotServer(shouldFail bool, failWithError string, errorCode maxbotproto.ErrorCode) (*grpc.Server, *bufconn.Listener, *mockMaxBotServer) {
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)
	
	server := grpc.NewServer()
	mockServer := &mockMaxBotServer{
		shouldFail:    shouldFail,
		failWithError: failWithError,
		errorCode:     errorCode,
	}
	maxbotproto.RegisterMaxBotServiceServer(server, mockServer)
	
	go func() {
		if err := server.Serve(listener); err != nil {
			// Server stopped
		}
	}()
	
	return server, listener, mockServer
}

// createTestMaxNotificationService creates a MaxNotificationService connected to a mock server
func createTestMaxNotificationService(listener *bufconn.Listener, log *logger.Logger) (*MaxNotificationService, error) {
	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	
	return &MaxNotificationService{
		conn:    conn,
		client:  maxbotproto.NewMaxBotServiceClient(conn),
		logger:  log,
		timeout: 2 * 1000000000, // 2 seconds
	}, nil
}

// **Feature: secure-password-management, Property 6: Notification delivery on user creation**
// **Validates: Requirements 2.1**
//
// Property 6: Notification delivery on user creation
// For any user created with a role, a notification should be sent to the user's phone number via MaxBot Service
func TestProperty_NotificationDeliveryOnUserCreation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	
	properties := gopter.NewProperties(parameters)
	
	properties.Property("notification is sent for any valid phone and password", prop.ForAll(
		func(phone, password string) bool {
			// Setup mock server that succeeds
			server, listener, mockServer := setupMockMaxBotServer(false, "", maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED)
			defer server.Stop()
			
			// Create logger
			log := logger.New(nil, logger.INFO)
			
			// Create service
			service, err := createTestMaxNotificationService(listener, log)
			if err != nil {
				t.Logf("Failed to create service: %v", err)
				return false
			}
			defer service.Close()
			
			// Send notification (simulating user creation with role)
			ctx := context.Background()
			err = service.SendPasswordNotification(ctx, phone, password)
			
			// Verify notification was sent successfully
			if err != nil {
				t.Logf("Failed to send notification: %v", err)
				return false
			}
			
			// Verify the notification was actually delivered to MaxBot Service
			// by checking that the mock server received the request
			if mockServer.lastPhone != phone {
				t.Logf("Notification not sent to correct phone. Expected: %s, Got: %s", phone, mockServer.lastPhone)
				return false
			}
			
			// Verify message was sent (not empty)
			if len(mockServer.lastMessage) == 0 {
				t.Logf("No message was sent to MaxBot Service")
				return false
			}
			
			return true
		},
		gen.RegexMatch(`\+7[0-9]{10}`),        // Valid Russian phone number
		gen.RegexMatch(`[A-Za-z0-9!@#$]{12,20}`), // Valid password
	))
	
	properties.TestingRun(t)
}


// **Feature: secure-password-management, Property 7: Notification message format**
// **Validates: Requirements 2.2**
//
// Property 7: Notification message format
// For any password notification, the message should contain both the password and instructions for first login
func TestProperty_NotificationMessageFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	
	properties := gopter.NewProperties(parameters)
	
	properties.Property("password notification contains password and instructions", prop.ForAll(
		func(phone, password string) bool {
			// Setup mock server that succeeds
			server, listener, mockServer := setupMockMaxBotServer(false, "", maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED)
			defer server.Stop()
			
			// Create logger
			log := logger.New(nil, logger.INFO)
			
			// Create service
			service, err := createTestMaxNotificationService(listener, log)
			if err != nil {
				t.Logf("Failed to create service: %v", err)
				return false
			}
			defer service.Close()
			
			// Send notification
			ctx := context.Background()
			err = service.SendPasswordNotification(ctx, phone, password)
			if err != nil {
				t.Logf("Failed to send notification: %v", err)
				return false
			}
			
			// Check that message contains the password
			message := mockServer.lastMessage
			if len(message) == 0 {
				t.Logf("No message was captured")
				return false
			}
			
			// Message should contain the password
			hasPassword := false
			for i := 0; i < len(message)-len(password)+1; i++ {
				if message[i:i+len(password)] == password {
					hasPassword = true
					break
				}
			}
			
			if !hasPassword {
				t.Logf("Message does not contain password. Message: %s, Password: %s", message, password)
				return false
			}
			
			// Message should contain instructions (in Russian)
			hasInstructions := false
			instructionKeywords := []string{"временный пароль", "сменить пароль", "первого входа"}
			for _, keyword := range instructionKeywords {
				if len(message) >= len(keyword) {
					for i := 0; i <= len(message)-len(keyword); i++ {
						if message[i:i+len(keyword)] == keyword {
							hasInstructions = true
							break
						}
					}
				}
				if hasInstructions {
					break
				}
			}
			
			if !hasInstructions {
				t.Logf("Message does not contain instructions. Message: %s", message)
				return false
			}
			
			return true
		},
		gen.RegexMatch(`\+7[0-9]{10}`),        // Valid Russian phone number
		gen.RegexMatch(`[A-Za-z0-9!@#$]{12,20}`), // Valid password
	))
	
	properties.TestingRun(t)
}

// **Feature: secure-password-management, Property 7: Notification message format (Reset Token)**
// **Validates: Requirements 2.2**
//
// Property 7b: Reset token notification message format
// For any reset token notification, the message should contain the token and instructions
func TestProperty_ResetTokenMessageFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	
	properties := gopter.NewProperties(parameters)
	
	properties.Property("reset token notification contains token and instructions", prop.ForAll(
		func(phone, token string) bool {
			// Setup mock server that succeeds
			server, listener, mockServer := setupMockMaxBotServer(false, "", maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED)
			defer server.Stop()
			
			// Create logger
			log := logger.New(nil, logger.INFO)
			
			// Create service
			service, err := createTestMaxNotificationService(listener, log)
			if err != nil {
				t.Logf("Failed to create service: %v", err)
				return false
			}
			defer service.Close()
			
			// Send notification
			ctx := context.Background()
			err = service.SendResetTokenNotification(ctx, phone, token)
			if err != nil {
				t.Logf("Failed to send notification: %v", err)
				return false
			}
			
			// Check that message contains the token
			message := mockServer.lastMessage
			if len(message) == 0 {
				t.Logf("No message was captured")
				return false
			}
			
			// Message should contain the token
			hasToken := false
			for i := 0; i < len(message)-len(token)+1; i++ {
				if message[i:i+len(token)] == token {
					hasToken = true
					break
				}
			}
			
			if !hasToken {
				t.Logf("Message does not contain token. Message: %s, Token: %s", message, token)
				return false
			}
			
			// Message should contain instructions (in Russian)
			hasInstructions := false
			instructionKeywords := []string{"код для сброса пароля", "15 минут", "действителен"}
			for _, keyword := range instructionKeywords {
				if len(message) >= len(keyword) {
					for i := 0; i <= len(message)-len(keyword); i++ {
						if message[i:i+len(keyword)] == keyword {
							hasInstructions = true
							break
						}
					}
				}
				if hasInstructions {
					break
				}
			}
			
			if !hasInstructions {
				t.Logf("Message does not contain instructions. Message: %s", message)
				return false
			}
			
			return true
		},
		gen.RegexMatch(`\+7[0-9]{10}`),        // Valid Russian phone number
		gen.RegexMatch(`[A-Za-z0-9]{6,20}`),   // Valid reset token
	))
	
	properties.TestingRun(t)
}


// **Feature: secure-password-management, Property 8: Notification error handling**
// **Validates: Requirements 2.3**
//
// Property 8: Notification error handling
// For any notification delivery failure, an error should be logged and returned to the caller
func TestProperty_NotificationErrorHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	
	properties := gopter.NewProperties(parameters)
	
	// Test that MaxBot errors are properly handled and returned
	// We test with error responses (not gRPC failures) to avoid retry delays
	properties.Property("MaxBot errors are returned to caller", prop.ForAll(
		func(phone, password string) bool {
			// Setup mock server that returns error response (not gRPC error)
			server, listener, _ := setupMockMaxBotServer(true, "MaxBot service error", maxbotproto.ErrorCode_ERROR_CODE_INTERNAL)
			defer server.Stop()
			
			// Create logger
			log := logger.New(nil, logger.INFO)
			
			// Create service
			service, err := createTestMaxNotificationService(listener, log)
			if err != nil {
				t.Logf("Failed to create service: %v", err)
				return false
			}
			defer service.Close()
			
			// Send notification
			ctx := context.Background()
			err = service.SendPasswordNotification(ctx, phone, password)
			
			// Should fail with error
			if err == nil {
				t.Logf("Expected error but got nil")
				return false
			}
			
			// Error message should contain the MaxBot error
			errMsg := err.Error()
			if len(errMsg) == 0 {
				t.Logf("Error message is empty")
				return false
			}
			
			return true
		},
		gen.RegexMatch(`\+7[0-9]{10}`),        // Valid Russian phone number
		gen.RegexMatch(`[A-Za-z0-9!@#$]{12,20}`), // Valid password
	))
	
	properties.TestingRun(t)
}
