package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway-service/internal/config"
)

// ServiceInfo holds information about a registered service
type ServiceInfo struct {
	Name    string
	Address string
	Status  string
	LastCheck time.Time
}

// ServiceRegistry manages service discovery and health monitoring
type ServiceRegistry struct {
	config   *config.Config
	services map[string]*ServiceInfo
	mu       sync.RWMutex
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(cfg *config.Config) *ServiceRegistry {
	return &ServiceRegistry{
		config:   cfg,
		services: make(map[string]*ServiceInfo),
	}
}

// RegisterServices registers all configured services
func (sr *ServiceRegistry) RegisterServices() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	
	// Register Auth Service
	sr.services["auth"] = &ServiceInfo{
		Name:    "auth",
		Address: sr.config.Services.Auth.Address,
		Status:  "unknown",
		LastCheck: time.Now(),
	}
	
	// Register Chat Service
	sr.services["chat"] = &ServiceInfo{
		Name:    "chat",
		Address: sr.config.Services.Chat.Address,
		Status:  "unknown",
		LastCheck: time.Now(),
	}
	
	// Register Employee Service
	sr.services["employee"] = &ServiceInfo{
		Name:    "employee",
		Address: sr.config.Services.Employee.Address,
		Status:  "unknown",
		LastCheck: time.Now(),
	}
	
	// Register Structure Service
	sr.services["structure"] = &ServiceInfo{
		Name:    "structure",
		Address: sr.config.Services.Structure.Address,
		Status:  "unknown",
		LastCheck: time.Now(),
	}
}

// GetService returns information about a specific service
func (sr *ServiceRegistry) GetService(name string) (*ServiceInfo, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	service, exists := sr.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	
	// Return a copy to avoid race conditions
	return &ServiceInfo{
		Name:      service.Name,
		Address:   service.Address,
		Status:    service.Status,
		LastCheck: service.LastCheck,
	}, nil
}

// GetAllServices returns information about all registered services
func (sr *ServiceRegistry) GetAllServices() map[string]*ServiceInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	result := make(map[string]*ServiceInfo)
	for name, service := range sr.services {
		result[name] = &ServiceInfo{
			Name:      service.Name,
			Address:   service.Address,
			Status:    service.Status,
			LastCheck: service.LastCheck,
		}
	}
	
	return result
}

// UpdateServiceStatus updates the status of a service
func (sr *ServiceRegistry) UpdateServiceStatus(name, status string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	
	if service, exists := sr.services[name]; exists {
		service.Status = status
		service.LastCheck = time.Now()
	}
}

// IsServiceHealthy checks if a service is healthy
func (sr *ServiceRegistry) IsServiceHealthy(name string) bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	service, exists := sr.services[name]
	if !exists {
		return false
	}
	
	return service.Status == "healthy"
}

// GetHealthyServices returns a list of healthy services
func (sr *ServiceRegistry) GetHealthyServices() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	var healthy []string
	for name, service := range sr.services {
		if service.Status == "healthy" {
			healthy = append(healthy, name)
		}
	}
	
	return healthy
}

// StartHealthMonitoring starts periodic health monitoring for all services
func (sr *ServiceRegistry) StartHealthMonitoring(ctx context.Context, clientManager *ClientManager) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sr.performHealthChecks(ctx, clientManager)
		}
	}
}

// performHealthChecks performs health checks on all registered services
func (sr *ServiceRegistry) performHealthChecks(ctx context.Context, clientManager *ClientManager) {
	// Get current connection status from client manager
	status := clientManager.HealthCheck(ctx)
	
	// Update service registry with current status
	for serviceName, serviceStatus := range status {
		sr.UpdateServiceStatus(serviceName, serviceStatus)
	}
}