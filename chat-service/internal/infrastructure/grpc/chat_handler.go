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

// GetAllChats получает все чаты с пагинацией и поиском
func (h *ChatHandler) GetAllChats(ctx context.Context, req *proto.GetAllChatsRequest) (*proto.GetAllChatsResponse, error) {
	page := int(req.Page)
	limit := int(req.Limit)
	
	// Валидация параметров пагинации
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	
	offset := (page - 1) * limit
	
	chats, total, err := h.chatService.GetAllChatsWithSortingAndSearch(
		limit, 
		offset, 
		req.SortBy, 
		req.SortOrder, 
		"", // no search query for GetAllChats
		nil, // no filter for now
	)
	if err != nil {
		return &proto.GetAllChatsResponse{
			Error: err.Error(),
		}, nil
	}

	protoChats := make([]*proto.Chat, len(chats))
	for i, chat := range chats {
		protoChats[i] = &proto.Chat{
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
		}
	}

	return &proto.GetAllChatsResponse{
		Chats: protoChats,
		Total: int32(total),
		Page:  int32(page),
		Limit: int32(limit),
	}, nil
}

// SearchChats ищет чаты по параметрам
func (h *ChatHandler) SearchChats(ctx context.Context, req *proto.SearchChatsRequest) (*proto.SearchChatsResponse, error) {
	page := int(req.Page)
	limit := int(req.Limit)
	
	// Валидация параметров пагинации
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	
	offset := (page - 1) * limit
	
	chats, total, err := h.chatService.GetAllChatsWithSortingAndSearch(
		limit, 
		offset, 
		req.SortBy, 
		req.SortOrder, 
		req.Query, // search query
		nil, // no filter for now
	)
	if err != nil {
		return &proto.SearchChatsResponse{
			Error: err.Error(),
		}, nil
	}

	protoChats := make([]*proto.Chat, len(chats))
	for i, chat := range chats {
		protoChats[i] = &proto.Chat{
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
		}
	}

	return &proto.SearchChatsResponse{
		Chats: protoChats,
		Total: int32(total),
		Page:  int32(page),
		Limit: int32(limit),
	}, nil
}

// GetAllAdministrators получает всех администраторов
func (h *ChatHandler) GetAllAdministrators(ctx context.Context, req *proto.GetAllAdministratorsRequest) (*proto.GetAllAdministratorsResponse, error) {
	page := int(req.Page)
	limit := int(req.Limit)
	
	// Валидация параметров пагинации
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	
	offset := (page - 1) * limit
	
	administrators, total, err := h.chatService.GetAllAdministrators("", limit, offset)
	if err != nil {
		return &proto.GetAllAdministratorsResponse{
			Error: err.Error(),
		}, nil
	}

	protoAdmins := make([]*proto.Administrator, len(administrators))
	for i, admin := range administrators {
		protoAdmins[i] = &proto.Administrator{
			Id:        admin.ID,
			ChatId:    admin.ChatID,
			Phone:     admin.Phone,
			MaxId:     admin.MaxID,
			AddUser:   admin.AddUser,
			AddAdmin:  admin.AddAdmin,
			CreatedAt: admin.CreatedAt.Format(time.RFC3339),
		}
	}

	return &proto.GetAllAdministratorsResponse{
		Administrators: protoAdmins,
		Total:          int32(total),
		Page:           int32(page),
		Limit:          int32(limit),
	}, nil
}

// GetAdministratorByID получает администратора по ID
func (h *ChatHandler) GetAdministratorByID(ctx context.Context, req *proto.GetAdministratorByIDRequest) (*proto.GetAdministratorByIDResponse, error) {
	admin, err := h.chatService.GetAdministratorByID(req.Id)
	if err != nil {
		return &proto.GetAdministratorByIDResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetAdministratorByIDResponse{
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

// AddAdministrator добавляет администратора к чату
func (h *ChatHandler) AddAdministrator(ctx context.Context, req *proto.AddAdministratorRequest) (*proto.AddAdministratorResponse, error) {
	admin, err := h.chatService.AddAdministrator(req.ChatId, req.Phone)
	if err != nil {
		return &proto.AddAdministratorResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.AddAdministratorResponse{
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

// RemoveAdministrator удаляет администратора
func (h *ChatHandler) RemoveAdministrator(ctx context.Context, req *proto.RemoveAdministratorRequest) (*proto.RemoveAdministratorResponse, error) {
	err := h.chatService.RemoveAdministrator(req.Id)
	if err != nil {
		return &proto.RemoveAdministratorResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.RemoveAdministratorResponse{
		Success: true,
	}, nil
}

// RefreshParticipantsCount обновляет количество участников чата
func (h *ChatHandler) RefreshParticipantsCount(ctx context.Context, req *proto.RefreshParticipantsCountRequest) (*proto.RefreshParticipantsCountResponse, error) {
	info, err := h.chatService.RefreshParticipantsCount(ctx, req.ChatId)
	if err != nil {
		return &proto.RefreshParticipantsCountResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.RefreshParticipantsCountResponse{
		ParticipantsCount: int32(info.Count),
	}, nil
}

// Health проверяет состояние сервиса
func (h *ChatHandler) Health(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	return &proto.HealthResponse{
		Status: "OK",
	}, nil
}

