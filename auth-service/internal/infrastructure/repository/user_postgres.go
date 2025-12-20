package repository

import (
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/database"
	"log"
)

type UserPostgres struct {
    db *database.DB
}

func NewUserPostgres(db *database.DB) *UserPostgres {
    return &UserPostgres{db: db}
}

func (r *UserPostgres) Create(u *domain.User) error {
    // Если роль не указана, устанавливаем роль по умолчанию
    if u.Role == "" {
        u.Role = domain.RoleOperator
    }
    return r.db.
        QueryRow(`INSERT INTO users (phone, email, password_hash, role, max_id, username, name) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
            u.Phone, u.Email, u.Password, u.Role, u.MaxID, u.Username, u.Name,
        ).Scan(&u.ID)
}

func (r *UserPostgres) GetByPhone(phone string) (*domain.User, error) {
    user := &domain.User{}
    query := `SELECT id, phone, email, password_hash, role, max_id, username, name FROM users WHERE phone=$1`
    err := r.db.QueryRow(query, phone).
        Scan(&user.ID, &user.Phone, &user.Email, &user.Password, &user.Role, &user.MaxID, &user.Username, &user.Name)
    
    // Добавим логирование для отладки
    if err != nil {
        log.Printf("[DEBUG] GetByPhone failed for phone=%s, error=%v", phone, err)
    } else {
        log.Printf("[DEBUG] GetByPhone success for phone=%s, found user ID=%d", phone, user.ID)
    }
    
    return user, err
}

func (r *UserPostgres) GetByEmail(email string) (*domain.User, error) {
    user := &domain.User{}
    err := r.db.QueryRow(`SELECT id, phone, email, password_hash, role, max_id, username, name FROM users WHERE email=$1`, email).
        Scan(&user.ID, &user.Phone, &user.Email, &user.Password, &user.Role, &user.MaxID, &user.Username, &user.Name)
    return user, err
}

func (r *UserPostgres) GetByID(id int64) (*domain.User, error) {
    user := &domain.User{}
    err := r.db.QueryRow(`SELECT id, phone, email, password_hash, role, max_id, username, name FROM users WHERE id=$1`, id).
        Scan(&user.ID, &user.Phone, &user.Email, &user.Password, &user.Role, &user.MaxID, &user.Username, &user.Name)
    return user, err
}

func (r *UserPostgres) Update(u *domain.User) error {
    _, err := r.db.Exec(
        `UPDATE users SET phone=$1, email=$2, password_hash=$3, role=$4, max_id=$5, username=$6, name=$7 WHERE id=$8`,
        u.Phone, u.Email, u.Password, u.Role, u.MaxID, u.Username, u.Name, u.ID,
    )
    return err
}

// GetByMaxID retrieves a user by their MAX platform ID
func (r *UserPostgres) GetByMaxID(maxID int64) (*domain.User, error) {
    user := &domain.User{}
    err := r.db.QueryRow(`SELECT id, phone, email, password_hash, role, max_id, username, name FROM users WHERE max_id=$1`, maxID).
        Scan(&user.ID, &user.Phone, &user.Email, &user.Password, &user.Role, &user.MaxID, &user.Username, &user.Name)
    return user, err
}