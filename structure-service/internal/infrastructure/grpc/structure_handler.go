package grpc

import (
	"context"
	structurepb "structure-service/api/proto"
	"structure-service/internal/usecase"
	"time"
)

type StructureHandler struct {
	structureService *usecase.StructureService
	structurepb.UnimplementedStructureServiceServer
}

func NewStructureHandler(structureService *usecase.StructureService) *StructureHandler {
	return &StructureHandler{
		structureService: structureService,
	}
}

func (h *StructureHandler) GetUniversityByID(ctx context.Context, req *structurepb.GetUniversityByIDRequest) (*structurepb.GetUniversityByIDResponse, error) {
	university, err := h.structureService.GetUniversity(req.Id)
	if err != nil {
		return &structurepb.GetUniversityByIDResponse{
			Error: err.Error(),
		}, nil
	}

	return &structurepb.GetUniversityByIDResponse{
		University: &structurepb.University{
			Id:        university.ID,
			Name:      university.Name,
			Inn:       university.INN,
			Kpp:       university.KPP,
			Foiv:      university.FOIV,
			CreatedAt: university.CreatedAt.Format(time.RFC3339),
			UpdatedAt: university.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *StructureHandler) GetUniversityByINN(ctx context.Context, req *structurepb.GetUniversityByINNRequest) (*structurepb.GetUniversityByINNResponse, error) {
	university, err := h.structureService.GetUniversityByINN(req.Inn)
	if err != nil {
		return &structurepb.GetUniversityByINNResponse{
			Error: err.Error(),
		}, nil
	}

	return &structurepb.GetUniversityByINNResponse{
		University: &structurepb.University{
			Id:        university.ID,
			Name:      university.Name,
			Inn:       university.INN,
			Kpp:       university.KPP,
			Foiv:      university.FOIV,
			CreatedAt: university.CreatedAt.Format(time.RFC3339),
			UpdatedAt: university.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}
