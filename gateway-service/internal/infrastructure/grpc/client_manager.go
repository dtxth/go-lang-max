package grpc

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	authpb "auth-service/api/proto"
	chatpb "chat-service/api/proto"
	employeepb "employee-service/api/proto"
	"gateway-service/internal/config"
	"gateway-service/internal/infrastructure/errors"
	"gateway-service/internal/infrastructure/retry"
	structurepb "structure-service/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientManager manages gRPC client connections to all microservices
type ClientManager struct {
	config *config.Config
	
	// gRPC connections
	authConn      *grpc.ClientConn
	chatConn      *grpc.ClientConn
	employeeConn  *grpc.ClientConn
	structureConn *grpc.ClientConn
	
	// gRPC clients
	authClient      authpb.AuthServiceClient
	chatClient      chatpb.ChatServiceClient
	employeeClient  employeepb.EmployeeServiceClient
	structureClient structurepb.StructureServiceClient
	
	// Circuit breakers for each service
	authCircuitBreaker      *CircuitBreaker
	chatCircuitBreaker      *CircuitBreaker
	employeeCircuitBreaker  *CircuitBreaker
	structureCircuitBreaker *CircuitBreaker
	
	// Retry handlers for each service
	authRetrier      *retry.Retrier
	chatRetrier      *retry.Retrier
	employeeRetrier  *retry.Retrier
	structureRetrier *retry.Retrier
	
	// Error handler
	errorHandler *errors.ErrorHandler
	
	// Connection management
	mu     sync.RWMutex
	closed bool
}

// NewClientManager creates a new gRPC client manager
func NewClientManager(cfg *config.Config) *ClientManager {
	errorHandler := errors.NewErrorHandler()
	
	return &ClientManager{
		config:       cfg,
		errorHandler: errorHandler,
		
		// Initialize circuit breakers
		authCircuitBreaker:      NewCircuitBreaker(cfg.Services.Auth.CircuitBreaker),
		chatCircuitBreaker:      NewCircuitBreaker(cfg.Services.Chat.CircuitBreaker),
		employeeCircuitBreaker:  NewCircuitBreaker(cfg.Services.Employee.CircuitBreaker),
		structureCircuitBreaker: NewCircuitBreaker(cfg.Services.Structure.CircuitBreaker),
		
		// Initialize retry handlers
		authRetrier:      retry.NewRetrier(cfg.Services.Auth, errorHandler),
		chatRetrier:      retry.NewRetrier(cfg.Services.Chat, errorHandler),
		employeeRetrier:  retry.NewRetrier(cfg.Services.Employee, errorHandler),
		structureRetrier: retry.NewRetrier(cfg.Services.Structure, errorHandler),
	}
}

// SetLogger sets the logger for the error handler
func (cm *ClientManager) SetLogger(logger errors.Logger) {
	cm.errorHandler = errors.NewErrorHandlerWithLogger(logger)
	
	// Update retry handlers with new error handler
	cm.authRetrier = retry.NewRetrier(cm.config.Services.Auth, cm.errorHandler)
	cm.chatRetrier = retry.NewRetrier(cm.config.Services.Chat, cm.errorHandler)
	cm.employeeRetrier = retry.NewRetrier(cm.config.Services.Employee, cm.errorHandler)
	cm.structureRetrier = retry.NewRetrier(cm.config.Services.Structure, cm.errorHandler)
}

// Start initializes all gRPC client connections
func (cm *ClientManager) Start(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if cm.closed {
		return fmt.Errorf("client manager is closed")
	}
	
	var errors []string
	
	// Initialize Auth Service connection
	if err := cm.initAuthClient(ctx); err != nil {
		errors = append(errors, fmt.Sprintf("auth: %v", err))
		log.Printf("Warning: Failed to initialize auth client: %v", err)
	}
	
	// Initialize Chat Service connection
	if err := cm.initChatClient(ctx); err != nil {
		errors = append(errors, fmt.Sprintf("chat: %v", err))
		log.Printf("Warning: Failed to initialize chat client: %v", err)
	}
	
	// Initialize Employee Service connection
	if err := cm.initEmployeeClient(ctx); err != nil {
		errors = append(errors, fmt.Sprintf("employee: %v", err))
		log.Printf("Warning: Failed to initialize employee client: %v", err)
	}
	
	// Initialize Structure Service connection
	if err := cm.initStructureClient(ctx); err != nil {
		errors = append(errors, fmt.Sprintf("structure: %v", err))
		log.Printf("Warning: Failed to initialize structure client: %v", err)
	}
	
	if len(errors) > 0 {
		log.Printf("Some gRPC clients failed to initialize: %v", errors)
		return fmt.Errorf("failed to initialize some clients: %v", errors)
	}
	
	log.Println("All gRPC client connections initialized successfully")
	return nil
}

