package usecase

import (
	"auth-service/internal/domain"
	"testing"
)

// Mock repositories for testing
type mockUserRoleRepository struct {
	roles []*domain.UserRoleWithDetails
}

func (m *mockUserRoleRepository) Create(ur *domain.UserRole) error {
	return nil
}

func (m *mockUserRoleRepository) GetByUserID(userID int64) ([]*domain.UserRoleWithDetails, error) {
	return m.roles, nil
}

func (m *mockUserRoleRepository) Delete(id int64) error {
	return nil
}

func (m *mockUserRoleRepository) DeleteByUserID(userID int64) error {
	return nil
}

func (m *mockUserRoleRepository) GetByUserIDAndRole(userID int64, roleName string) (*domain.UserRoleWithDetails, error) {
	for _, role := range m.roles {
		if role.RoleName == roleName {
			return role, nil
		}
	}
	return nil, nil
}

type mockRoleRepository struct{}

func (m *mockRoleRepository) GetByName(name string) (*domain.Role, error) {
	return &domain.Role{ID: 1, Name: name}, nil
}

func (m *mockRoleRepository) GetByID(id int64) (*domain.Role, error) {
	return &domain.Role{ID: id, Name: "test"}, nil
}

func (m *mockRoleRepository) List() ([]*domain.Role, error) {
	return []*domain.Role{}, nil
}

func TestValidatePermission_Superadmin(t *testing.T) {
	userRoleRepo := &mockUserRoleRepository{
		roles: []*domain.UserRoleWithDetails{
			{
				UserRole: domain.UserRole{
					ID:     1,
					UserID: 1,
					RoleID: 1,
				},
				RoleName: "superadmin",
			},
		},
	}
	
	roleRepo := &mockRoleRepository{}
	uc := NewValidatePermissionUseCase(userRoleRepo, roleRepo)
	
	ctx := &domain.PermissionContext{
		UserID:               1,
		Role:                 "superadmin",
		ResourceUniversityID: int64Ptr(100),
	}
	
	permission := &domain.Permission{
		Resource: "chat",
		Action:   "read",
	}
	
	hasPermission, err := uc.ValidatePermission(ctx, permission)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !hasPermission {
		t.Error("Superadmin should have access to all resources")
	}
}

func TestValidatePermission_Curator_SameUniversity(t *testing.T) {
	universityID := int64(100)
	
	userRoleRepo := &mockUserRoleRepository{
		roles: []*domain.UserRoleWithDetails{
			{
				UserRole: domain.UserRole{
					ID:           1,
					UserID:       2,
					RoleID:       2,
					UniversityID: &universityID,
				},
				RoleName: "curator",
			},
		},
	}
	
	roleRepo := &mockRoleRepository{}
	uc := NewValidatePermissionUseCase(userRoleRepo, roleRepo)
	
	ctx := &domain.PermissionContext{
		UserID:               2,
		Role:                 "curator",
		UniversityID:         &universityID,
		ResourceUniversityID: &universityID,
	}
	
	permission := &domain.Permission{
		Resource: "chat",
		Action:   "read",
	}
	
	hasPermission, err := uc.ValidatePermission(ctx, permission)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !hasPermission {
		t.Error("Curator should have access to resources in their university")
	}
}

func TestValidatePermission_Curator_DifferentUniversity(t *testing.T) {
	curatorUniversityID := int64(100)
	resourceUniversityID := int64(200)
	
	userRoleRepo := &mockUserRoleRepository{
		roles: []*domain.UserRoleWithDetails{
			{
				UserRole: domain.UserRole{
					ID:           1,
					UserID:       2,
					RoleID:       2,
					UniversityID: &curatorUniversityID,
				},
				RoleName: "curator",
			},
		},
	}
	
	roleRepo := &mockRoleRepository{}
	uc := NewValidatePermissionUseCase(userRoleRepo, roleRepo)
	
	ctx := &domain.PermissionContext{
		UserID:               2,
		Role:                 "curator",
		UniversityID:         &curatorUniversityID,
		ResourceUniversityID: &resourceUniversityID,
	}
	
	permission := &domain.Permission{
		Resource: "chat",
		Action:   "read",
	}
	
	hasPermission, err := uc.ValidatePermission(ctx, permission)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if hasPermission {
		t.Error("Curator should not have access to resources in different university")
	}
}

func TestValidatePermission_Operator_SameBranch(t *testing.T) {
	universityID := int64(100)
	branchID := int64(10)
	
	userRoleRepo := &mockUserRoleRepository{
		roles: []*domain.UserRoleWithDetails{
			{
				UserRole: domain.UserRole{
					ID:           1,
					UserID:       3,
					RoleID:       3,
					UniversityID: &universityID,
					BranchID:     &branchID,
				},
				RoleName: "operator",
			},
		},
	}
	
	roleRepo := &mockRoleRepository{}
	uc := NewValidatePermissionUseCase(userRoleRepo, roleRepo)
	
	ctx := &domain.PermissionContext{
		UserID:            3,
		Role:              "operator",
		UniversityID:      &universityID,
		BranchID:          &branchID,
		ResourceBranchID:  &branchID,
	}
	
	permission := &domain.Permission{
		Resource: "chat",
		Action:   "read",
	}
	
	hasPermission, err := uc.ValidatePermission(ctx, permission)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !hasPermission {
		t.Error("Operator should have access to resources in their branch")
	}
}

func TestValidatePermission_Operator_DifferentBranch(t *testing.T) {
	universityID := int64(100)
	operatorBranchID := int64(10)
	resourceBranchID := int64(20)
	
	userRoleRepo := &mockUserRoleRepository{
		roles: []*domain.UserRoleWithDetails{
			{
				UserRole: domain.UserRole{
					ID:           1,
					UserID:       3,
					RoleID:       3,
					UniversityID: &universityID,
					BranchID:     &operatorBranchID,
				},
				RoleName: "operator",
			},
		},
	}
	
	roleRepo := &mockRoleRepository{}
	uc := NewValidatePermissionUseCase(userRoleRepo, roleRepo)
	
	ctx := &domain.PermissionContext{
		UserID:            3,
		Role:              "operator",
		UniversityID:      &universityID,
		BranchID:          &operatorBranchID,
		ResourceBranchID:  &resourceBranchID,
	}
	
	permission := &domain.Permission{
		Resource: "chat",
		Action:   "read",
	}
	
	hasPermission, err := uc.ValidatePermission(ctx, permission)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if hasPermission {
		t.Error("Operator should not have access to resources in different branch")
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
