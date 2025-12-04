package domain

import "context"

// ChatData represents chat data for migration
type ChatData struct {
	Name           string
	URL            string
	ExternalChatID string  // ID чата из внешней системы (Excel колонка 14)
	UniversityID   int
	BranchID       *int
	FacultyID      *int
	Source         string
	AdminPhone     string
}

// AdministratorData represents administrator data for migration
type AdministratorData struct {
	ChatID   int
	Phone    string
	MaxID    string
	AddUser  bool // Может ли добавлять пользователей (Excel колонка 16)
	AddAdmin bool // Может ли добавлять администраторов (Excel колонка 17)
}

// UniversityData represents university data for migration
type UniversityData struct {
	INN  string
	KPP  string
	Name string
}

// ChatService defines the interface for interacting with Chat Service
type ChatService interface {
	// CreateChat creates a new chat
	CreateChat(ctx context.Context, chat *ChatData) (int, error)

	// AddAdministrator adds an administrator to a chat
	AddAdministrator(ctx context.Context, admin *AdministratorData) error

	// CreateOrGetUniversity creates or gets a university by INN/KPP
	CreateOrGetUniversity(ctx context.Context, university *UniversityData) (int, error)
}
