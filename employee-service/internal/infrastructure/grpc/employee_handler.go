package grpc

import (
	"employee-service/api/proto"
	"employee-service/internal/usecase"
	"context"
	"time"
)

type EmployeeHandler struct {
	employeeService *usecase.EmployeeService
	proto.UnimplementedEmployeeServiceServer
}

func NewEmployeeHandler(employeeService *usecase.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
	}
}

func (h *EmployeeHandler) GetUniversityByID(ctx context.Context, req *proto.GetUniversityByIDRequest) (*proto.GetUniversityByIDResponse, error) {
	university, err := h.employeeService.GetUniversityByID(req.Id)
	if err != nil {
		return &proto.GetUniversityByIDResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetUniversityByIDResponse{
		University: &proto.University{
			Id:        university.ID,
			Name:      university.Name,
			Inn:       university.INN,
			Kpp:       university.KPP,
			CreatedAt: university.CreatedAt.Format(time.RFC3339),
			UpdatedAt: university.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *EmployeeHandler) GetUniversityByINN(ctx context.Context, req *proto.GetUniversityByINNRequest) (*proto.GetUniversityByINNResponse, error) {
	university, err := h.employeeService.GetUniversityByINN(req.Inn)
	if err != nil {
		return &proto.GetUniversityByINNResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetUniversityByINNResponse{
		University: &proto.University{
			Id:        university.ID,
			Name:      university.Name,
			Inn:       university.INN,
			Kpp:       university.KPP,
			CreatedAt: university.CreatedAt.Format(time.RFC3339),
			UpdatedAt: university.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *EmployeeHandler) GetUniversityByINNAndKPP(ctx context.Context, req *proto.GetUniversityByINNAndKPPRequest) (*proto.GetUniversityByINNAndKPPResponse, error) {
	university, err := h.employeeService.GetUniversityByINNAndKPP(req.Inn, req.Kpp)
	if err != nil {
		return &proto.GetUniversityByINNAndKPPResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetUniversityByINNAndKPPResponse{
		University: &proto.University{
			Id:        university.ID,
			Name:      university.Name,
			Inn:       university.INN,
			Kpp:       university.KPP,
			CreatedAt: university.CreatedAt.Format(time.RFC3339),
			UpdatedAt: university.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

