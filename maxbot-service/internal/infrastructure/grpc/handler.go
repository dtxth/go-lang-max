package grpc

import (
	"context"

	maxbotproto "maxbot-service/api/proto/maxbotproto"
	"maxbot-service/internal/domain"
	"maxbot-service/internal/usecase"
)

type MaxBotHandler struct {
	service *usecase.MaxBotService
	maxbotproto.UnimplementedMaxBotServiceServer
}

func NewMaxBotHandler(service *usecase.MaxBotService) *MaxBotHandler {
	return &MaxBotHandler{service: service}
}

func (h *MaxBotHandler) GetMaxIDByPhone(ctx context.Context, req *maxbotproto.GetMaxIDByPhoneRequest) (*maxbotproto.GetMaxIDByPhoneResponse, error) {
	maxID, err := h.service.GetMaxIDByPhone(ctx, req.GetPhone())
	if err != nil {
		return &maxbotproto.GetMaxIDByPhoneResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	return &maxbotproto.GetMaxIDByPhoneResponse{MaxId: maxID}, nil
}

func (h *MaxBotHandler) ValidatePhone(ctx context.Context, req *maxbotproto.ValidatePhoneRequest) (*maxbotproto.ValidatePhoneResponse, error) {
	valid, normalized, err := h.service.ValidatePhone(req.GetPhone())
	if err != nil {
		return &maxbotproto.ValidatePhoneResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	return &maxbotproto.ValidatePhoneResponse{
		Valid:           valid,
		NormalizedPhone: normalized,
	}, nil
}

func (h *MaxBotHandler) SendMessage(ctx context.Context, req *maxbotproto.SendMessageRequest) (*maxbotproto.SendMessageResponse, error) {
	var chatID, userID int64
	
	switch recipient := req.Recipient.(type) {
	case *maxbotproto.SendMessageRequest_ChatId:
		chatID = recipient.ChatId
	case *maxbotproto.SendMessageRequest_UserId:
		userID = recipient.UserId
	}

	messageID, err := h.service.SendMessage(ctx, chatID, userID, req.GetText())
	if err != nil {
		return &maxbotproto.SendMessageResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	return &maxbotproto.SendMessageResponse{MessageId: messageID}, nil
}

func (h *MaxBotHandler) SendNotification(ctx context.Context, req *maxbotproto.SendNotificationRequest) (*maxbotproto.SendNotificationResponse, error) {
	err := h.service.SendNotification(ctx, req.GetPhone(), req.GetText())
	if err != nil {
		return &maxbotproto.SendNotificationResponse{
			Success:   false,
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	return &maxbotproto.SendNotificationResponse{Success: true}, nil
}

func (h *MaxBotHandler) GetChatInfo(ctx context.Context, req *maxbotproto.GetChatInfoRequest) (*maxbotproto.GetChatInfoResponse, error) {
	chatInfo, err := h.service.GetChatInfo(ctx, req.GetChatId())
	if err != nil {
		return &maxbotproto.GetChatInfoResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	return &maxbotproto.GetChatInfoResponse{
		Chat: &maxbotproto.ChatInfo{
			ChatId:            chatInfo.ChatID,
			Title:             chatInfo.Title,
			Type:              chatInfo.Type,
			ParticipantsCount: int32(chatInfo.ParticipantsCount),
			Description:       chatInfo.Description,
		},
	}, nil
}

func (h *MaxBotHandler) GetChatMembers(ctx context.Context, req *maxbotproto.GetChatMembersRequest) (*maxbotproto.GetChatMembersResponse, error) {
	membersList, err := h.service.GetChatMembers(ctx, req.GetChatId(), int(req.GetLimit()), req.GetMarker())
	if err != nil {
		return &maxbotproto.GetChatMembersResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	members := make([]*maxbotproto.ChatMember, 0, len(membersList.Members))
	for _, member := range membersList.Members {
		members = append(members, &maxbotproto.ChatMember{
			UserId:  member.UserID,
			Name:    member.Name,
			IsAdmin: member.IsAdmin,
			IsOwner: member.IsOwner,
		})
	}

	return &maxbotproto.GetChatMembersResponse{
		Members: members,
		Marker:  membersList.Marker,
	}, nil
}

func (h *MaxBotHandler) GetChatAdmins(ctx context.Context, req *maxbotproto.GetChatAdminsRequest) (*maxbotproto.GetChatAdminsResponse, error) {
	admins, err := h.service.GetChatAdmins(ctx, req.GetChatId())
	if err != nil {
		return &maxbotproto.GetChatAdminsResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	adminsList := make([]*maxbotproto.ChatMember, 0, len(admins))
	for _, admin := range admins {
		adminsList = append(adminsList, &maxbotproto.ChatMember{
			UserId:  admin.UserID,
			Name:    admin.Name,
			IsAdmin: admin.IsAdmin,
			IsOwner: admin.IsOwner,
		})
	}

	return &maxbotproto.GetChatAdminsResponse{Admins: adminsList}, nil
}

func (h *MaxBotHandler) CheckPhoneNumbers(ctx context.Context, req *maxbotproto.CheckPhoneNumbersRequest) (*maxbotproto.CheckPhoneNumbersResponse, error) {
	existingPhones, err := h.service.CheckPhoneNumbers(ctx, req.GetPhones())
	if err != nil {
		return &maxbotproto.CheckPhoneNumbersResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	return &maxbotproto.CheckPhoneNumbersResponse{ExistingPhones: existingPhones}, nil
}

func (h *MaxBotHandler) NormalizePhone(ctx context.Context, req *maxbotproto.NormalizePhoneRequest) (*maxbotproto.NormalizePhoneResponse, error) {
	normalized, err := h.service.NormalizePhone(req.GetPhone())
	if err != nil {
		return &maxbotproto.NormalizePhoneResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	return &maxbotproto.NormalizePhoneResponse{NormalizedPhone: normalized}, nil
}

func (h *MaxBotHandler) BatchGetUsersByPhone(ctx context.Context, req *maxbotproto.BatchGetUsersByPhoneRequest) (*maxbotproto.BatchGetUsersByPhoneResponse, error) {
	mappings, err := h.service.BatchGetUsersByPhone(ctx, req.GetPhones())
	if err != nil {
		return &maxbotproto.BatchGetUsersByPhoneResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	protoMappings := make([]*maxbotproto.UserPhoneMapping, 0, len(mappings))
	for _, mapping := range mappings {
		protoMappings = append(protoMappings, &maxbotproto.UserPhoneMapping{
			Phone: mapping.Phone,
			MaxId: mapping.MaxID,
			Found: mapping.Found,
		})
	}

	return &maxbotproto.BatchGetUsersByPhoneResponse{Mappings: protoMappings}, nil
}

func (h *MaxBotHandler) GetMe(ctx context.Context, req *maxbotproto.GetMeRequest) (*maxbotproto.GetMeResponse, error) {
	botInfo, err := h.service.GetMe(ctx)
	if err != nil {
		return &maxbotproto.GetMeResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	return &maxbotproto.GetMeResponse{
		Bot: &maxbotproto.BotInfo{
			Name:    botInfo.Name,
			AddLink: botInfo.AddLink,
		},
	}, nil
}

func (h *MaxBotHandler) GetInternalUsers(ctx context.Context, req *maxbotproto.GetInternalUsersRequest) (*maxbotproto.GetInternalUsersResponse, error) {
	if len(req.GetPhoneNumbers()) == 0 {
		return &maxbotproto.GetInternalUsersResponse{
			Users:               []*maxbotproto.InternalUser{},
			FailedPhoneNumbers:  []string{},
			ErrorCode:           maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED,
		}, nil
	}

	if len(req.GetPhoneNumbers()) > 100 {
		return &maxbotproto.GetInternalUsersResponse{
			Error:     "batch size exceeds maximum of 100 phones",
			ErrorCode: maxbotproto.ErrorCode_ERROR_CODE_INTERNAL,
		}, nil
	}

	users, failedPhones, err := h.service.GetInternalUsers(ctx, req.GetPhoneNumbers())
	if err != nil {
		return &maxbotproto.GetInternalUsersResponse{
			Error:     err.Error(),
			ErrorCode: mapError(err),
		}, nil
	}

	// Convert domain users to protobuf users
	protoUsers := make([]*maxbotproto.InternalUser, 0, len(users))
	for _, user := range users {
		protoUsers = append(protoUsers, &maxbotproto.InternalUser{
			UserId:        user.UserID,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			IsBot:         user.IsBot,
			Username:      user.Username,
			AvatarUrl:     user.AvatarURL,
			FullAvatarUrl: user.FullAvatarURL,
			Link:          user.Link,
			PhoneNumber:   user.PhoneNumber,
		})
	}

	return &maxbotproto.GetInternalUsersResponse{
		Users:              protoUsers,
		FailedPhoneNumbers: failedPhones,
		ErrorCode:          maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED,
	}, nil
}

// TODO: Uncomment when protobuf files are regenerated with UserProfile definitions
// func (h *MaxBotHandler) GetUserProfileByPhone(ctx context.Context, req *maxbotproto.GetUserProfileByPhoneRequest) (*maxbotproto.GetUserProfileByPhoneResponse, error) {
// 	if req.Phone == "" {
// 		return &maxbotproto.GetUserProfileByPhoneResponse{
// 			ErrorCode: maxbotproto.ErrorCode_ERROR_CODE_INVALID_PHONE,
// 			Error:     "phone number is required",
// 		}, nil
// 	}

// 	profile, err := h.service.GetUserProfileByPhone(ctx, req.Phone)
// 	if err != nil {
// 		return &maxbotproto.GetUserProfileByPhoneResponse{
// 			ErrorCode: mapError(err),
// 			Error:     err.Error(),
// 		}, nil
// 	}

// 	return &maxbotproto.GetUserProfileByPhoneResponse{
// 		Profile: &maxbotproto.UserProfile{
// 			MaxId:     profile.MaxID,
// 			FirstName: profile.FirstName,
// 			LastName:  profile.LastName,
// 			Phone:     profile.Phone,
// 		},
// 		ErrorCode: maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED,
// 	}, nil
// }

func mapError(err error) maxbotproto.ErrorCode {
	switch err {
	case domain.ErrInvalidPhone:
		return maxbotproto.ErrorCode_ERROR_CODE_INVALID_PHONE
	case domain.ErrMaxIDNotFound:
		return maxbotproto.ErrorCode_ERROR_CODE_MAX_ID_NOT_FOUND
	default:
		return maxbotproto.ErrorCode_ERROR_CODE_INTERNAL
	}
}
