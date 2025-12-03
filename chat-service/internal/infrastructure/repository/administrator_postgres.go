package repository

import (
	"chat-service/internal/domain"
	"database/sql"
)

type AdministratorPostgres struct {
	db *sql.DB
}

func NewAdministratorPostgres(db *sql.DB) *AdministratorPostgres {
	return &AdministratorPostgres{db: db}
}

func (r *AdministratorPostgres) Create(admin *domain.Administrator) error {
	err := r.db.QueryRow(
		`INSERT INTO administrators (chat_id, phone, max_id, add_user, add_admin) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		admin.ChatID, admin.Phone, admin.MaxID, admin.AddUser, admin.AddAdmin,
	).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)
	return err
}

func (r *AdministratorPostgres) GetByID(id int64) (*domain.Administrator, error) {
	admin := &domain.Administrator{}
	err := r.db.QueryRow(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE id = $1`,
		id,
	).Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
		&admin.AddUser, &admin.AddAdmin,
		&admin.CreatedAt, &admin.UpdatedAt)
	return admin, err
}

func (r *AdministratorPostgres) GetByChatID(chatID int64) ([]*domain.Administrator, error) {
	rows, err := r.db.Query(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE chat_id = $1 ORDER BY created_at`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var administrators []*domain.Administrator
	for rows.Next() {
		admin := &domain.Administrator{}
		err := rows.Scan(
			&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
			&admin.AddUser, &admin.AddAdmin,
			&admin.CreatedAt, &admin.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		administrators = append(administrators, admin)
	}
	return administrators, rows.Err()
}

func (r *AdministratorPostgres) GetByPhoneAndChatID(phone string, chatID int64) (*domain.Administrator, error) {
	admin := &domain.Administrator{}
	err := r.db.QueryRow(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE phone = $1 AND chat_id = $2`,
		phone, chatID,
	).Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
		&admin.AddUser, &admin.AddAdmin,
		&admin.CreatedAt, &admin.UpdatedAt)
	return admin, err
}

func (r *AdministratorPostgres) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM administrators WHERE id = $1`, id)
	return err
}

func (r *AdministratorPostgres) CountByChatID(chatID int64) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM administrators WHERE chat_id = $1`,
		chatID,
	).Scan(&count)
	return count, err
}

