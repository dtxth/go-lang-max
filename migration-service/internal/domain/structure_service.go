package domain

import "context"

// StructureData represents structure data for migration
type StructureData struct {
	INN          string
	KPP          string
	FOIV         string
	OrgName      string
	BranchName   string
	FacultyName  string
	Course       int
	GroupNumber  string
	ChatName     string
	ChatURL      string
	AdminPhone   string
}

// StructureResult represents the result of structure creation
type StructureResult struct {
	UniversityID int
	BranchID     *int
	FacultyID    *int
	GroupID      int
}

// StructureService defines the interface for interacting with Structure Service
type StructureService interface {
	// CreateStructure creates or updates the full structure hierarchy
	CreateStructure(ctx context.Context, data *StructureData) (*StructureResult, error)

	// LinkGroupToChat links a group to a chat
	LinkGroupToChat(ctx context.Context, groupID int, chatID int) error
}
