package repository

import (
	"chat-service/internal/domain"
	"database/sql"
	"strconv"
	"strings"
)

type ChatPostgres struct {
	db *sql.DB
}

func NewChatPostgres(db *sql.DB) *ChatPostgres {
	return &ChatPostgres{db: db}
}

func (r *ChatPostgres) Create(chat *domain.Chat) error {
	var universityID interface{}
	if chat.UniversityID != nil {
		universityID = *chat.UniversityID
	} else {
		universityID = nil
	}

	err := r.db.QueryRow(
		`INSERT INTO chats (name, url, max_chat_id, participants_count, university_id, department, source) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`,
		chat.Name, chat.URL, chat.MaxChatID, chat.ParticipantsCount, universityID, chat.Department, chat.Source,
	).Scan(&chat.ID, &chat.CreatedAt, &chat.UpdatedAt)
	return err
}

func (r *ChatPostgres) GetByID(id int64) (*domain.Chat, error) {
	chat := &domain.Chat{}
	var universityID sql.NullInt64
	var universityIDFromJoin sql.NullInt64
	var universityName sql.NullString
	var universityINN sql.NullString
	var universityKPP sql.NullString

	err := r.db.QueryRow(
		`SELECT c.id, c.name, c.url, c.max_chat_id, c.participants_count, 
		        c.university_id, c.department, c.source, c.created_at, c.updated_at,
		        u.id, u.name, u.inn, u.kpp
		 FROM chats c
		 LEFT JOIN universities u ON c.university_id = u.id
		 WHERE c.id = $1`,
		id,
	).Scan(
		&chat.ID, &chat.Name, &chat.URL, &chat.MaxChatID, &chat.ParticipantsCount,
		&universityID, &chat.Department, &chat.Source, &chat.CreatedAt, &chat.UpdatedAt,
		&universityIDFromJoin, &universityName, &universityINN, &universityKPP,
	)

	if err != nil {
		return nil, err
	}

	// Загружаем администраторов
	administrators, err := r.loadAdministrators(id)
	if err != nil {
		return nil, err
	}
	chat.Administrators = administrators

	// Устанавливаем вуз, если он есть
	if universityID.Valid {
		univID := universityID.Int64
		chat.UniversityID = &univID
		if universityIDFromJoin.Valid && universityName.Valid {
			chat.University = &domain.University{
				ID:   universityIDFromJoin.Int64,
				Name: universityName.String,
				INN:  universityINN.String,
				KPP:  universityKPP.String,
			}
		}
	}

	return chat, nil
}

func (r *ChatPostgres) GetByMaxChatID(maxChatID string) (*domain.Chat, error) {
	chat := &domain.Chat{}
	var universityID sql.NullInt64
	var universityIDFromJoin sql.NullInt64
	var universityName sql.NullString
	var universityINN sql.NullString
	var universityKPP sql.NullString

	err := r.db.QueryRow(
		`SELECT c.id, c.name, c.url, c.max_chat_id, c.participants_count, 
		        c.university_id, c.department, c.source, c.created_at, c.updated_at,
		        u.id, u.name, u.inn, u.kpp
		 FROM chats c
		 LEFT JOIN universities u ON c.university_id = u.id
		 WHERE c.max_chat_id = $1`,
		maxChatID,
	).Scan(
		&chat.ID, &chat.Name, &chat.URL, &chat.MaxChatID, &chat.ParticipantsCount,
		&universityID, &chat.Department, &chat.Source, &chat.CreatedAt, &chat.UpdatedAt,
		&universityIDFromJoin, &universityName, &universityINN, &universityKPP,
	)

	if err != nil {
		return nil, err
	}

	// Загружаем администраторов
	administrators, err := r.loadAdministrators(chat.ID)
	if err != nil {
		return nil, err
	}
	chat.Administrators = administrators

	// Устанавливаем вуз, если он есть
	if universityID.Valid {
		univID := universityID.Int64
		chat.UniversityID = &univID
		if universityIDFromJoin.Valid && universityName.Valid {
			chat.University = &domain.University{
				ID:   universityIDFromJoin.Int64,
				Name: universityName.String,
				INN:  universityINN.String,
				KPP:  universityKPP.String,
			}
		}
	}

	return chat, nil
}

