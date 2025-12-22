package grpc

import (
	"auth-service/api/proto"
	"auth-service/internal/domain"
	"auth-service/internal/usecase"
	"context"
	"fmt"
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

// Register registers a new user
func (h *AuthHandler) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	var createdUser *domain.User
	var err error

	// Determine if registering by phone or email
	if req.Phone != "" {
		createdUser, err = h.authService.RegisterByPhone(req.Phone, req.Password, req.Role)
	} else if req.Email != "" {
		createdUser, err = h.authService.Register(req.Email, req.Password, req.Role)
	} else {
		return &proto.RegisterResponse{
			Error: "either email or phone is required",
		}, nil
	}

	if err != nil {
		return &proto.RegisterResponse{
			Error: err.Error(),
		}, nil
	}

	// Generate tokens for the newly created user
	var tokens *usecase.TokensWithJTIResult
	if req.Phone != "" {
		tokens, err = h.authService.LoginByIdentifier(req.Phone, req.Password)
	} else {
		tokens, err = h.authService.LoginByIdentifier(req.Email, req.Password)
	}

	if err != nil {
		return &proto.RegisterResponse{
			Error: fmt.Sprintf("user created but login failed: %v", err),
		}, nil
	}

	return &proto.RegisterResponse{
		Tokens: &proto.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
		User: &proto.User{
			Id:    createdUser.ID,
			Email: createdUser.Email,
			Phone: createdUser.Phone,
			Role:  createdUser.Role,
		},
	}, nil
}

// Login authenticates a user by email
func (h *AuthHandler) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	if req.Email == "" {
		return &proto.LoginResponse{
			Error: "email is required",
		}, nil
	}

	if req.Password == "" {
		return &proto.LoginResponse{
			Error: "password is required",
		}, nil
	}

	tokens, err := h.authService.LoginByIdentifier(req.Email, req.Password)
	if err != nil {
		return &proto.LoginResponse{
			Error: err.Error(),
		}, nil
	}

	// Get user info
	userID, email, role, _, err := h.authService.ValidateTokenWithContext(tokens.AccessToken)
	if err != nil {
		return &proto.LoginResponse{
			Error: fmt.Sprintf("login successful but failed to get user info: %v", err),
		}, nil
	}

	return &proto.LoginResponse{
		Tokens: &proto.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
		User: &proto.User{
			Id:    userID,
			Email: email,
			Role:  role,
		},
	}, nil
}

// LoginByPhone authenticates a user by phone
func (h *AuthHandler) LoginByPhone(ctx context.Context, req *proto.LoginByPhoneRequest) (*proto.LoginResponse, error) {
	if req.Phone == "" {
		return &proto.LoginResponse{
			Error: "phone is required",
		}, nil
	}

	if req.Password == "" {
		return &proto.LoginResponse{
			Error: "password is required",
		}, nil
	}

	tokens, err := h.authService.LoginByIdentifier(req.Phone, req.Password)
	if err != nil {
		return &proto.LoginResponse{
			Error: err.Error(),
		}, nil
	}

	// Get user info
	userID, identifier, role, _, err := h.authService.ValidateTokenWithContext(tokens.AccessToken)
	if err != nil {
		return &proto.LoginResponse{
			Error: fmt.Sprintf("login successful but failed to get user info: %v", err),
		}, nil
	}

	return &proto.LoginResponse{
		Tokens: &proto.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
		User: &proto.User{
			Id:    userID,
			Phone: identifier,
			Role:  role,
		},
	}, nil
}

// Refresh refreshes access tokens
func (h *AuthHandler) Refresh(ctx context.Context, req *proto.RefreshRequest) (*proto.RefreshResponse, error) {
	if req.RefreshToken == "" {
		return &proto.RefreshResponse{
			Error: "refresh token is required",
		}, nil
	}

	tokens, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		return &proto.RefreshResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.RefreshResponse{
		Tokens: &proto.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
	}, nil
}

