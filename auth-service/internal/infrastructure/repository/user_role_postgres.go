package repository

import (
	"auth-service/internal/domain"
	"database/sql"
)

type UserRolePostgres struct {
	db *sql.DB
}

func NewUserRolePostgres(db *sql.DB) *UserRolePostgres {
	return &UserRolePostgres{db: db}
}

func (r *UserRolePostgres) Create(ur *domain.UserRole) error {
	return r.db.QueryRow(
		`INSERT INTO user_roles (user_id, role_id, university_id, branch_id, faculty_id, assigned_by, assigned_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		ur.UserID, ur.RoleID, ur.UniversityID, ur.BranchID, ur.FacultyID, ur.AssignedBy, ur.AssignedAt,
	).Scan(&ur.ID)
}

func (r *UserRolePostgres) GetByUserID(userID int64) ([]*domain.UserRoleWithDetails, error) {
	rows, err := r.db.Query(
		`SELECT ur.id, ur.user_id, ur.role_id, ur.university_id, ur.branch_id, ur.faculty_id, 
		        ur.assigned_by, ur.assigned_at, ro.name
		 FROM user_roles ur
		 JOIN roles ro ON ur.role_id = ro.id
		 WHERE ur.user_id = $1
		 ORDER BY ur.assigned_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var userRoles []*domain.UserRoleWithDetails
	for rows.Next() {
		ur := &domain.UserRoleWithDetails{}
		if err := rows.Scan(
			&ur.ID, &ur.UserID, &ur.RoleID, &ur.UniversityID, &ur.BranchID, &ur.FacultyID,
			&ur.AssignedBy, &ur.AssignedAt, &ur.RoleName,
		); err != nil {
			return nil, err
		}
		userRoles = append(userRoles, ur)
	}
	
	return userRoles, rows.Err()
}

func (r *UserRolePostgres) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM user_roles WHERE id = $1`, id)
	return err
}

func (r *UserRolePostgres) DeleteByUserID(userID int64) error {
	_, err := r.db.Exec(`DELETE FROM user_roles WHERE user_id = $1`, userID)
	return err
}

func (r *UserRolePostgres) GetByUserIDAndRole(userID int64, roleName string) (*domain.UserRoleWithDetails, error) {
	ur := &domain.UserRoleWithDetails{}
	err := r.db.QueryRow(
		`SELECT ur.id, ur.user_id, ur.role_id, ur.university_id, ur.branch_id, ur.faculty_id,
		        ur.assigned_by, ur.assigned_at, ro.name
		 FROM user_roles ur
		 JOIN roles ro ON ur.role_id = ro.id
		 WHERE ur.user_id = $1 AND ro.name = $2
		 LIMIT 1`,
		userID, roleName,
	).Scan(
		&ur.ID, &ur.UserID, &ur.RoleID, &ur.UniversityID, &ur.BranchID, &ur.FacultyID,
		&ur.AssignedBy, &ur.AssignedAt, &ur.RoleName,
	)
	
	if err != nil {
		return nil, err
	}
	
	return ur, nil
}

func (r *UserRolePostgres) GetRoleByName(name string) (*domain.Role, error) {
	role := &domain.Role{}
	err := r.db.QueryRow(
		`SELECT id, name, description, created_at
		 FROM roles
		 WHERE name = $1`,
		name,
	).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return role, nil
}
