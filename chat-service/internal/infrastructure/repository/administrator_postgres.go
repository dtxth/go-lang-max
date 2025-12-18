package repository

import (
	"chat-service/internal/domain"
	"database/sql"
	"fmt"
)

type AdministratorPostgres struct {
	db  *sql.DB
	dsn string
}

func NewAdministratorPostgres(db *sql.DB) *AdministratorPostgres {
	return &AdministratorPostgres{db: db}
}

func NewAdministratorPostgresWithDSN(db *sql.DB, dsn string) *AdministratorPostgres {
	return &AdministratorPostgres{db: db, dsn: dsn}
}

// getDB returns a working database connection
func (r *AdministratorPostgres) getDB() (*sql.DB, error) {
	if r.db == nil {
		return nil, fmt.Errorf("no database connection available")
	}
	return r.db, nil
}

func (r *AdministratorPostgres) Create(admin *domain.Administrator) error {
	db, err := r.getDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	
	err = db.QueryRow(
		`INSERT INTO administrators (chat_id, phone, max_id, add_user, add_admin) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		admin.ChatID, admin.Phone, admin.MaxID, admin.AddUser, admin.AddAdmin,
	).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)
	return err
}

func (r *AdministratorPostgres) GetByID(id int64) (*domain.Administrator, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	
	admin := &domain.Administrator{}
	err = db.QueryRow(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE id = $1`,
		id,
	).Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
		&admin.AddUser, &admin.AddAdmin,
		&admin.CreatedAt, &admin.UpdatedAt)
	return admin, err
}

func (r *AdministratorPostgres) GetByChatID(chatID int64) ([]*domain.Administrator, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	
	rows, err := db.Query(
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
	db, err := r.getDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	
	admin := &domain.Administrator{}
	err = db.QueryRow(
		`SELECT id, chat_id, phone, max_id, add_user, add_admin, created_at, updated_at 
		 FROM administrators WHERE phone = $1 AND chat_id = $2`,
		phone, chatID,
	).Scan(&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
		&admin.AddUser, &admin.AddAdmin,
		&admin.CreatedAt, &admin.UpdatedAt)
	return admin, err
}

func (r *AdministratorPostgres) Delete(id int64) error {
	db, err := r.getDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	
	_, err = db.Exec(`DELETE FROM administrators WHERE id = $1`, id)
	return err
}

func (r *AdministratorPostgres) CountByChatID(chatID int64) (int, error) {
	db, err := r.getDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}
	
	var count int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM administrators WHERE chat_id = $1`,
		chatID,
	).Scan(&count)
	return count, err
}

func (r *AdministratorPostgres) GetAll(query string, limit, offset int) ([]*domain.Administrator, int, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get database connection: %w", err)
	}
	
	// Устанавливаем значения по умолчанию
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Базовый SQL запрос
	baseQuery := `FROM administrators a 
		LEFT JOIN chats c ON a.chat_id = c.id`
	
	whereClause := ""
	args := []interface{}{}

	// Добавляем поиск, если указан query
	if query != "" {
		whereClause = ` WHERE (a.phone ILIKE $1 OR a.max_id ILIKE $1 OR c.name ILIKE $1)`
		args = append(args, "%"+query+"%")
	}

	// Подсчитываем общее количество
	var totalCount int
	countQuery := `SELECT COUNT(*) ` + baseQuery + whereClause
	err = db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией
	var selectQuery string
	if query != "" {
		selectQuery = `SELECT a.id, a.chat_id, a.phone, a.max_id, a.add_user, a.add_admin, a.created_at, a.updated_at 
			` + baseQuery + whereClause + ` 
			ORDER BY a.created_at DESC 
			LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		selectQuery = `SELECT a.id, a.chat_id, a.phone, a.max_id, a.add_user, a.add_admin, a.created_at, a.updated_at 
			` + baseQuery + ` 
			ORDER BY a.created_at DESC 
			LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}
		administrators = append(administrators, admin)
	}

	return administrators, totalCount, rows.Err()
}

