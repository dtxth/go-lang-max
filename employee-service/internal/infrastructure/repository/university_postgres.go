package repository

import (
	"employee-service/internal/domain"
	"employee-service/internal/infrastructure/database"
	"strings"
)

type UniversityPostgres struct {
	db  *database.DB
	dsn string
}

func NewUniversityPostgres(db *database.DB) *UniversityPostgres {
	return &UniversityPostgres{db: db}
}

func NewUniversityPostgresWithDSN(db *database.DB, dsn string) *UniversityPostgres {
	return &UniversityPostgres{db: db, dsn: dsn}
}

// getDB returns a working database connection
func (r *UniversityPostgres) getDB() *database.DB {
	return r.db
}

func (r *UniversityPostgres) Create(university *domain.University) error {
	db := r.getDB()
	err := db.QueryRow(
		`INSERT INTO universities (name, inn, kpp) 
		 VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`,
		university.Name, university.INN, university.KPP,
	).Scan(&university.ID, &university.CreatedAt, &university.UpdatedAt)
	return err
}

func (r *UniversityPostgres) GetByID(id int64) (*domain.University, error) {
	university := &domain.University{}
	db := r.getDB()
	err := db.QueryRow(
		`SELECT id, name, inn, kpp, created_at, updated_at 
		 FROM universities WHERE id = $1`,
		id,
	).Scan(&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt)
	return university, err
}

func (r *UniversityPostgres) GetByINN(inn string) (*domain.University, error) {
	university := &domain.University{}
	db := r.getDB()
	err := db.QueryRow(
		`SELECT id, name, inn, kpp, created_at, updated_at 
		 FROM universities WHERE inn = $1 LIMIT 1`,
		inn,
	).Scan(&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt)
	return university, err
}

func (r *UniversityPostgres) GetByINNAndKPP(inn, kpp string) (*domain.University, error) {
	university := &domain.University{}
	db := r.getDB()
	err := db.QueryRow(
		`SELECT id, name, inn, kpp, created_at, updated_at 
		 FROM universities WHERE inn = $1 AND kpp = $2`,
		inn, kpp,
	).Scan(&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt)
	return university, err
}

func (r *UniversityPostgres) SearchByName(query string) ([]*domain.University, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return r.GetAll()
	}
	
	searchPattern := "%" + strings.ToLower(query) + "%"
	
	db := r.getDB()
	rows, err := db.Query(
		`SELECT id, name, inn, kpp, created_at, updated_at 
		 FROM universities 
		 WHERE LOWER(name) LIKE $1 
		 ORDER BY name`,
		searchPattern,
	)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var universities []*domain.University
	for rows.Next() {
		university := &domain.University{}
		err := rows.Scan(
			&university.ID, &university.Name, &university.INN, &university.KPP,
			&university.CreatedAt, &university.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		universities = append(universities, university)
	}
	
	return universities, rows.Err()
}

func (r *UniversityPostgres) GetAll() ([]*domain.University, error) {
	db := r.getDB()
	rows, err := db.Query(
		`SELECT id, name, inn, kpp, created_at, updated_at 
		 FROM universities 
		 ORDER BY name`,
	)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var universities []*domain.University
	for rows.Next() {
		university := &domain.University{}
		err := rows.Scan(
			&university.ID, &university.Name, &university.INN, &university.KPP,
			&university.CreatedAt, &university.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		universities = append(universities, university)
	}
	
	return universities, rows.Err()
}

