package repository

import (
	"auth-service/internal/domain"
	"database/sql"
)

type RolePostgres struct {
	db *sql.DB
}

func NewRolePostgres(db *sql.DB) *RolePostgres {
	return &RolePostgres{db: db}
}

func (r *RolePostgres) GetByName(name string) (*domain.Role, error) {
	role := &domain.Role{}
	err := r.db.QueryRow(
		`SELECT id, name, description, created_at FROM roles WHERE name = $1`,
		name,
	).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return role, nil
}

func (r *RolePostgres) GetByID(id int64) (*domain.Role, error) {
	role := &domain.Role{}
	err := r.db.QueryRow(
		`SELECT id, name, description, created_at FROM roles WHERE id = $1`,
		id,
	).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return role, nil
}

func (r *RolePostgres) List() ([]*domain.Role, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, created_at FROM roles ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	
	return roles, rows.Err()
}
