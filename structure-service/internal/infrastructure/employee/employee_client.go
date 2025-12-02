package employee

import (
	"context"
	"fmt"
	pb "structure-service/api/proto/employee"
	"structure-service/internal/domain"
	grpcretry "structure-service/internal/infrastructure/grpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// EmployeeClient реализует интерфейс domain.EmployeeService через gRPC
type EmployeeClient struct {
	conn   *grpc.ClientConn
	client pb.EmployeeServiceClient
}

// NewEmployeeClient создает новый gRPC клиент для Employee Service
func NewEmployeeClient(address string) (*EmployeeClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to employee service: %w", err)
	}

	client := pb.NewEmployeeServiceClient(conn)
	return &EmployeeClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close закрывает соединение с Employee Service
func (c *EmployeeClient) Close() error {
	return c.conn.Close()
}

// GetEmployeeByID получает сотрудника по ID через gRPC
func (c *EmployeeClient) GetEmployeeByID(id int64) (*domain.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetEmployeeByIDRequest{
		Id: id,
	}

	var resp *pb.GetEmployeeByIDResponse
	err := grpcretry.WithRetry(ctx, "Employee.GetEmployeeByID", func() error {
		var callErr error
		resp, callErr = c.client.GetEmployeeByID(ctx, req)
		return callErr
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	if resp.Error != "" {
		if resp.Error == "employee not found" {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, fmt.Errorf("employee service error: %s", resp.Error)
	}

	if resp.Employee == nil {
		return nil, domain.ErrEmployeeNotFound
	}

	return &domain.Employee{
		ID:           resp.Employee.Id,
		FirstName:    resp.Employee.FirstName,
		LastName:     resp.Employee.LastName,
		Phone:        resp.Employee.Phone,
		Role:         resp.Employee.Role,
		UniversityID: resp.Employee.UniversityId,
	}, nil
}