// Logout logs out a user
func (h *AuthHandler) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	if req.RefreshToken == "" {
		return &proto.LogoutResponse{
			Success: false,
			Error:   "refresh token is required",
		}, nil
	}

	err := h.authService.Logout(req.RefreshToken)
	if err != nil {
		return &proto.LogoutResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.LogoutResponse{
		Success: true,
	}, nil
}

// AuthenticateMAX authenticates a user via MAX Mini App
func (h *AuthHandler) AuthenticateMAX(ctx context.Context, req *proto.AuthenticateMAXRequest) (*proto.AuthenticateMAXResponse, error) {
	if req.InitData == "" {
		return &proto.AuthenticateMAXResponse{
			Error: "init_data is required",
		}, nil
	}

	tokens, err := h.authService.AuthenticateMAX(req.InitData)
	if err != nil {
		return &proto.AuthenticateMAXResponse{
			Error: err.Error(),
		}, nil
	}

	// Get user info from the token
	userID, identifier, role, _, err := h.authService.ValidateTokenWithContext(tokens.AccessToken)
	if err != nil {
		return &proto.AuthenticateMAXResponse{
			Error: fmt.Sprintf("authentication successful but failed to get user info: %v", err),
		}, nil
	}

	return &proto.AuthenticateMAXResponse{
		Tokens: &proto.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
		User: &proto.User{
			Id:   userID,
			Role: role,
			// For MAX auth, identifier is in format "max_{id}"
			Email: identifier,
		},
	}, nil
}

