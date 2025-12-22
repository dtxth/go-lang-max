package grpc

import (
	"context"
	"strconv"
	structurepb "structure-service/api/proto"
	"structure-service/internal/domain"
	"structure-service/internal/usecase"
	"time"
)

type StructureHandler struct {
	structureService              *usecase.StructureService
	createStructureUseCase        *usecase.CreateStructureFromRowUseCase
	getUniversityStructureUseCase *usecase.GetUniversityStructureUseCase
	assignOperatorUseCase         *usecase.AssignOperatorToDepartmentUseCase
	importStructureUseCase        *usecase.ImportStructureFromExcelUseCase
	dmRepo                        domain.DepartmentManagerRepository
	structurepb.UnimplementedStructureServiceServer
}

func NewStructureHandler(
	structureService *usecase.StructureService,
	createStructureUseCase *usecase.CreateStructureFromRowUseCase,
	getUniversityStructureUseCase *usecase.GetUniversityStructureUseCase,
	assignOperatorUseCase *usecase.AssignOperatorToDepartmentUseCase,
	importStructureUseCase *usecase.ImportStructureFromExcelUseCase,
	dmRepo domain.DepartmentManagerRepository,
) *StructureHandler {
	return &StructureHandler{
		structureService:              structureService,
		createStructureUseCase:        createStructureUseCase,
		getUniversityStructureUseCase: getUniversityStructureUseCase,
		assignOperatorUseCase:         assignOperatorUseCase,
		importStructureUseCase:        importStructureUseCase,
		dmRepo:                        dmRepo,
	}
}

