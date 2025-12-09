package grpc

import (
	"auth-service/api/proto"
	"auth-service/internal/usecase"
	"context"
	"log"
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
	userID, err := h.authService.CreateUser(req.Phone, req.Password)
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
