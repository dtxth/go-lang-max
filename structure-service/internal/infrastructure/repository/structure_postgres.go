package repository

import (
	"database/sql"
	"fmt"
	"structure-service/internal/domain"
	"time"
)

type StructurePostgres struct {
	db  *sql.DB
	dsn string
}

func NewStructurePostgres(db *sql.DB) domain.StructureRepository {
	return &StructurePostgres{db: db}
}

func NewStructurePostgresWithDSN(db *sql.DB, dsn string) domain.StructureRepository {
	return &StructurePostgres{db: db, dsn: dsn}
}

// getDB returns a working database connection, reconnecting if necessary
func (r *StructurePostgres) getDB() (*sql.DB, error) {
	// Try to ping the existing connection
	if r.db != nil {
		if err := r.db.Ping(); err == nil {
			return r.db, nil
		}
	}

	// If we have a DSN, try to reconnect
	if r.dsn != "" {
		db, err := sql.Open("postgres", r.dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to reconnect to database: %w", err)
		}
		
		// Configure connection pool
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(0)
		
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to ping reconnected database: %w", err)
		}
		
		r.db = db
		return db, nil
	}

	return r.db, nil
}

// University methods
func (r *StructurePostgres) CreateUniversity(u *domain.University) error {
	db, err := r.getDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `INSERT INTO universities (name, inn, kpp, foiv, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = db.QueryRow(query, u.Name, u.INN, u.KPP, u.FOIV, time.Now(), time.Now()).Scan(&u.ID)
	if err != nil {
		return fmt.Errorf("failed to create university: %w", err)
	}
	// Устанавливаем chats_count в 0 для новых университетов
	u.ChatsCount = 0
	return nil
}

func (r *StructurePostgres) GetUniversityByID(id int64) (*domain.University, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	u := &domain.University{}
	query := `SELECT u.id, u.name, u.inn, u.kpp, u.foiv, u.created_at, u.updated_at,
		         (SELECT COUNT(DISTINCT g.chat_id) 
		          FROM branches b
		          LEFT JOIN faculties f ON f.branch_id = b.id
		          LEFT JOIN groups g ON g.faculty_id = f.id AND g.chat_id IS NOT NULL
		          WHERE b.university_id = u.id) as chats_count
		 FROM universities u
		 WHERE u.id = $1`
	err = db.QueryRow(query, id).Scan(&u.ID, &u.Name, &u.INN, &u.KPP, &u.FOIV, &u.CreatedAt, &u.UpdatedAt, &u.ChatsCount)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUniversityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get university by ID: %w", err)
	}
	return u, nil
}

func (r *StructurePostgres) GetUniversityByINN(inn string) (*domain.University, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	u := &domain.University{}
	query := `SELECT u.id, u.name, u.inn, u.kpp, u.foiv, u.created_at, u.updated_at,
		         (SELECT COUNT(DISTINCT g.chat_id) 
		          FROM branches b
		          LEFT JOIN faculties f ON f.branch_id = b.id
		          LEFT JOIN groups g ON g.faculty_id = f.id AND g.chat_id IS NOT NULL
		          WHERE b.university_id = u.id) as chats_count
		 FROM universities u
		 WHERE u.inn = $1 LIMIT 1`
	err = db.QueryRow(query, inn).Scan(&u.ID, &u.Name, &u.INN, &u.KPP, &u.FOIV, &u.CreatedAt, &u.UpdatedAt, &u.ChatsCount)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUniversityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get university by INN: %w", err)
	}
	return u, nil
}

func (r *StructurePostgres) GetUniversityByINNAndKPP(inn, kpp string) (*domain.University, error) {
	u := &domain.University{}
	query := `SELECT u.id, u.name, u.inn, u.kpp, u.foiv, u.created_at, u.updated_at,
		         (SELECT COUNT(DISTINCT g.chat_id) 
		          FROM branches b
		          LEFT JOIN faculties f ON f.branch_id = b.id
		          LEFT JOIN groups g ON g.faculty_id = f.id AND g.chat_id IS NOT NULL
		          WHERE b.university_id = u.id) as chats_count
		 FROM universities u
		 WHERE u.inn = $1 AND u.kpp = $2`
	err := r.db.QueryRow(query, inn, kpp).Scan(&u.ID, &u.Name, &u.INN, &u.KPP, &u.FOIV, &u.CreatedAt, &u.UpdatedAt, &u.ChatsCount)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUniversityNotFound
	}
	return u, err
}

func (r *StructurePostgres) UpdateUniversity(u *domain.University) error {
	query := `UPDATE universities SET name = $1, inn = $2, kpp = $3, foiv = $4, updated_at = $5 
			  WHERE id = $6`
	_, err := r.db.Exec(query, u.Name, u.INN, u.KPP, u.FOIV, time.Now(), u.ID)
	return err
}

func (r *StructurePostgres) DeleteUniversity(id int64) error {
	_, err := r.db.Exec("DELETE FROM universities WHERE id = $1", id)
	return err
}

func (r *StructurePostgres) GetAllUniversities() ([]*domain.University, error) {
	query := `SELECT u.id, u.name, u.inn, u.kpp, u.foiv, u.created_at, u.updated_at,
		         (SELECT COUNT(DISTINCT g.chat_id) 
		          FROM branches b
		          LEFT JOIN faculties f ON f.branch_id = b.id
		          LEFT JOIN groups g ON g.faculty_id = f.id AND g.chat_id IS NOT NULL
		          WHERE b.university_id = u.id) as chats_count
		 FROM universities u
		 ORDER BY u.name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var universities []*domain.University
	for rows.Next() {
		u := &domain.University{}
		if err := rows.Scan(&u.ID, &u.Name, &u.INN, &u.KPP, &u.FOIV, &u.CreatedAt, &u.UpdatedAt, &u.ChatsCount); err != nil {
			return nil, err
		}
		universities = append(universities, u)
	}
	return universities, rows.Err()
}

