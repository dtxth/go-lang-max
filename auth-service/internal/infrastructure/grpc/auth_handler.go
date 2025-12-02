package grpc

import (
	"auth-service/internal/usecase"
	"auth-service/api/proto"
	"context"
)

type AuthHandler struct {
	authService *usecase.AuthService
	proto.UnimplementedAuthServiceServer
}

func NewAuthHandler(authService *usecase.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	userID, email, role, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		return &proto.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &proto.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Email:  email,
		Role:   role,
	}, nil
}

func (h *AuthHandler) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	user, err := h.authService.GetUserByID(req.UserId)
	if err != nil {
		return &proto.GetUserResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetUserResponse{
		Id:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

