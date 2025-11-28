package domain

import "context"

// ChatData represents chat data for migration
type ChatData struct {
	Name          string
	URL           string
	UniversityID  int
	BranchID      *int
	FacultyID     *int
	Source        string
	AdminPhone    string
}

// ChatService defines the interface for interacting with Chat Service
type ChatService interface {
	// CreateChat creates a new chat
	CreateChat(ctx context.Context, chat *ChatData) (int, error)

	// AddAdministrator adds an administrator to a chat
	AddAdministrator(ctx context.Context, chatID int, phone string) error
}
