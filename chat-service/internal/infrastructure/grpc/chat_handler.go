package grpc

import (
	"chat-service/api/proto"
	"chat-service/internal/usecase"
	"context"
	"time"
)

type ChatHandler struct {
	chatService *usecase.ChatService
	proto.UnimplementedChatServiceServer
}

func NewChatHandler(chatService *usecase.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) GetChatByID(ctx context.Context, req *proto.GetChatByIDRequest) (*proto.GetChatByIDResponse, error) {
	chat, err := h.chatService.GetChatByID(req.Id)
	if err != nil {
		return &proto.GetChatByIDResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetChatByIDResponse{
		Chat: &proto.Chat{
			Id:                chat.ID,
			Name:              chat.Name,
			Url:               chat.URL,
			MaxChatId:         chat.MaxChatID,
			ParticipantsCount: int32(chat.ParticipantsCount),
			UniversityId:      chat.UniversityID,
			Department:        chat.Department,
			Source:            chat.Source,
			CreatedAt:         chat.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         chat.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *ChatHandler) CreateChat(ctx context.Context, req *proto.CreateChatRequest) (*proto.CreateChatResponse, error) {
	var universityID *int64
	if req.UniversityId != nil {
		universityID = req.UniversityId
	}

	chat, err := h.chatService.CreateChat(
		req.Name,
		req.Url,
		req.MaxChatId,
		req.Source,
		int(req.ParticipantsCount),
		universityID,
		req.Department,
	)
	if err != nil {
		return &proto.CreateChatResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.CreateChatResponse{
		Chat: &proto.Chat{
			Id:                chat.ID,
			Name:              chat.Name,
			Url:               chat.URL,
			MaxChatId:         chat.MaxChatID,
			ParticipantsCount: int32(chat.ParticipantsCount),
			UniversityId:      chat.UniversityID,
			Department:        chat.Department,
			Source:            chat.Source,
			CreatedAt:         chat.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         chat.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *ChatHandler) AddAdministratorForMigration(ctx context.Context, req *proto.AddAdministratorForMigrationRequest) (*proto.AddAdministratorForMigrationResponse, error) {
	// Используем метод с флагом skipPhoneValidation=true для миграции
	admin, err := h.chatService.AddAdministratorWithFlags(
		req.ChatId,
		req.Phone,
		req.MaxId,
		req.AddUser,
		req.AddAdmin,
		true, // skipPhoneValidation = true для миграции
	)
	if err != nil {
		// Log error for debugging
		return &proto.AddAdministratorForMigrationResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.AddAdministratorForMigrationResponse{
		Administrator: &proto.Administrator{
			Id:        admin.ID,
			ChatId:    admin.ChatID,
			Phone:     admin.Phone,
			MaxId:     admin.MaxID,
			AddUser:   admin.AddUser,
			AddAdmin:  admin.AddAdmin,
			CreatedAt: admin.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

