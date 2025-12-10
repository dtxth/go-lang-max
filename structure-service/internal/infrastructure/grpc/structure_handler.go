package grpc

import (
	"context"
	structurepb "structure-service/api/proto"
	"structure-service/internal/usecase"
	"time"
)

type StructureHandler struct {
	structureService       *usecase.StructureService
	createStructureUseCase *usecase.CreateStructureFromRowUseCase
	structurepb.UnimplementedStructureServiceServer
}

func NewStructureHandler(
	structureService *usecase.StructureService,
	createStructureUseCase *usecase.CreateStructureFromRowUseCase,
) *StructureHandler {
	return &StructureHandler{
		structureService:       structureService,
		createStructureUseCase: createStructureUseCase,
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

func (h *StructureHandler) CreateOrGetUniversity(ctx context.Context, req *structurepb.CreateOrGetUniversityRequest) (*structurepb.CreateOrGetUniversityResponse, error) {
	university, err := h.structureService.CreateOrGetUniversity(req.Inn, req.Kpp, req.Name, req.Foiv)
	if err != nil {
		return &structurepb.CreateOrGetUniversityResponse{
			Error: err.Error(),
		}, nil
	}

	return &structurepb.CreateOrGetUniversityResponse{
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

func (h *StructureHandler) CreateStructure(ctx context.Context, req *structurepb.CreateStructureRequest) (*structurepb.CreateStructureResponse, error) {
	ucReq := &usecase.CreateStructureRequest{
		INN:         req.Inn,
		KPP:         req.Kpp,
		FOIV:        req.Foiv,
		OrgName:     req.OrgName,
		BranchName:  req.BranchName,
		FacultyName: req.FacultyName,
		Course:      int(req.Course),
		GroupNumber: req.GroupNumber,
	}

	result, err := h.createStructureUseCase.Execute(ctx, ucReq)
	if err != nil {
		return &structurepb.CreateStructureResponse{
			Error: err.Error(),
		}, nil
	}

	response := &structurepb.CreateStructureResponse{
		UniversityId: result.UniversityID,
		GroupId:      result.GroupID,
	}

	if result.BranchID != nil {
		response.BranchId = result.BranchID
	}
	if result.FacultyID != nil {
		response.FacultyId = result.FacultyID
	}

	return response, nil
}

func (h *StructureHandler) LinkGroupToChat(ctx context.Context, req *structurepb.LinkGroupToChatRequest) (*structurepb.LinkGroupToChatResponse, error) {
	// Get group
	group, err := h.structureService.GetGroupByID(req.GroupId)
	if err != nil {
		return &structurepb.LinkGroupToChatResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Update chat_id
	group.ChatID = &req.ChatId
	if err := h.structureService.UpdateGroup(group); err != nil {
		return &structurepb.LinkGroupToChatResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &structurepb.LinkGroupToChatResponse{
		Success: true,
	}, nil
}
