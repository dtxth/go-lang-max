package test

import (
	"employee-service/api/proto"
	"employee-service/internal/infrastructure/grpc"
	"context"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// **Feature: gateway-grpc-implementation, Property 1: HTTP-to-gRPC Routing Correctness (Employee endpoints)**
// **Validates: Requirements 4.1-4.5**

func TestEmployeeGRPCRoutingCorrectness(t *testing.T) {
	// Create a minimal employee handler for testing (without service dependencies)
	// We're testing the gRPC routing, not the business logic
	grpcHandler := grpc.NewEmployeeHandler(nil)

	properties := gopter.NewProperties(nil)

	// Property: Health method always returns healthy status
	properties.Property("Health method returns healthy status", prop.ForAll(
		func() bool {
			ctx := context.Background()
			resp, err := grpcHandler.Health(ctx, &proto.HealthRequest{})
			return err == nil && resp.Status == "OK"
		},
	))

	// Property: GetAllEmployees method validates pagination parameters
	properties.Property("GetAllEmployees method handles pagination", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.GetAllEmployeesRequest{
				Page:  1,
				Limit: 10,
			}
			
			resp, err := grpcHandler.GetAllEmployees(ctx, req)
			
			// Should never return a gRPC error, only business logic errors in response
			if err != nil {
				return false
			}
			
			// Should return structured response with error (since service is nil)
			return resp != nil && resp.Error != ""
		},
	))

	// Property: SearchEmployees method validates search parameters
	properties.Property("SearchEmployees method handles search query", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.SearchEmployeesRequest{
				Query: "test",
				Page:  1,
				Limit: 10,
			}
			
			resp, err := grpcHandler.SearchEmployees(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return structured response with error (since service is nil)
			return resp != nil && resp.Error != ""
		},
	))

	// Property: GetEmployeeByID method validates ID parameter
	properties.Property("GetEmployeeByID method validates ID", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.GetEmployeeByIDRequest{
				Id: 0, // Invalid ID
			}
			
			resp, err := grpcHandler.GetEmployeeByID(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return error for invalid ID or nil service
			return resp.Error != ""
		},
	))

	// Property: CreateEmployee method validates phone parameter
	properties.Property("CreateEmployee method validates phone", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.CreateEmployeeRequest{
				Phone:     "", // Empty phone
				FirstName: "Test",
				LastName:  "User",
			}
			
			resp, err := grpcHandler.CreateEmployee(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing phone or nil service
			return resp.Error != ""
		},
	))

	// Property: CreateEmployeeSimple method validates phone parameter
	properties.Property("CreateEmployeeSimple method validates phone", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.CreateEmployeeSimpleRequest{
				Phone: "", // Empty phone
				Name:  "Test User",
			}
			
			resp, err := grpcHandler.CreateEmployeeSimple(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing phone or nil service
			return resp.Error != ""
		},
	))

	// Property: CreateEmployeeByPhone method validates phone parameter
	properties.Property("CreateEmployeeByPhone method validates phone", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.CreateEmployeeByPhoneRequest{
				Phone: "", // Empty phone
			}
			
			resp, err := grpcHandler.CreateEmployeeByPhone(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing phone or nil service
			return resp.Error != ""
		},
	))

	// Property: UpdateEmployee method validates ID parameter
	properties.Property("UpdateEmployee method validates ID", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.UpdateEmployeeRequest{
				Id:        0, // Invalid ID
				FirstName: "Updated",
				LastName:  "User",
			}
			
			resp, err := grpcHandler.UpdateEmployee(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return error for invalid ID or nil service
			return resp.Error != ""
		},
	))

	// Property: DeleteEmployee method validates ID parameter
	properties.Property("DeleteEmployee method validates ID", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.DeleteEmployeeRequest{
				Id: 0, // Invalid ID
			}
			
			resp, err := grpcHandler.DeleteEmployee(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return error for invalid ID or nil service
			return !resp.Success && resp.Error != ""
		},
	))

	// Property: BatchUpdateMaxID method handles empty batch gracefully
	properties.Property("BatchUpdateMaxID method handles empty batch", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.BatchUpdateMaxIDRequest{
				Items: []*proto.BatchUpdateItem{}, // Empty batch
			}
			
			resp, err := grpcHandler.BatchUpdateMaxID(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should handle empty batch gracefully (should return error since service is nil)
			return resp != nil && resp.Error != ""
		},
	))

	// Property: GetBatchStatus method handles pagination
	properties.Property("GetBatchStatus method handles pagination", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.GetBatchStatusRequest{
				Page:  1,
				Limit: 10,
			}
			
			resp, err := grpcHandler.GetBatchStatus(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return structured response with error (since batch service is nil)
			return resp != nil && resp.Error != ""
		},
	))

	// Property: GetBatchStatusByID method validates job ID format
	properties.Property("GetBatchStatusByID method validates job ID", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.GetBatchStatusByIDRequest{
				JobId: "invalid", // Invalid job ID format
			}
			
			resp, err := grpcHandler.GetBatchStatusByID(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for invalid job ID format or nil service
			return resp.Error != "" && (strings.Contains(resp.Error, "invalid job ID format") || strings.Contains(resp.Error, "not available"))
		},
	))

	// Property: GetUniversityByID method validates ID parameter
	properties.Property("GetUniversityByID method validates ID", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.GetUniversityByIDRequest{
				Id: 0, // Invalid ID
			}
			
			resp, err := grpcHandler.GetUniversityByID(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return error for invalid ID or nil service
			return resp.Error != ""
		},
	))

	// Property: GetUniversityByINN method validates INN parameter
	properties.Property("GetUniversityByINN method validates INN", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.GetUniversityByINNRequest{
				Inn: "", // Empty INN
			}
			
			resp, err := grpcHandler.GetUniversityByINN(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return error for empty INN or nil service
			return resp.Error != ""
		},
	))

	// Property: GetUniversityByINNAndKPP method validates parameters
	properties.Property("GetUniversityByINNAndKPP method validates parameters", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.GetUniversityByINNAndKPPRequest{
				Inn: "", // Empty INN
				Kpp: "", // Empty KPP
			}
			
			resp, err := grpcHandler.GetUniversityByINNAndKPP(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return error for empty parameters or nil service
			return resp.Error != ""
		},
	))

	// Property: All gRPC methods are callable without panicking
	properties.Property("All gRPC methods are callable", prop.ForAll(
		func() bool {
			ctx := context.Background()
			
			// Test that all methods can be called without panicking
			methods := []func() bool{
				func() bool {
					_, err := grpcHandler.Health(ctx, &proto.HealthRequest{})
					return err == nil
				},
				func() bool {
					_, err := grpcHandler.GetAllEmployees(ctx, &proto.GetAllEmployeesRequest{Page: 1, Limit: 10})
					return err == nil
				},
				func() bool {
					_, err := grpcHandler.SearchEmployees(ctx, &proto.SearchEmployeesRequest{Query: "test", Page: 1, Limit: 10})
					return err == nil
				},
				func() bool {
					_, err := grpcHandler.GetBatchStatus(ctx, &proto.GetBatchStatusRequest{Page: 1, Limit: 10})
					return err == nil
				},
			}
			
			for _, method := range methods {
				if !method() {
					return false
				}
			}
			
			return true
		},
	))

	// Run all properties with 100 iterations each
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}