func (h *StructureHandler) GetAllUniversities(ctx context.Context, req *structurepb.GetAllUniversitiesRequest) (*structurepb.GetAllUniversitiesResponse, error) {
	// Set defaults
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	
	offset := (page - 1) * limit
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "name"
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "asc"
	}

	universities, total, err := h.structureService.GetAllUniversitiesWithSortingAndSearch(int(limit), int(offset), sortBy, sortOrder, "")
	if err != nil {
		return &structurepb.GetAllUniversitiesResponse{
			Error: err.Error(),
		}, nil
	}

	var pbUniversities []*structurepb.University
	for _, u := range universities {
		pbUniversities = append(pbUniversities, &structurepb.University{
			Id:        u.ID,
			Name:      u.Name,
			Inn:       u.INN,
			Kpp:       u.KPP,
			Foiv:      u.FOIV,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &structurepb.GetAllUniversitiesResponse{
		Universities: pbUniversities,
		Total:        int32(total),
		Page:         page,
		Limit:        limit,
	}, nil
}

func (h *StructureHandler) CreateUniversity(ctx context.Context, req *structurepb.CreateUniversityRequest) (*structurepb.CreateUniversityResponse, error) {
	university := &domain.University{
		Name: req.Name,
		INN:  req.Inn,
		KPP:  req.Kpp,
		FOIV: req.Foiv,
	}

	err := h.structureService.CreateUniversity(university)
	if err != nil {
		return &structurepb.CreateUniversityResponse{
			Error: err.Error(),
		}, nil
	}

	return &structurepb.CreateUniversityResponse{
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

func (h *StructureHandler) GetUniversityStructure(ctx context.Context, req *structurepb.GetUniversityStructureRequest) (*structurepb.GetUniversityStructureResponse, error) {
	structure, err := h.getUniversityStructureUseCase.Execute(ctx, req.UniversityId)
	if err != nil {
		return &structurepb.GetUniversityStructureResponse{
			Error: err.Error(),
		}, nil
	}

	// Convert domain structure to protobuf
	pbStructure := h.convertStructureNodeToPB(structure)

	return &structurepb.GetUniversityStructureResponse{
		Structure: pbStructure,
	}, nil
}

func (h *StructureHandler) UpdateUniversityName(ctx context.Context, req *structurepb.UpdateUniversityNameRequest) (*structurepb.UpdateUniversityNameResponse, error) {
	err := h.structureService.UpdateUniversityName(req.UniversityId, req.Name)
	if err != nil {
		return &structurepb.UpdateUniversityNameResponse{
			Error: err.Error(),
		}, nil
	}

	// Get updated university
	university, err := h.structureService.GetUniversity(req.UniversityId)
	if err != nil {
		return &structurepb.UpdateUniversityNameResponse{
			Error: err.Error(),
		}, nil
	}

	return &structurepb.UpdateUniversityNameResponse{
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

func (h *StructureHandler) ImportExcel(ctx context.Context, req *structurepb.ImportExcelRequest) (*structurepb.ImportExcelResponse, error) {
	// Parse Excel data - this would need to be implemented based on your Excel parsing logic
	// For now, return a placeholder response
	return &structurepb.ImportExcelResponse{
		ProcessedRows:  0,
		SuccessfulRows: 0,
		FailedRows:     0,
		Errors:         []string{"Excel import not yet implemented"},
		Error:          "Excel import functionality needs to be implemented",
	}, nil
}

func (h *StructureHandler) UpdateBranchName(ctx context.Context, req *structurepb.UpdateBranchNameRequest) (*structurepb.UpdateBranchNameResponse, error) {
	err := h.structureService.UpdateBranchName(req.BranchId, req.Name)
	if err != nil {
		return &structurepb.UpdateBranchNameResponse{
			Error: err.Error(),
		}, nil
	}

	// Get updated branch
	branch, err := h.structureService.GetBranchByID(req.BranchId)
	if err != nil {
		return &structurepb.UpdateBranchNameResponse{
			Error: err.Error(),
		}, nil
	}

	return &structurepb.UpdateBranchNameResponse{
		Branch: &structurepb.Branch{
			Id:           branch.ID,
			Name:         branch.Name,
			UniversityId: branch.UniversityID,
			CreatedAt:    branch.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    branch.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *StructureHandler) UpdateFacultyName(ctx context.Context, req *structurepb.UpdateFacultyNameRequest) (*structurepb.UpdateFacultyNameResponse, error) {
	err := h.structureService.UpdateFacultyName(req.FacultyId, req.Name)
	if err != nil {
		return &structurepb.UpdateFacultyNameResponse{
			Error: err.Error(),
		}, nil
	}

	// Get updated faculty
	faculty, err := h.structureService.GetFacultyByID(req.FacultyId)
	if err != nil {
		return &structurepb.UpdateFacultyNameResponse{
			Error: err.Error(),
		}, nil
	}

	pbFaculty := &structurepb.Faculty{
		Id:        faculty.ID,
		Name:      faculty.Name,
		CreatedAt: faculty.CreatedAt.Format(time.RFC3339),
		UpdatedAt: faculty.UpdatedAt.Format(time.RFC3339),
	}
	if faculty.BranchID != nil {
		pbFaculty.BranchId = *faculty.BranchID
	}

	return &structurepb.UpdateFacultyNameResponse{
		Faculty: pbFaculty,
	}, nil
}

func (h *StructureHandler) UpdateGroupName(ctx context.Context, req *structurepb.UpdateGroupNameRequest) (*structurepb.UpdateGroupNameResponse, error) {
	err := h.structureService.UpdateGroupName(req.GroupId, req.Number)
	if err != nil {
		return &structurepb.UpdateGroupNameResponse{
			Error: err.Error(),
		}, nil
	}

	// Get updated group
	group, err := h.structureService.GetGroupByID(req.GroupId)
	if err != nil {
		return &structurepb.UpdateGroupNameResponse{
			Error: err.Error(),
		}, nil
	}

	pbGroup := &structurepb.Group{
		Id:        group.ID,
		Number:    group.Number,
		Course:    int32(group.Course),
		FacultyId: group.FacultyID,
		CreatedAt: group.CreatedAt.Format(time.RFC3339),
		UpdatedAt: group.UpdatedAt.Format(time.RFC3339),
	}
	if group.ChatID != nil {
		pbGroup.ChatId = group.ChatID
	}

	return &structurepb.UpdateGroupNameResponse{
		Group: pbGroup,
	}, nil
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

func (h *StructureHandler) GetAllDepartmentManagers(ctx context.Context, req *structurepb.GetAllDepartmentManagersRequest) (*structurepb.GetAllDepartmentManagersResponse, error) {
	managers, err := h.dmRepo.GetAllDepartmentManagers()
	if err != nil {
		return &structurepb.GetAllDepartmentManagersResponse{
			Error: err.Error(),
		}, nil
	}

	// Apply pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	start := (page - 1) * limit
	end := start + limit
	total := int32(len(managers))

	var paginatedManagers []*domain.DepartmentManager
	if int(start) < len(managers) {
		if int(end) > len(managers) {
			end = int32(len(managers))
		}
		paginatedManagers = managers[start:end]
	}

	var pbManagers []*structurepb.DepartmentManager
	for _, dm := range paginatedManagers {
		pbDM := &structurepb.DepartmentManager{
			Id:        dm.ID,
			UserId:    strconv.FormatInt(dm.EmployeeID, 10), // Convert int64 to string
			CreatedAt: dm.AssignedAt.Format(time.RFC3339),
		}
		if dm.BranchID != nil {
			pbDM.DepartmentId = *dm.BranchID
		} else if dm.FacultyID != nil {
			pbDM.DepartmentId = *dm.FacultyID
		}
		pbManagers = append(pbManagers, pbDM)
	}

	return &structurepb.GetAllDepartmentManagersResponse{
		Managers: pbManagers,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

func (h *StructureHandler) CreateDepartmentManager(ctx context.Context, req *structurepb.CreateDepartmentManagerRequest) (*structurepb.CreateDepartmentManagerResponse, error) {
	// Convert string user_id to int64 employee_id
	employeeID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return &structurepb.CreateDepartmentManagerResponse{
			Error: "invalid user_id format",
		}, nil
	}
	
	// Determine if it's a branch or faculty department
	var branchID, facultyID *int64
	if req.DepartmentId > 0 {
		// For simplicity, assume it's a faculty ID - in real implementation you'd need to determine the type
		facultyID = &req.DepartmentId
	}

	dm, err := h.assignOperatorUseCase.Execute(employeeID, branchID, facultyID, nil)
	if err != nil {
		return &structurepb.CreateDepartmentManagerResponse{
			Error: err.Error(),
		}, nil
	}

	pbDM := &structurepb.DepartmentManager{
		Id:           dm.ID,
		UserId:       req.UserId,
		DepartmentId: req.DepartmentId,
		CreatedAt:    dm.AssignedAt.Format(time.RFC3339),
	}

	return &structurepb.CreateDepartmentManagerResponse{
		Manager: pbDM,
	}, nil
}

func (h *StructureHandler) RemoveDepartmentManager(ctx context.Context, req *structurepb.RemoveDepartmentManagerRequest) (*structurepb.RemoveDepartmentManagerResponse, error) {
	err := h.dmRepo.DeleteDepartmentManager(req.ManagerId)
	if err != nil {
		return &structurepb.RemoveDepartmentManagerResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &structurepb.RemoveDepartmentManagerResponse{
		Success: true,
	}, nil
}

func (h *StructureHandler) Health(ctx context.Context, req *structurepb.HealthRequest) (*structurepb.HealthResponse, error) {
	return &structurepb.HealthResponse{
		Status: "healthy",
	}, nil
}

// Helper method to convert domain structure to protobuf
func (h *StructureHandler) convertStructureNodeToPB(node *domain.StructureNode) *structurepb.UniversityStructure {
	if node == nil || node.Type != "university" {
		return nil
	}

	university := &structurepb.University{
		Id:   node.ID,
		Name: node.Name,
	}

	var branches []*structurepb.Branch
	for _, child := range node.Children {
		if child.Type == "branch" {
			branch := &structurepb.Branch{
				Id:           child.ID,
				Name:         child.Name,
				UniversityId: node.ID,
			}
			
			// Convert faculties
			for _, facultyNode := range child.Children {
				if facultyNode.Type == "faculty" {
					faculty := &structurepb.Faculty{
						Id:       facultyNode.ID,
						Name:     facultyNode.Name,
						BranchId: child.ID,
					}
					
					// Convert groups
					for _, groupNode := range facultyNode.Children {
						if groupNode.Type == "group" {
							group := &structurepb.Group{
								Id:        groupNode.ID,
								Number:    groupNode.Name,
								FacultyId: facultyNode.ID,
							}
							if groupNode.Course != nil {
								group.Course = int32(*groupNode.Course)
							}
							if groupNode.Chat != nil {
								group.ChatId = &groupNode.Chat.ID
							}
							faculty.Groups = append(faculty.Groups, group)
						}
					}
					branch.Faculties = append(branch.Faculties, faculty)
				}
			}
			branches = append(branches, branch)
		}
	}

	return &structurepb.UniversityStructure{
		University: university,
		Branches:   branches,
	}
}