package grpc

import (
	"context"

	maxbotproto "maxbot-service/api/proto"
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
