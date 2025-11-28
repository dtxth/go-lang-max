package auth

import (
	"context"
	"time"

	"chat-service/internal/domain"
	grpcretry "chat-service/internal/infrastructure/grpc"
	authproto "auth-service/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	conn    *grpc.ClientConn
	client  authproto.AuthServiceClient
	timeout time.Duration
}

func NewAuthClient(address string, timeout time.Duration) (*AuthClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		conn:    conn,
		client:  authproto.NewAuthServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *AuthClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *AuthClient) ValidateToken(token string) (*domain.TokenInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	var resp *authproto.ValidateTokenResponse
	err := grpcretry.WithRetry(ctx, "Auth.ValidateToken", func() error {
		var callErr error
		resp, callErr = c.client.ValidateToken(ctx, &authproto.ValidateTokenRequest{Token: token})
		return callErr
	})
	
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, domain.ErrInvalidToken
	}

	if !resp.Valid {
		return nil, domain.ErrInvalidToken
	}

	tokenInfo := &domain.TokenInfo{
		Valid:  resp.Valid,
		UserID: resp.UserId,
		Email:  resp.Email,
		Role:   resp.Role,
	}

	// Устанавливаем опциональные поля только если они не равны 0
	if resp.UniversityId != 0 {
		tokenInfo.UniversityID = &resp.UniversityId
	}
	if resp.BranchId != 0 {
		tokenInfo.BranchID = &resp.BranchId
	}
	if resp.FacultyId != 0 {
		tokenInfo.FacultyID = &resp.FacultyId
	}

	return tokenInfo, nil
}