// GetBotMe retrieves bot information
func (h *AuthHandler) GetBotMe(ctx context.Context, req *proto.GetBotMeRequest) (*proto.GetBotMeResponse, error) {
	botInfo, err := h.authService.GetBotInfo(ctx)
	if err != nil {
		return &proto.GetBotMeResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetBotMeResponse{
		Bot: &proto.BotInfo{
			Id:        "1", // Default ID since BotInfo doesn't have ID field
			Username:  botInfo.Name,
			FirstName: botInfo.Name,
		},
	}, nil
}

// GetMetrics retrieves service metrics
func (h *AuthHandler) GetMetrics(ctx context.Context, req *proto.GetMetricsRequest) (*proto.GetMetricsResponse, error) {
	metrics := h.authService.GetMetrics()
	if metrics == nil {
		return &proto.GetMetricsResponse{
			Error: "metrics not available",
		}, nil
	}

	snapshot := metrics.GetMetrics()
	metricsMap := make(map[string]int64)
	metricsMap["user_creations"] = snapshot.UserCreations
	metricsMap["password_resets"] = snapshot.PasswordResets
	metricsMap["password_changes"] = snapshot.PasswordChanges
	metricsMap["tokens_generated"] = snapshot.TokensGenerated
	metricsMap["tokens_used"] = snapshot.TokensUsed
	metricsMap["tokens_expired"] = snapshot.TokensExpired
	metricsMap["tokens_invalidated"] = snapshot.TokensInvalidated
	metricsMap["notifications_sent"] = snapshot.NotificationsSent
	metricsMap["notifications_failed"] = snapshot.NotificationsFailed

	return &proto.GetMetricsResponse{
		Metrics: metricsMap,
	}, nil
}

// Health checks service health
func (h *AuthHandler) Health(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	return &proto.HealthResponse{
		Status: "healthy",
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	userID, email, role, tokenCtx, err := h.authService.ValidateTokenWithContext(req.Token)
	if err != nil {
		return &proto.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	resp := &proto.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Email:  email,
		Role:   role,
	}
	
	// Добавляем контекстную информацию, если она есть
	if tokenCtx != nil {
		if tokenCtx.UniversityID != nil {
			resp.UniversityId = *tokenCtx.UniversityID
		}
		if tokenCtx.BranchID != nil {
			resp.BranchId = *tokenCtx.BranchID
		}
		if tokenCtx.FacultyID != nil {
			resp.FacultyId = *tokenCtx.FacultyID
		}
	}

	return resp, nil
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

func (h *AuthHandler) GetUserPermissions(ctx context.Context, req *proto.GetUserPermissionsRequest) (*proto.GetUserPermissionsResponse, error) {
	permissions, err := h.authService.GetUserPermissions(req.UserId)
	if err != nil {
		return &proto.GetUserPermissionsResponse{
			Error: err.Error(),
		}, nil
	}

	var protoPermissions []*proto.UserPermission
	for _, perm := range permissions {
		protoPerm := &proto.UserPermission{
			Id:       perm.ID,
			UserId:   perm.UserID,
			RoleId:   perm.RoleID,
			RoleName: perm.RoleName,
		}
		
		if perm.UniversityID != nil {
			protoPerm.UniversityId = *perm.UniversityID
		}
		if perm.BranchID != nil {
			protoPerm.BranchId = *perm.BranchID
		}
		if perm.FacultyID != nil {
			protoPerm.FacultyId = *perm.FacultyID
		}
		
		protoPermissions = append(protoPermissions, protoPerm)
	}

	return &proto.GetUserPermissionsResponse{
		Permissions: protoPermissions,
	}, nil
}


func (h *AuthHandler) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	// Note: Email field is used for phone number due to proto file mismatch
	userID, err := h.authService.CreateUser(req.Email, req.Password)
	if err != nil {
		return &proto.CreateUserResponse{
			Error: err.Error(),
		}, nil
	}
	
	return &proto.CreateUserResponse{
		UserId: userID,
	}, nil
}

func (h *AuthHandler) AssignRole(ctx context.Context, req *proto.AssignRoleRequest) (*proto.AssignRoleResponse, error) {
	var universityID, branchID, facultyID *int64
	
	if req.UniversityId > 0 {
		universityID = &req.UniversityId
	}
	if req.BranchId > 0 {
		branchID = &req.BranchId
	}
	if req.FacultyId > 0 {
		facultyID = &req.FacultyId
	}
	
	err := h.authService.AssignRoleToUser(req.UserId, req.Role, universityID, branchID, facultyID)
	if err != nil {
		return &proto.AssignRoleResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	return &proto.AssignRoleResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) RevokeUserRoles(ctx context.Context, req *proto.RevokeUserRolesRequest) (*proto.RevokeUserRolesResponse, error) {
	err := h.authService.RevokeAllUserRoles(req.UserId)
	if err != nil {
		return &proto.RevokeUserRolesResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	return &proto.RevokeUserRolesResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *proto.RequestPasswordResetRequest) (*proto.RequestPasswordResetResponse, error) {
	if req.Phone == "" {
		return &proto.RequestPasswordResetResponse{
			Success: false,
			Error:   "phone number is required",
		}, nil
	}

	err := h.authService.RequestPasswordReset(req.Phone)
	if err != nil {
		return &proto.RequestPasswordResetResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.RequestPasswordResetResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *proto.ResetPasswordRequest) (*proto.ResetPasswordResponse, error) {
	if req.Token == "" {
		return &proto.ResetPasswordResponse{
			Success: false,
			Error:   "reset token is required",
		}, nil
	}

	if req.NewPassword == "" {
		return &proto.ResetPasswordResponse{
			Success: false,
			Error:   "new password is required",
		}, nil
	}

	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		return &proto.ResetPasswordResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.ResetPasswordResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) ChangePassword(ctx context.Context, req *proto.ChangePasswordRequest) (*proto.ChangePasswordResponse, error) {
	if req.UserId == 0 {
		return &proto.ChangePasswordResponse{
			Success: false,
			Error:   "user ID is required",
		}, nil
	}

	if req.CurrentPassword == "" {
		return &proto.ChangePasswordResponse{
			Success: false,
			Error:   "current password is required",
		}, nil
	}

	if req.NewPassword == "" {
		return &proto.ChangePasswordResponse{
			Success: false,
			Error:   "new password is required",
		}, nil
	}

	err := h.authService.ChangePassword(req.UserId, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return &proto.ChangePasswordResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.ChangePasswordResponse{
		Success: true,
	}, nil
}
