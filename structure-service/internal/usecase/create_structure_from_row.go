package usecase

import (
	"context"
	"fmt"
	"structure-service/internal/domain"
)

// CreateStructureFromRowUseCase создает полную структуру из одной строки данных
type CreateStructureFromRowUseCase struct {
	repo domain.StructureRepository
}

func NewCreateStructureFromRowUseCase(repo domain.StructureRepository) *CreateStructureFromRowUseCase {
	return &CreateStructureFromRowUseCase{
		repo: repo,
	}
}

// CreateStructureRequest представляет запрос на создание структуры
type CreateStructureRequest struct {
	INN         string `json:"inn"`
	KPP         string `json:"kpp"`
	FOIV        string `json:"foiv"`
	OrgName     string `json:"org_name"`
	BranchName  string `json:"branch_name,omitempty"`
	FacultyName string `json:"faculty_name"`
	Course      int    `json:"course"`
	GroupNumber string `json:"group_number"`
}

// CreateStructureResponse представляет ответ с созданной структурой
type CreateStructureResponse struct {
	UniversityID int64  `json:"university_id"`
	BranchID     *int64 `json:"branch_id,omitempty"`
	FacultyID    *int64 `json:"faculty_id,omitempty"`
	GroupID      int64  `json:"group_id"`
}

// Execute создает или находит все элементы структуры
func (uc *CreateStructureFromRowUseCase) Execute(ctx context.Context, req *CreateStructureRequest) (*CreateStructureResponse, error) {
	// 1. Обработка University
	university, err := uc.repo.GetUniversityByINN(req.INN)
	if err == domain.ErrUniversityNotFound {
		// Создаем новый вуз
		university = &domain.University{
			Name: req.OrgName,
			INN:  req.INN,
			KPP:  req.KPP,
			FOIV: req.FOIV,
		}
		if err := uc.repo.CreateUniversity(university); err != nil {
			return nil, fmt.Errorf("failed to create university: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get university: %w", err)
	}

	response := &CreateStructureResponse{
		UniversityID: university.ID,
	}

	// 2. Обработка Branch (если указан)
	var branch *domain.Branch
	if req.BranchName != "" {
		branch, err = uc.repo.GetBranchByUniversityAndName(university.ID, req.BranchName)
		if err == domain.ErrBranchNotFound {
			// Создаем новый филиал
			branch = &domain.Branch{
				UniversityID: university.ID,
				Name:         req.BranchName,
			}
			if err := uc.repo.CreateBranch(branch); err != nil {
				return nil, fmt.Errorf("failed to create branch: %w", err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to get branch: %w", err)
		}
		response.BranchID = &branch.ID
	}

	// 3. Обработка Faculty
	var branchIDPtr *int64
	if branch != nil {
		branchIDPtr = &branch.ID
	}

	faculty, err := uc.repo.GetFacultyByBranchAndName(branchIDPtr, req.FacultyName)
	if err == domain.ErrFacultyNotFound {
		// Создаем новый факультет
		faculty = &domain.Faculty{
			Name:     req.FacultyName,
			BranchID: branchIDPtr,
		}
		if err := uc.repo.CreateFaculty(faculty); err != nil {
			return nil, fmt.Errorf("failed to create faculty: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get faculty: %w", err)
	}
	response.FacultyID = &faculty.ID

	// 4. Обработка Group
	group, err := uc.repo.GetGroupByFacultyAndNumber(faculty.ID, req.Course, req.GroupNumber)
	if err == domain.ErrGroupNotFound {
		// Создаем новую группу
		group = &domain.Group{
			FacultyID: faculty.ID,
			Number:    req.GroupNumber,
			Course:    req.Course,
		}
		if err := uc.repo.CreateGroup(group); err != nil {
			return nil, fmt.Errorf("failed to create group: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	response.GroupID = group.ID

	return response, nil
}