func (r *ChatPostgres) Search(query string, limit, offset int, userRole string, universityID *int64) ([]*domain.Chat, int, error) {
	query = strings.TrimSpace(query)
	searchPattern := "%" + strings.ToLower(query) + "%"

	// Строим WHERE условие в зависимости от роли
	whereClause := "WHERE LOWER(c.name) LIKE $1"
	args := []interface{}{searchPattern}
	argIndex := 2

	// Фильтрация по роли и университету
	if userRole != "superadmin" && universityID != nil {
		whereClause += " AND c.university_id = $" + strconv.Itoa(argIndex)
		args = append(args, *universityID)
		argIndex++
	}

	// Подсчет общего количества
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM chats c ` + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Получение данных
	args = append(args, limit, offset)
	rows, err := r.db.Query(
		`SELECT c.id, c.name, c.url, c.max_chat_id, c.participants_count, 
		        c.university_id, c.department, c.source, c.created_at, c.updated_at,
		        u.id, u.name, u.inn, u.kpp
		 FROM chats c
		 LEFT JOIN universities u ON c.university_id = u.id
		 `+whereClause+`
		 ORDER BY c.name
		 LIMIT $`+strconv.Itoa(argIndex)+` OFFSET $`+strconv.Itoa(argIndex+1),
		args...,
	)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var chats []*domain.Chat
	chatIDs := make([]int64, 0)

	for rows.Next() {
		chat := &domain.Chat{}
		var universityID sql.NullInt64
		var universityIDFromJoin sql.NullInt64
		var universityName sql.NullString
		var universityINN sql.NullString
		var universityKPP sql.NullString

		err := rows.Scan(
			&chat.ID, &chat.Name, &chat.URL, &chat.MaxChatID, &chat.ParticipantsCount,
			&universityID, &chat.Department, &chat.Source, &chat.CreatedAt, &chat.UpdatedAt,
			&universityIDFromJoin, &universityName, &universityINN, &universityKPP,
		)
		if err != nil {
			return nil, 0, err
		}

		// Устанавливаем вуз, если он есть
		if universityID.Valid {
			univID := universityID.Int64
			chat.UniversityID = &univID
			if universityIDFromJoin.Valid && universityName.Valid {
				chat.University = &domain.University{
					ID:   universityIDFromJoin.Int64,
					Name: universityName.String,
					INN:  universityINN.String,
					KPP:  universityKPP.String,
				}
			}
		}

		chatIDs = append(chatIDs, chat.ID)
		chats = append(chats, chat)
	}

	// Загружаем администраторов для всех чатов одним запросом
	if len(chatIDs) > 0 {
		administratorsMap, err := r.loadAdministratorsBatch(chatIDs)
		if err == nil {
			for _, chat := range chats {
				chat.Administrators = administratorsMap[chat.ID]
			}
		}
	}

	return chats, totalCount, rows.Err()
}

func (r *ChatPostgres) GetAll(limit, offset int, userRole string, universityID *int64) ([]*domain.Chat, int, error) {
	return r.Search("", limit, offset, userRole, universityID)
}

func (r *ChatPostgres) Update(chat *domain.Chat) error {
	var universityID interface{}
	if chat.UniversityID != nil {
		universityID = *chat.UniversityID
	} else {
		universityID = nil
	}

	_, err := r.db.Exec(
		`UPDATE chats 
		 SET name = $1, url = $2, max_chat_id = $3, participants_count = $4, 
		     university_id = $5, department = $6, source = $7, updated_at = now()
		 WHERE id = $8`,
		chat.Name, chat.URL, chat.MaxChatID, chat.ParticipantsCount,
		universityID, chat.Department, chat.Source, chat.ID,
	)
	return err
}

func (r *ChatPostgres) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM chats WHERE id = $1`, id)
	return err
}

// loadAdministrators загружает администраторов для одного чата
func (r *ChatPostgres) loadAdministrators(chatID int64) ([]domain.Administrator, error) {
	rows, err := r.db.Query(
		`SELECT id, chat_id, phone, max_id, created_at, updated_at 
		 FROM administrators WHERE chat_id = $1 ORDER BY created_at`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var administrators []domain.Administrator
	for rows.Next() {
		var admin domain.Administrator
		err := rows.Scan(
			&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
			&admin.CreatedAt, &admin.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		administrators = append(administrators, admin)
	}
	return administrators, rows.Err()
}

// loadAdministratorsBatch загружает администраторов для нескольких чатов
func (r *ChatPostgres) loadAdministratorsBatch(chatIDs []int64) (map[int64][]domain.Administrator, error) {
	if len(chatIDs) == 0 {
		return make(map[int64][]domain.Administrator), nil
	}

	// Создаем плейсхолдеры для IN запроса
	placeholders := make([]string, len(chatIDs))
	args := make([]interface{}, len(chatIDs))
	for i, id := range chatIDs {
		placeholders[i] = "$" + string(rune('1'+i))
		args[i] = id
	}

	rows, err := r.db.Query(
		`SELECT id, chat_id, phone, max_id, created_at, updated_at 
		 FROM administrators 
		 WHERE chat_id IN (`+strings.Join(placeholders, ",")+`)
		 ORDER BY chat_id, created_at`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	administratorsMap := make(map[int64][]domain.Administrator)
	for rows.Next() {
		var admin domain.Administrator
		err := rows.Scan(
			&admin.ID, &admin.ChatID, &admin.Phone, &admin.MaxID,
			&admin.CreatedAt, &admin.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		administratorsMap[admin.ChatID] = append(administratorsMap[admin.ChatID], admin)
	}
	return administratorsMap, rows.Err()
}

