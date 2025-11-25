package repository

import (
	"auth-service/internal/domain"
	"database/sql"
)

type UserPostgres struct {
    db *sql.DB
}

func NewUserPostgres(db *sql.DB) *UserPostgres {
    return &UserPostgres{db: db}
}

func (r *UserPostgres) Create(u *domain.User) error {
    // Если роль не указана, устанавливаем роль по умолчанию
    if u.Role == "" {
        u.Role = domain.RoleOperator
    }
    return r.db.
        QueryRow(`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id`,
            u.Email, u.Password, u.Role,
        ).Scan(&u.ID)
}

func (r *UserPostgres) GetByEmail(email string) (*domain.User, error) {
    user := &domain.User{}
    err := r.db.QueryRow(`SELECT id, email, password_hash, role FROM users WHERE email=$1`, email).
        Scan(&user.ID, &user.Email, &user.Password, &user.Role)
    return user, err
}

func (r *UserPostgres) GetByID(id int64) (*domain.User, error) {
    user := &domain.User{}
    err := r.db.QueryRow(`SELECT id, email, password_hash, role FROM users WHERE id=$1`, id).
        Scan(&user.ID, &user.Email, &user.Password, &user.Role)
    return user, err
}