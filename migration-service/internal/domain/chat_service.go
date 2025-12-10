package domain

import "context"

// ChatData represents chat data for migration
type ChatData struct {
	Name           string
	URL            string
	ExternalChatID string  // ID чата из внешней системы (Excel колонка 14)
	UniversityID   *int    // Опциональный ID университета
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



// ChatService defines the interface for interacting with Chat Service
type ChatService interface {
	// CreateChat creates a new chat
	CreateChat(ctx context.Context, chat *ChatData) (int, error)

	// AddAdministrator adds an administrator to a chat
	AddAdministrator(ctx context.Context, admin *AdministratorData) error
}
