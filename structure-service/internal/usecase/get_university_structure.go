package usecase

import (
	"context"
	"log"
	"sort"
	"structure-service/internal/domain"
)

// ChatService defines the interface for chat service operations
type ChatService interface {
	GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error)
}

// GetUniversityStructureUseCase handles retrieving the full university structure hierarchy
type GetUniversityStructureUseCase struct {
	repo        domain.StructureRepository
	chatService ChatService
}

// NewGetUniversityStructureUseCase creates a new instance of GetUniversityStructureUseCase
func NewGetUniversityStructureUseCase(repo domain.StructureRepository, chatService ChatService) *GetUniversityStructureUseCase {
	return &GetUniversityStructureUseCase{
		repo:        repo,
		chatService: chatService,
	}
}

// Execute retrieves the full university structure with nested hierarchy and chat details
// Requirements: 10.2, 10.3, 10.5, 13.1, 13.2, 13.3, 13.5
func (uc *GetUniversityStructureUseCase) Execute(ctx context.Context, universityID int64) (*domain.StructureNode, error) {
	// Get university
	university, err := uc.repo.GetUniversityByID(universityID)
	if err != nil {
		return nil, err
	}

	// Create root node
	root := &domain.StructureNode{
		Type:     "university",
		ID:       university.ID,
		Name:     university.Name,
		Children: []*domain.StructureNode{},
	}

	// Get branches
	branches, err := uc.repo.GetBranchesByUniversityID(universityID)
	if err != nil {
		return nil, err
	}

	// Sort branches alphabetically (Requirement 13.5)
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name < branches[j].Name
	})

	if len(branches) > 0 {
		// Structure with branches: University → Branch → Faculty → Group → Chat
		for _, branch := range branches {
			branchNode := &domain.StructureNode{
				Type:     "branch",
				ID:       branch.ID,
				Name:     branch.Name,
				Children: []*domain.StructureNode{},
			}

			// Get faculties for this branch
			faculties, err := uc.repo.GetFacultiesByBranchID(branch.ID)
			if err != nil {
				return nil, err
			}

			// Sort faculties alphabetically (Requirement 13.5)
			sort.Slice(faculties, func(i, j int) bool {
				return faculties[i].Name < faculties[j].Name
			})

			for _, faculty := range faculties {
				facultyNode := uc.buildFacultyNode(ctx, faculty)
				branchNode.Children = append(branchNode.Children, facultyNode)
			}

			root.Children = append(root.Children, branchNode)
		}
	} else {
		// Structure without branches: University → Faculty → Group → Chat (Requirement 10.5)
		faculties, err := uc.repo.GetFacultiesByUniversityID(universityID)
		if err != nil {
			return nil, err
		}

		// Filter faculties without branch_id
		var directFaculties []*domain.Faculty
		for _, faculty := range faculties {
			if faculty.BranchID == nil {
				directFaculties = append(directFaculties, faculty)
			}
		}

		// Sort faculties alphabetically (Requirement 13.5)
		sort.Slice(directFaculties, func(i, j int) bool {
			return directFaculties[i].Name < directFaculties[j].Name
		})

		for _, faculty := range directFaculties {
			facultyNode := uc.buildFacultyNode(ctx, faculty)
			root.Children = append(root.Children, facultyNode)
		}
	}

	return root, nil
}

// buildFacultyNode builds a faculty node with its groups and chat details
func (uc *GetUniversityStructureUseCase) buildFacultyNode(ctx context.Context, faculty *domain.Faculty) *domain.StructureNode {
	facultyNode := &domain.StructureNode{
		Type:     "faculty",
		ID:       faculty.ID,
		Name:     faculty.Name,
		Children: []*domain.StructureNode{},
	}

	// Get groups for this faculty
	groups, err := uc.repo.GetGroupsByFacultyID(faculty.ID)
	if err != nil {
		log.Printf("Error getting groups for faculty %d: %v", faculty.ID, err)
		return facultyNode
	}

	// Sort groups alphabetically by number (Requirement 13.5)
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Course != groups[j].Course {
			return groups[i].Course < groups[j].Course
		}
		return groups[i].Number < groups[j].Number
	})

	for _, group := range groups {
		groupNode := &domain.StructureNode{
			Type:     "group",
			ID:       group.ID,
			Name:     group.Number,
			Course:   &group.Course,
			GroupNum: &group.Number,
		}

		// If group has a chat_id, fetch chat details from chat service (Requirement 10.2, 10.3)
		if group.ChatID != nil {
			chat, err := uc.chatService.GetChatByID(ctx, *group.ChatID)
			if err != nil {
				log.Printf("Error getting chat %d for group %d: %v", *group.ChatID, group.ID, err)
				// Continue without chat details - chat may have been deleted (Requirement 10.4)
			} else if chat != nil {
				groupNode.Chat = chat
			}
		}

		facultyNode.Children = append(facultyNode.Children, groupNode)
	}

	return facultyNode
}