// Stop closes all gRPC client connections
func (cm *ClientManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if cm.closed {
		return nil
	}
	
	var errors []error
	
	if cm.authConn != nil {
		if err := cm.authConn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close auth connection: %w", err))
		}
	}
	
	if cm.chatConn != nil {
		if err := cm.chatConn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close chat connection: %w", err))
		}
	}
	
	if cm.employeeConn != nil {
		if err := cm.employeeConn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close employee connection: %w", err))
		}
	}
	
	if cm.structureConn != nil {
		if err := cm.structureConn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close structure connection: %w", err))
		}
	}
	
	cm.closed = true
	
	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}
	
	log.Println("All gRPC client connections closed successfully")
	return nil
}

// GetAuthClient returns the Auth Service gRPC client
func (cm *ClientManager) GetAuthClient() authpb.AuthServiceClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.authClient
}

// GetChatClient returns the Chat Service gRPC client
func (cm *ClientManager) GetChatClient() chatpb.ChatServiceClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.chatClient
}

// GetEmployeeClient returns the Employee Service gRPC client
func (cm *ClientManager) GetEmployeeClient() employeepb.EmployeeServiceClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.employeeClient
}

// GetStructureClient returns the Structure Service gRPC client
func (cm *ClientManager) GetStructureClient() structurepb.StructureServiceClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.structureClient
}

// GetAuthCircuitBreaker returns the Auth Service circuit breaker
func (cm *ClientManager) GetAuthCircuitBreaker() *CircuitBreaker {
	return cm.authCircuitBreaker
}

// GetChatCircuitBreaker returns the Chat Service circuit breaker
func (cm *ClientManager) GetChatCircuitBreaker() *CircuitBreaker {
	return cm.chatCircuitBreaker
}

// GetEmployeeCircuitBreaker returns the Employee Service circuit breaker
func (cm *ClientManager) GetEmployeeCircuitBreaker() *CircuitBreaker {
	return cm.employeeCircuitBreaker
}

// GetStructureCircuitBreaker returns the Structure Service circuit breaker
func (cm *ClientManager) GetStructureCircuitBreaker() *CircuitBreaker {
	return cm.structureCircuitBreaker
}

// GetAuthRetrier returns the Auth Service retrier
func (cm *ClientManager) GetAuthRetrier() *retry.Retrier {
	return cm.authRetrier
}

// GetChatRetrier returns the Chat Service retrier
func (cm *ClientManager) GetChatRetrier() *retry.Retrier {
	return cm.chatRetrier
}

// GetEmployeeRetrier returns the Employee Service retrier
func (cm *ClientManager) GetEmployeeRetrier() *retry.Retrier {
	return cm.employeeRetrier
}

// GetStructureRetrier returns the Structure Service retrier
func (cm *ClientManager) GetStructureRetrier() *retry.Retrier {
	return cm.structureRetrier
}

// GetErrorHandler returns the error handler
func (cm *ClientManager) GetErrorHandler() *errors.ErrorHandler {
	return cm.errorHandler
}