func (r *StructurePostgres) GetAllUniversitiesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.University, int, error) {
	// Валидация параметров сортировки
	validSortFields := map[string]string{
		"id":         "id",
		"name":       "name",
		"inn":        "inn",
		"kpp":        "kpp",
		"foiv":       "foiv",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
	
	sortField, exists := validSortFields[sortBy]
	if !exists {
		sortField = "name" // по умолчанию сортировка по названию
	}
	
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc" // по умолчанию по возрастанию
	}
	
	// Построение WHERE условия для поиска
	whereClause := ""
	args := []interface{}{}
	argIndex := 1
	
	if search != "" {
		searchPattern := "%" + search + "%"
		whereClause = `WHERE (LOWER(u.name) LIKE LOWER($` + fmt.Sprintf("%d", argIndex) + `) 
		                  OR LOWER(u.inn) LIKE LOWER($` + fmt.Sprintf("%d", argIndex) + `) 
		                  OR LOWER(u.kpp) LIKE LOWER($` + fmt.Sprintf("%d", argIndex) + `)
		                  OR LOWER(u.foiv) LIKE LOWER($` + fmt.Sprintf("%d", argIndex) + `))`
		args = append(args, searchPattern)
		argIndex++
	}
	
	// Подсчет общего количества
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM universities u ` + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Добавляем LIMIT и OFFSET
	args = append(args, limit, offset)
	limitArg := fmt.Sprintf("$%d", argIndex)
	offsetArg := fmt.Sprintf("$%d", argIndex+1)
	
	// Запрос с подсчетом чатов
	query := `SELECT u.id, u.name, u.inn, u.kpp, u.foiv, u.created_at, u.updated_at,
		         (SELECT COUNT(DISTINCT g.chat_id) 
		          FROM branches b
		          LEFT JOIN faculties f ON f.branch_id = b.id
		          LEFT JOIN groups g ON g.faculty_id = f.id AND g.chat_id IS NOT NULL
		          WHERE b.university_id = u.id) as chats_count
		 FROM universities u ` +
		whereClause + `
		 ORDER BY u.` + sortField + ` ` + sortOrder + `
		 LIMIT ` + limitArg + ` OFFSET ` + offsetArg
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var universities []*domain.University
	for rows.Next() {
		u := &domain.University{}
		err := rows.Scan(&u.ID, &u.Name, &u.INN, &u.KPP, &u.FOIV, &u.CreatedAt, &u.UpdatedAt, &u.ChatsCount)
		if err != nil {
			return nil, 0, err
		}
		universities = append(universities, u)
	}
	
	return universities, totalCount, rows.Err()
}

// Branch methods
func (r *StructurePostgres) CreateBranch(b *domain.Branch) error {
	db, err := r.getDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `INSERT INTO branches (university_id, name, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`
	err = db.QueryRow(query, b.UniversityID, b.Name, time.Now(), time.Now()).Scan(&b.ID)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

func (r *StructurePostgres) GetBranchByID(id int64) (*domain.Branch, error) {
	b := &domain.Branch{}
	query := `SELECT id, university_id, name, created_at, updated_at FROM branches WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&b.ID, &b.UniversityID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrBranchNotFound
	}
	return b, err
}

