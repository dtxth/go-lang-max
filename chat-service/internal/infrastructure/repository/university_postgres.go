package repository

import (
	"chat-service/internal/domain"
	"database/sql"
)

type UniversityPostgres struct {
	db *sql.DB
}

func NewUniversityPostgres(db *sql.DB) *UniversityPostgres {
	return &UniversityPostgres{db: db}
}

func (r *UniversityPostgres) GetByID(id int64) (*domain.University, error) {
	university := &domain.University{}
	err := r.db.QueryRow(
		`SELECT id, name, inn, kpp, created_at, updated_at 
		 FROM universities WHERE id = $1`,
		id,
	).Scan(&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt)
	return university, err
}

func (r *UniversityPostgres) GetByINN(inn string) (*domain.University, error) {
	university := &domain.University{}
	err := r.db.QueryRow(
		`SELECT id, name, inn, kpp, created_at, updated_at 
		 FROM universities WHERE inn = $1 LIMIT 1`,
		inn,
	).Scan(&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUniversityNotFound
	}
	return university, err
}

func (r *UniversityPostgres) GetByINNAndKPP(inn, kpp string) (*domain.University, error) {
	university := &domain.University{}
	err := r.db.QueryRow(
		`SELECT id, name, inn, kpp, created_at, updated_at 
		 FROM universities WHERE inn = $1 AND kpp = $2`,
		inn, kpp,
	).Scan(&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUniversityNotFound
	}
	return university, err
}

func (r *UniversityPostgres) Create(university *domain.University) error {
	err := r.db.QueryRow(
		`INSERT INTO universities (name, inn, kpp) 
		 VALUES ($1, $2, $3) 
		 ON CONFLICT (inn, kpp) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id, created_at, updated_at`,
		university.Name, university.INN, university.KPP,
	).Scan(&university.ID, &university.CreatedAt, &university.UpdatedAt)
	return err
}