// HealthCheck checks the health of all gRPC connections
func (cm *ClientManager) HealthCheck(ctx context.Context) map[string]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	status := make(map[string]string)
	
	// Check Auth Service
	if cm.authConn != nil {
		state := cm.authConn.GetState()
		if state == connectivity.Ready {
			status["auth"] = "healthy"
		} else {
			status["auth"] = fmt.Sprintf("unhealthy: %s", state.String())
		}
	} else {
		status["auth"] = "disconnected"
	}
	
	// Check Chat Service
	if cm.chatConn != nil {
		state := cm.chatConn.GetState()
		if state == connectivity.Ready {
			status["chat"] = "healthy"
		} else {
			status["chat"] = fmt.Sprintf("unhealthy: %s", state.String())
		}
	} else {
		status["chat"] = "disconnected"
	}
	
	// Check Employee Service
	if cm.employeeConn != nil {
		state := cm.employeeConn.GetState()
		if state == connectivity.Ready {
			status["employee"] = "healthy"
		} else {
			status["employee"] = fmt.Sprintf("unhealthy: %s", state.String())
		}
	} else {
		status["employee"] = "disconnected"
	}
	
	// Check Structure Service
	if cm.structureConn != nil {
		state := cm.structureConn.GetState()
		if state == connectivity.Ready {
			status["structure"] = "healthy"
		} else {
			status["structure"] = fmt.Sprintf("unhealthy: %s", state.String())
		}
	} else {
		status["structure"] = "disconnected"
	}
	
	return status
}

// initAuthClient initializes the Auth Service gRPC client
func (cm *ClientManager) initAuthClient(ctx context.Context) error {
	conn, err := cm.createConnection(ctx, cm.config.Services.Auth)
	if err != nil {
		return fmt.Errorf("failed to create auth connection: %w", err)
	}
	
	cm.authConn = conn
	cm.authClient = authpb.NewAuthServiceClient(conn)
	
	log.Printf("Auth Service gRPC client connected to %s", cm.config.Services.Auth.Address)
	return nil
}

// initChatClient initializes the Chat Service gRPC client
func (cm *ClientManager) initChatClient(ctx context.Context) error {
	conn, err := cm.createConnection(ctx, cm.config.Services.Chat)
	if err != nil {
		return fmt.Errorf("failed to create chat connection: %w", err)
	}
	
	cm.chatConn = conn
	cm.chatClient = chatpb.NewChatServiceClient(conn)
	
	log.Printf("Chat Service gRPC client connected to %s", cm.config.Services.Chat.Address)
	return nil
}

// initEmployeeClient initializes the Employee Service gRPC client
func (cm *ClientManager) initEmployeeClient(ctx context.Context) error {
	conn, err := cm.createConnection(ctx, cm.config.Services.Employee)
	if err != nil {
		return fmt.Errorf("failed to create employee connection: %w", err)
	}
	
	cm.employeeConn = conn
	cm.employeeClient = employeepb.NewEmployeeServiceClient(conn)
	
	log.Printf("Employee Service gRPC client connected to %s", cm.config.Services.Employee.Address)
	return nil
}

// initStructureClient initializes the Structure Service gRPC client
func (cm *ClientManager) initStructureClient(ctx context.Context) error {
	conn, err := cm.createConnection(ctx, cm.config.Services.Structure)
	if err != nil {
		return fmt.Errorf("failed to create structure connection: %w", err)
	}
	
	cm.structureConn = conn
	cm.structureClient = structurepb.NewStructureServiceClient(conn)
	
	log.Printf("Structure Service gRPC client connected to %s", cm.config.Services.Structure.Address)
	return nil
}

// createConnection creates a gRPC connection with proper configuration
func (cm *ClientManager) createConnection(ctx context.Context, cfg config.ServiceConfig) (*grpc.ClientConn, error) {
	// Connection timeout context
	connCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	
	// gRPC dial options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Wait for connection to be established
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // Send keepalive pings every 10 seconds
			Timeout:             time.Second,      // Wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // Send pings even without active streams
		}),
		// Connection pool settings
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB max receive message size
			grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB max send message size
		),
	}
	
	// Create connection
	conn, err := grpc.DialContext(connCtx, cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", cfg.Address, err)
	}
	
	// Wait for connection to be ready
	if !conn.WaitForStateChange(connCtx, connectivity.Connecting) {
		conn.Close()
		return nil, fmt.Errorf("connection to %s did not become ready within timeout", cfg.Address)
	}
	
	return conn, nil
}