package repository

import (
	"chat-service/internal/domain"
	"database/sql"
	"fmt"
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

	var externalChatID interface{}
	if chat.ExternalChatID != nil {
		externalChatID = *chat.ExternalChatID
	} else {
		externalChatID = nil
	}

	err := r.db.QueryRow(
		`INSERT INTO chats (name, url, max_chat_id, external_chat_id, participants_count, university_id, department, source) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`,
		chat.Name, chat.URL, chat.MaxChatID, externalChatID, chat.ParticipantsCount, universityID, chat.Department, chat.Source,
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

	var externalChatID sql.NullString

	err := r.db.QueryRow(
		`SELECT c.id, c.name, c.url, c.max_chat_id, c.external_chat_id, c.participants_count, 
		        c.university_id, c.department, c.source, c.created_at, c.updated_at,
		        u.id, u.name, u.inn, u.kpp
		 FROM chats c
		 LEFT JOIN universities u ON c.university_id = u.id
		 WHERE c.id = $1`,
		id,
	).Scan(
		&chat.ID, &chat.Name, &chat.URL, &chat.MaxChatID, &externalChatID, &chat.ParticipantsCount,
		&universityID, &chat.Department, &chat.Source, &chat.CreatedAt, &chat.UpdatedAt,
		&universityIDFromJoin, &universityName, &universityINN, &universityKPP,
	)

	if err != nil {
		return nil, err
	}

	if externalChatID.Valid {
		chat.ExternalChatID = &externalChatID.String
	}

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
	var externalChatID sql.NullString

	err := r.db.QueryRow(
		`SELECT c.id, c.name, c.url, c.max_chat_id, c.external_chat_id, c.participants_count, 
		        c.university_id, c.department, c.source, c.created_at, c.updated_at,
		        u.id, u.name, u.inn, u.kpp
		 FROM chats c
		 LEFT JOIN universities u ON c.university_id = u.id
		 WHERE c.max_chat_id = $1`,
		maxChatID,
	).Scan(
		&chat.ID, &chat.Name, &chat.URL, &chat.MaxChatID, &externalChatID, &chat.ParticipantsCount,
		&universityID, &chat.Department, &chat.Source, &chat.CreatedAt, &chat.UpdatedAt,
		&universityIDFromJoin, &universityName, &universityINN, &universityKPP,
	)

	if err != nil {
		return nil, err
	}

	if externalChatID.Valid {
		chat.ExternalChatID = &externalChatID.String
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

func (r *ChatPostgres) Search(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	query = strings.TrimSpace(query)
	
	// Если запрос пустой, возвращаем все чаты с фильтрацией
	if query == "" {
		return r.GetAll(limit, offset, filter)
	}

	// Используем ILIKE для поиска, так как он более гибкий с специальными символами
	// Разбиваем запрос на слова, используя пробелы и специальные символы как разделители
	// Заменяем специальные символы на пробелы для правильного разбиения
	normalizedQuery := strings.Map(func(r rune) rune {
		if (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == 'ё' || r == 'Ё' {
			return r
		}
		return ' ' // Заменяем специальные символы на пробелы
	}, query)
	
	words := strings.Fields(normalizedQuery)
	if len(words) == 0 {
		return r.GetAll(limit, offset, filter)
	}

	// Строим WHERE условие с ILIKE для каждого слова
	// Ищем по нескольким полям: название чата, department, название университета
	whereClause := "WHERE "
	whereParts := make([]string, len(words))
	args := []interface{}{}
	argIndex := 1
	
	for i, word := range words {
		// Экранируем специальные символы LIKE
		escapedWord := strings.ReplaceAll(word, "%", "\\%")
		escapedWord = strings.ReplaceAll(escapedWord, "_", "\\_")
		
		// Для каждого слова ищем в любом из полей (OR внутри, AND между словами)
		whereParts[i] = fmt.Sprintf(
			"(c.name ILIKE $%d OR c.department ILIKE $%d OR u.name ILIKE $%d)",
			argIndex, argIndex, argIndex,
		)
		args = append(args, "%"+escapedWord+"%")
		argIndex++
	}
	
	whereClause += strings.Join(whereParts, " AND ")

	// Фильтрация по роли и контексту
	if filter != nil {
		if filter.IsSuperadmin() {
			// Суперадмин видит все чаты - не добавляем фильтры
		} else if filter.IsCurator() && filter.UniversityID != nil {
			// Куратор видит только чаты своего вуза
			whereClause += " AND c.university_id = $" + strconv.Itoa(argIndex)
			args = append(args, *filter.UniversityID)
			argIndex++
		} else if filter.IsOperator() && filter.UniversityID != nil {
			// Оператор видит только чаты своего вуза
			// TODO: В будущем добавить фильтрацию по branch_id и faculty_id
			whereClause += " AND c.university_id = $" + strconv.Itoa(argIndex)
			args = append(args, *filter.UniversityID)
			argIndex++
		}
	}

	// Подсчет общего количества
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM chats c LEFT JOIN universities u ON c.university_id = u.id ` + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Получение данных с сортировкой по имени
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

func (r *ChatPostgres) GetAll(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	// Строим WHERE условие в зависимости от роли
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	// Фильтрация по роли и контексту
	if filter != nil {
		if filter.IsSuperadmin() {
			// Суперадмин видит все чаты - не добавляем фильтры
		} else if filter.IsCurator() && filter.UniversityID != nil {
			// Куратор видит только чаты своего вуза
			whereClause = "WHERE c.university_id = $" + strconv.Itoa(argIndex)
			args = append(args, *filter.UniversityID)
			argIndex++
		} else if filter.IsOperator() && filter.UniversityID != nil {
			// Оператор видит только чаты своего вуза
			// TODO: В будущем добавить фильтрацию по branch_id и faculty_id
			whereClause = "WHERE c.university_id = $" + strconv.Itoa(argIndex)
			args = append(args, *filter.UniversityID)
			argIndex++
		}
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
		`SELECT id, chat_id, phone, COALESCE(max_id, ''), add_user, add_admin, created_at, updated_at 
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

// loadAdministratorsBatch загружает администраторов для нескольких чатов
func (r *ChatPostgres) loadAdministratorsBatch(chatIDs []int64) (map[int64][]domain.Administrator, error) {
	if len(chatIDs) == 0 {
		return make(map[int64][]domain.Administrator), nil
	}

	// Создаем плейсхолдеры для IN запроса
	placeholders := make([]string, len(chatIDs))
	args := make([]interface{}, len(chatIDs))
	for i, id := range chatIDs {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	rows, err := r.db.Query(
		`SELECT id, chat_id, phone, COALESCE(max_id, ''), add_user, add_admin, created_at, updated_at 
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
			&admin.AddUser, &admin.AddAdmin,
			&admin.CreatedAt, &admin.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		administratorsMap[admin.ChatID] = append(administratorsMap[admin.ChatID], admin)
	}
	return administratorsMap, rows.Err()
}

func (r *ChatPostgres) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	// Валидация параметров сортировки
	validSortFields := map[string]string{
		"id":                 "c.id",
		"name":               "c.name",
		"url":                "c.url",
		"max_chat_id":        "c.max_chat_id",
		"participants_count": "c.participants_count",
		"department":         "c.department",
		"source":             "c.source",
		"university":         "u.name",
		"created_at":         "c.created_at",
		"updated_at":         "c.updated_at",
	}
	
	sortField, exists := validSortFields[sortBy]
	if !exists {
		sortField = "c.name" // по умолчанию сортировка по названию
	}
	
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc" // по умолчанию по возрастанию
	}
	
	// Построение WHERE условия для поиска
	whereClause := ""
	args := []interface{}{}
	argIndex := 1
	
	if search != "" {
		// Нормализуем поисковый запрос
		normalizedQuery := strings.Map(func(r rune) rune {
			if (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') ||
				(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') || r == 'ё' || r == 'Ё' {
				return r
			}
			return ' ' // Заменяем специальные символы на пробелы
		}, search)
		
		words := strings.Fields(normalizedQuery)
		if len(words) > 0 {
			whereParts := make([]string, len(words))
			for i, word := range words {
				// Экранируем специальные символы LIKE
				escapedWord := strings.ReplaceAll(word, "%", "\\%")
				escapedWord = strings.ReplaceAll(escapedWord, "_", "\\_")
				
				// Для каждого слова ищем в любом из полей
				whereParts[i] = fmt.Sprintf(
					"(c.name ILIKE $%d OR c.department ILIKE $%d OR u.name ILIKE $%d OR c.max_chat_id ILIKE $%d OR c.source ILIKE $%d)",
					argIndex, argIndex, argIndex, argIndex, argIndex,
				)
				args = append(args, "%"+escapedWord+"%")
				argIndex++
			}
			whereClause = "WHERE " + strings.Join(whereParts, " AND ")
		}
	}
	
	// Фильтрация по роли и контексту
	if filter != nil {
		roleFilter := ""
		if filter.IsSuperadmin() {
			// Суперадмин видит все чаты - не добавляем фильтры
		} else if filter.IsCurator() && filter.UniversityID != nil {
			// Куратор видит только чаты своего вуза
			roleFilter = "c.university_id = $" + strconv.Itoa(argIndex)
			args = append(args, *filter.UniversityID)
			argIndex++
		} else if filter.IsOperator() && filter.UniversityID != nil {
			// Оператор видит только чаты своего вуза
			roleFilter = "c.university_id = $" + strconv.Itoa(argIndex)
			args = append(args, *filter.UniversityID)
			argIndex++
		}
		
		if roleFilter != "" {
			if whereClause != "" {
				whereClause += " AND " + roleFilter
			} else {
				whereClause = "WHERE " + roleFilter
			}
		}
	}
	
	// Подсчет общего количества
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM chats c LEFT JOIN universities u ON c.university_id = u.id ` + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Добавляем LIMIT и OFFSET
	args = append(args, limit, offset)
	limitArg := "$" + strconv.Itoa(argIndex)
	offsetArg := "$" + strconv.Itoa(argIndex+1)
	
	query := `SELECT c.id, c.name, c.url, c.max_chat_id, c.external_chat_id, c.participants_count, 
		        c.university_id, c.department, c.source, c.created_at, c.updated_at,
		        u.id, u.name, u.inn, u.kpp, u.created_at, u.updated_at
		 FROM chats c
		 LEFT JOIN universities u ON c.university_id = u.id ` +
		whereClause + `
		 ORDER BY ` + sortField + ` ` + sortOrder + `
		 LIMIT ` + limitArg + ` OFFSET ` + offsetArg
	
	rows, err := r.db.Query(query, args...)
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
		var universityCreatedAt sql.NullTime
		var universityUpdatedAt sql.NullTime
		var externalChatID sql.NullString
		
		err := rows.Scan(
			&chat.ID, &chat.Name, &chat.URL, &chat.MaxChatID, &externalChatID, &chat.ParticipantsCount,
			&universityID, &chat.Department, &chat.Source, &chat.CreatedAt, &chat.UpdatedAt,
			&universityIDFromJoin, &universityName, &universityINN, &universityKPP,
			&universityCreatedAt, &universityUpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		if externalChatID.Valid {
			chat.ExternalChatID = &externalChatID.String
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
				if universityCreatedAt.Valid {
					chat.University.CreatedAt = universityCreatedAt.Time
				}
				if universityUpdatedAt.Valid {
					chat.University.UpdatedAt = universityUpdatedAt.Time
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
