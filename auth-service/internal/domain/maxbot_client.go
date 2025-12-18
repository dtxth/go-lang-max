package domain

import "context"

// MaxBotClient defines the interface for interacting with MaxBot service
type MaxBotClient interface {
	// GetBotInfo retrieves bot information (name and add link)
	GetBotInfo(ctx context.Context) (*BotInfo, error)
}

// BotInfo contains bot information
type BotInfo struct {
	Name    string // Bot name
	AddLink string // Link to add the bot
}