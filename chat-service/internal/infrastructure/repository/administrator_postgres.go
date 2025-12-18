package repository

import (
	"chat-service/internal/domain"
	"chat-service/internal/infrastructure/database"
	"database/sql"
	"fmt"
)

type AdministratorPostgres struct {
	db  *database.DB
	dsn string
}

func NewAdministratorPostgres(db *database.DB) *AdministratorPostgres {
	return &AdministratorPostgres{db: db}
}

func NewAdministratorPostgresWithDSN(db *database.DB, dsn string) *AdministratorPostgres {
	return &AdministratorPostgres{db: db, dsn: dsn}
}

func (r *AdministratorPostgres) Create(admin *domain.Administrator) error {
	db := r.db
	err := db.QueryRow(
		`INSERT INTO administrators (chat_id, phone, max_id, add_user, add_admin) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		admin.ChatID, admin.Phone, admin.MaxID, admin.AddUser, admin.AddAdmin,
	).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)
	return err
}

func (r *AdministratorPostgres) GetByID(id int64) (*domain.Administrator, error) {
	db := r.db
	admin := &domain.Administrator{}
	err := db.QueryRow(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE id = $1`,
		id,
	).Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
		&admin.AddUser, &admin.AddAdmin, &admin.CreatedAt, &admin.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAdministratorNotFound
		}
		return nil, err
	}
	return admin, nil
}

func (r *AdministratorPostgres) GetByChatID(chatID int64) ([]*domain.Administrator, error) {
	db := r.db
	rows, err := db.Query(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE chat_id = $1`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var administrators []*domain.Administrator
	for rows.Next() {
		admin := &domain.Administrator{}
		err := rows.Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
			&admin.AddUser, &admin.AddAdmin, &admin.CreatedAt, &admin.UpdatedAt)
		if err != nil {
			return nil, err
		}
		administrators = append(administrators, admin)
	}
	return administrators, rows.Err()
}

func (r *AdministratorPostgres) Update(admin *domain.Administrator) error {
	db := r.db
	_, err := db.Exec(
		`UPDATE administrators SET phone = $1, max_id = $2, add_user = $3, add_admin = $4 
		 WHERE id = $5`,
		admin.Phone, admin.MaxID, admin.AddUser, admin.AddAdmin, admin.ID,
	)
	return err
}

func (r *AdministratorPostgres) Delete(id int64) error {
	db := r.db
	_, err := db.Exec(`DELETE FROM administrators WHERE id = $1`, id)
	return err
}

func (r *AdministratorPostgres) GetByPhone(phone string) (*domain.Administrator, error) {
	db := r.db
	admin := &domain.Administrator{}
	err := db.QueryRow(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE phone = $1`,
		phone,
	).Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
		&admin.AddUser, &admin.AddAdmin, &admin.CreatedAt, &admin.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAdministratorNotFound
		}
		return nil, err
	}
	return admin, nil
}

func (r *AdministratorPostgres) CountByChatID(chatID int64) (int, error) {
	db := r.db
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM administrators WHERE chat_id = $1`, chatID).Scan(&count)
	return count, err
}

func (r *AdministratorPostgres) GetAll(search string, limit, offset int) ([]*domain.Administrator, int, error) {
	db := r.db
	
	// Build query with search
	query := `SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		      FROM administrators`
	countQuery := `SELECT COUNT(*) FROM administrators`
	args := []interface{}{}
	
	if search != "" {
		query += ` WHERE phone ILIKE $1 OR max_id ILIKE $1`
		countQuery += ` WHERE phone ILIKE $1 OR max_id ILIKE $1`
		args = append(args, "%"+search+"%")
	}
	
	// Get total count
	var total int
	countArgs := args
	err := db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get records
	query += ` ORDER BY id LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var administrators []*domain.Administrator
	for rows.Next() {
		admin := &domain.Administrator{}
		err := rows.Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
			&admin.AddUser, &admin.AddAdmin, &admin.CreatedAt, &admin.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		administrators = append(administrators, admin)
	}
	return administrators, total, rows.Err()
}

func (r *AdministratorPostgres) GetByPhoneAndChatID(phone string, chatID int64) (*domain.Administrator, error) {
	db := r.db
	admin := &domain.Administrator{}
	err := db.QueryRow(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE phone = $1 AND chat_id = $2`,
		phone, chatID,
	).Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
		&admin.AddUser, &admin.AddAdmin, &admin.CreatedAt, &admin.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAdministratorNotFound
		}
		return nil, err
	}
	return admin, nil
}