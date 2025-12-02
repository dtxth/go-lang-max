package auth

import (
	"context"
	"fmt"
	"time"

	authpb "auth-service/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient представляет клиент для взаимодействия с Auth Service
type AuthClient struct {
	client authpb.AuthServiceClient
	conn   *grpc.ClientConn
}

// NewAuthClient создает новый клиент Auth Service
func NewAuthClient(authServiceAddr string) (*AuthClient, error) {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	client := authpb.NewAuthServiceClient(conn)
	return &AuthClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close закрывает соединение с Auth Service
func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// AssignRole назначает роль пользователю
func (c *AuthClient) AssignRole(ctx context.Context, userID int64, role string, universityID, branchID, facultyID *int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &authpb.AssignRoleRequest{
		UserId:       userID,
		Role:         role,
		UniversityId: 0,
		BranchId:     0,
		FacultyId:    0,
	}

	if universityID != nil {
		req.UniversityId = *universityID
	}
	if branchID != nil {
		req.BranchId = *branchID
	}
	if facultyID != nil {
		req.FacultyId = *facultyID
	}

	resp, err := c.client.AssignRole(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("auth service error: %s", resp.Error)
	}

	return nil
}

// ValidateToken проверяет токен и возвращает информацию о пользователе
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*authpb.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &authpb.ValidateTokenRequest{
		Token: token,
	}

	resp, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !resp.Valid {
		return nil, fmt.Errorf("invalid token: %s", resp.Error)
	}

	return resp, nil
}

// RevokeUserRoles отзывает все роли пользователя
func (c *AuthClient) RevokeUserRoles(ctx context.Context, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &authpb.RevokeUserRolesRequest{
		UserId: userID,
	}

	resp, err := c.client.RevokeUserRoles(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to revoke user roles: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("auth service error: %s", resp.Error)
	}

	return nil
}