func (r *StructurePostgres) GetBranchByUniversityAndName(universityID int64, name string) (*domain.Branch, error) {
	b := &domain.Branch{}
	query := `SELECT id, university_id, name, created_at, updated_at FROM branches 
			  WHERE university_id = $1 AND name = $2`
	err := r.db.QueryRow(query, universityID, name).Scan(&b.ID, &b.UniversityID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrBranchNotFound
	}
	return b, err
}

func (r *StructurePostgres) GetBranchesByUniversityID(universityID int64) ([]*domain.Branch, error) {
	query := `SELECT id, university_id, name, created_at, updated_at 
			  FROM branches WHERE university_id = $1 ORDER BY name`
	rows, err := r.db.Query(query, universityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*domain.Branch
	for rows.Next() {
		b := &domain.Branch{}
		if err := rows.Scan(&b.ID, &b.UniversityID, &b.Name, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, rows.Err()
}

func (r *StructurePostgres) UpdateBranch(b *domain.Branch) error {
	query := `UPDATE branches SET name = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, b.Name, time.Now(), b.ID)
	return err
}

func (r *StructurePostgres) DeleteBranch(id int64) error {
	_, err := r.db.Exec("DELETE FROM branches WHERE id = $1", id)
	return err
}

// Faculty methods
func (r *StructurePostgres) CreateFaculty(f *domain.Faculty) error {
	db, err := r.getDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `INSERT INTO faculties (branch_id, name, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`
	err = db.QueryRow(query, f.BranchID, f.Name, time.Now(), time.Now()).Scan(&f.ID)
	if err != nil {
		return fmt.Errorf("failed to create faculty: %w", err)
	}
	return nil
}

func (r *StructurePostgres) GetFacultyByID(id int64) (*domain.Faculty, error) {
	f := &domain.Faculty{}
	query := `SELECT id, branch_id, name, created_at, updated_at FROM faculties WHERE id = $1`
	var branchID sql.NullInt64
	err := r.db.QueryRow(query, id).Scan(&f.ID, &branchID, &f.Name, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrFacultyNotFound
	}
	if branchID.Valid {
		f.BranchID = &branchID.Int64
	}
	return f, err
}

func (r *StructurePostgres) GetFacultyByBranchAndName(branchID *int64, name string) (*domain.Faculty, error) {
	f := &domain.Faculty{}
	var query string
	var err error
	var dbBranchID sql.NullInt64
	
	if branchID == nil {
		query = `SELECT id, branch_id, name, created_at, updated_at FROM faculties 
				 WHERE branch_id IS NULL AND name = $1`
		err = r.db.QueryRow(query, name).Scan(&f.ID, &dbBranchID, &f.Name, &f.CreatedAt, &f.UpdatedAt)
	} else {
		query = `SELECT id, branch_id, name, created_at, updated_at FROM faculties 
				 WHERE branch_id = $1 AND name = $2`
		err = r.db.QueryRow(query, *branchID, name).Scan(&f.ID, &dbBranchID, &f.Name, &f.CreatedAt, &f.UpdatedAt)
	}
	
	if err == sql.ErrNoRows {
		return nil, domain.ErrFacultyNotFound
	}
	if dbBranchID.Valid {
		f.BranchID = &dbBranchID.Int64
	}
	return f, err
}

func (r *StructurePostgres) GetFacultiesByBranchID(branchID int64) ([]*domain.Faculty, error) {
	query := `SELECT id, branch_id, name, created_at, updated_at 
			  FROM faculties WHERE branch_id = $1 ORDER BY name`
	rows, err := r.db.Query(query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var faculties []*domain.Faculty
	for rows.Next() {
		f := &domain.Faculty{}
		var branchID sql.NullInt64
		if err := rows.Scan(&f.ID, &branchID, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		if branchID.Valid {
			f.BranchID = &branchID.Int64
		}
		faculties = append(faculties, f)
	}
	return faculties, rows.Err()
}

func (r *StructurePostgres) GetFacultiesByUniversityID(universityID int64) ([]*domain.Faculty, error) {
	query := `SELECT f.id, f.branch_id, f.name, f.created_at, f.updated_at 
			  FROM faculties f
			  LEFT JOIN branches b ON f.branch_id = b.id
			  WHERE f.branch_id IS NULL OR b.university_id = $1
			  ORDER BY f.name`
	rows, err := r.db.Query(query, universityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var faculties []*domain.Faculty
	for rows.Next() {
		f := &domain.Faculty{}
		var branchID sql.NullInt64
		if err := rows.Scan(&f.ID, &branchID, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		if branchID.Valid {
			f.BranchID = &branchID.Int64
		}
		faculties = append(faculties, f)
	}
	return faculties, rows.Err()
}

func (r *StructurePostgres) UpdateFaculty(f *domain.Faculty) error {
	query := `UPDATE faculties SET branch_id = $1, name = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(query, f.BranchID, f.Name, time.Now(), f.ID)
	return err
}

func (r *StructurePostgres) DeleteFaculty(id int64) error {
	_, err := r.db.Exec("DELETE FROM faculties WHERE id = $1", id)
	return err
}

// Group methods
func (r *StructurePostgres) CreateGroup(g *domain.Group) error {
	db, err := r.getDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `INSERT INTO groups (faculty_id, course, number, chat_id, chat_url, chat_name, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err = db.QueryRow(query, g.FacultyID, g.Course, g.Number, g.ChatID, g.ChatURL, g.ChatName, time.Now(), time.Now()).Scan(&g.ID)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	return nil
}

func (r *StructurePostgres) GetGroupByID(id int64) (*domain.Group, error) {
	g := &domain.Group{}
	query := `SELECT id, faculty_id, course, number, chat_id, chat_url, chat_name, created_at, updated_at FROM groups WHERE id = $1`
	var chatID sql.NullInt64
	var chatURL, chatName sql.NullString
	err := r.db.QueryRow(query, id).Scan(&g.ID, &g.FacultyID, &g.Course, &g.Number, &chatID, &chatURL, &chatName, &g.CreatedAt, &g.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrGroupNotFound
	}
	if chatID.Valid {
		g.ChatID = &chatID.Int64
	}
	if chatURL.Valid {
		g.ChatURL = chatURL.String
	}
	if chatName.Valid {
		g.ChatName = chatName.String
	}
	return g, err
}

func (r *StructurePostgres) GetGroupByFacultyAndNumber(facultyID int64, course int, number string) (*domain.Group, error) {
	g := &domain.Group{}
	query := `SELECT id, faculty_id, course, number, chat_id, chat_url, chat_name, created_at, updated_at 
			  FROM groups WHERE faculty_id = $1 AND course = $2 AND number = $3`
	var chatID sql.NullInt64
	var chatURL, chatName sql.NullString
	err := r.db.QueryRow(query, facultyID, course, number).Scan(&g.ID, &g.FacultyID, &g.Course, &g.Number, &chatID, &chatURL, &chatName, &g.CreatedAt, &g.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrGroupNotFound
	}
	if chatID.Valid {
		g.ChatID = &chatID.Int64
	}
	if chatURL.Valid {
		g.ChatURL = chatURL.String
	}
	if chatName.Valid {
		g.ChatName = chatName.String
	}
	return g, err
}

func (r *StructurePostgres) GetGroupsByFacultyID(facultyID int64) ([]*domain.Group, error) {
	query := `SELECT id, faculty_id, course, number, chat_id, chat_url, chat_name, created_at, updated_at 
			  FROM groups WHERE faculty_id = $1 ORDER BY course, number`
	rows, err := r.db.Query(query, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		g := &domain.Group{}
		var chatID sql.NullInt64
		var chatURL, chatName sql.NullString
		if err := rows.Scan(&g.ID, &g.FacultyID, &g.Course, &g.Number, &chatID, &chatURL, &chatName, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		if chatID.Valid {
			g.ChatID = &chatID.Int64
		}
		if chatURL.Valid {
			g.ChatURL = chatURL.String
		}
		if chatName.Valid {
			g.ChatName = chatName.String
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *StructurePostgres) UpdateGroup(g *domain.Group) error {
	query := `UPDATE groups SET faculty_id = $1, course = $2, number = $3, chat_id = $4, chat_url = $5, chat_name = $6, updated_at = $7 
			  WHERE id = $8`
	_, err := r.db.Exec(query, g.FacultyID, g.Course, g.Number, g.ChatID, g.ChatURL, g.ChatName, time.Now(), g.ID)
	return err
}

func (r *StructurePostgres) DeleteGroup(id int64) error {
	_, err := r.db.Exec("DELETE FROM groups WHERE id = $1", id)
	return err
}

// GetStructureByUniversityID получает полную иерархическую структуру
func (r *StructurePostgres) GetStructureByUniversityID(universityID int64) (*domain.StructureNode, error) {
	university, err := r.GetUniversityByID(universityID)
	if err != nil {
		return nil, err
	}

	node := &domain.StructureNode{
		Type:     "university",
		ID:       university.ID,
		Name:     university.Name,
		Children: []*domain.StructureNode{},
	}

	// Получаем филиалы
	branches, err := r.GetBranchesByUniversityID(universityID)
	if err != nil {
		return nil, err
	}

	// Если есть филиалы, создаем узлы для них
	if len(branches) > 0 {
		for _, branch := range branches {
			branchNode := &domain.StructureNode{
				Type:     "branch",
				ID:       branch.ID,
				Name:     branch.Name,
				Children: []*domain.StructureNode{},
			}

			// Получаем факультеты для филиала
			faculties, err := r.GetFacultiesByBranchID(branch.ID)
			if err != nil {
				return nil, err
			}

			for _, faculty := range faculties {
				facultyNode := &domain.StructureNode{
					Type:     "faculty",
					ID:       faculty.ID,
					Name:     faculty.Name,
					Children: []*domain.StructureNode{},
				}

				// Получаем группы для факультета
				groups, err := r.GetGroupsByFacultyID(faculty.ID)
				if err != nil {
					return nil, err
				}

				for _, group := range groups {
					groupNode := &domain.StructureNode{
						Type:     "group",
						ID:       group.ID,
						Name:     group.Number,
						Course:   &group.Course,
						GroupNum: &group.Number,
					}

					if group.ChatID != nil {
						// Здесь можно получить информацию о чате из chat-service
						// Пока оставляем только ID
						groupNode.Chat = &domain.Chat{ID: *group.ChatID}
					}

					facultyNode.Children = append(facultyNode.Children, groupNode)
				}

				branchNode.Children = append(branchNode.Children, facultyNode)
			}

			node.Children = append(node.Children, branchNode)
		}
	} else {
		// Если нет филиалов, получаем факультеты напрямую для вуза
		faculties, err := r.GetFacultiesByUniversityID(universityID)
		if err != nil {
			return nil, err
		}

		for _, faculty := range faculties {
			facultyNode := &domain.StructureNode{
				Type:     "faculty",
				ID:       faculty.ID,
				Name:     faculty.Name,
				Children: []*domain.StructureNode{},
			}

			// Получаем группы для факультета
			groups, err := r.GetGroupsByFacultyID(faculty.ID)
			if err != nil {
				return nil, err
			}

			for _, group := range groups {
				groupNode := &domain.StructureNode{
					Type:     "group",
					ID:       group.ID,
					Name:     group.Number,
					Course:   &group.Course,
					GroupNum: &group.Number,
				}

				if group.ChatID != nil {
					groupNode.Chat = &domain.Chat{ID: *group.ChatID}
				}

				facultyNode.Children = append(facultyNode.Children, groupNode)
			}

			node.Children = append(node.Children, facultyNode)
		}
	}

	return node, nil
}

// Chat counting methods

// GetChatCountForUniversity counts all chats in a university
func (r *StructurePostgres) GetChatCountForUniversity(universityID int64) (int, error) {
	query := `SELECT COUNT(DISTINCT g.chat_id) 
			  FROM groups g
			  JOIN faculties f ON g.faculty_id = f.id
			  LEFT JOIN branches b ON f.branch_id = b.id
			  WHERE g.chat_id IS NOT NULL 
			  AND (b.university_id = $1 OR (f.branch_id IS NULL AND EXISTS(
				  SELECT 1 FROM universities u WHERE u.id = $1
			  )))`
	
	var count int
	err := r.db.QueryRow(query, universityID).Scan(&count)
	return count, err
}

// GetChatCountForBranch counts all chats in a branch
func (r *StructurePostgres) GetChatCountForBranch(branchID int64) (int, error) {
	query := `SELECT COUNT(DISTINCT g.chat_id) 
			  FROM groups g
			  JOIN faculties f ON g.faculty_id = f.id
			  WHERE f.branch_id = $1 AND g.chat_id IS NOT NULL`
	
	var count int
	err := r.db.QueryRow(query, branchID).Scan(&count)
	return count, err
}

// GetChatCountForFaculty counts all chats in a faculty
func (r *StructurePostgres) GetChatCountForFaculty(facultyID int64) (int, error) {
	query := `SELECT COUNT(DISTINCT g.chat_id) 
			  FROM groups g
			  WHERE g.faculty_id = $1 AND g.chat_id IS NOT NULL`
	
	var count int
	err := r.db.QueryRow(query, facultyID).Scan(&count)
	return count, err
}