package repository

import (
	"employee-service/internal/domain"
	"database/sql"
	"strings"
)

type EmployeePostgres struct {
	db *sql.DB
}

func NewEmployeePostgres(db *sql.DB) *EmployeePostgres {
	return &EmployeePostgres{db: db}
}

func (r *EmployeePostgres) Create(employee *domain.Employee) error {
	err := r.db.QueryRow(
		`INSERT INTO employees (first_name, last_name, middle_name, phone, max_id, inn, kpp, university_id) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`,
		employee.FirstName, employee.LastName, employee.MiddleName, employee.Phone,
		employee.MaxID, employee.INN, employee.KPP, employee.UniversityID,
	).Scan(&employee.ID, &employee.CreatedAt, &employee.UpdatedAt)
	return err
}

func (r *EmployeePostgres) GetByID(id int64) (*domain.Employee, error) {
	employee := &domain.Employee{}
	university := &domain.University{}
	
	err := r.db.QueryRow(
		`SELECT e.id, e.first_name, e.last_name, e.middle_name, e.phone, e.max_id, e.inn, e.kpp, 
		        e.university_id, e.created_at, e.updated_at,
		        u.id, u.name, u.inn, u.kpp, u.created_at, u.updated_at
		 FROM employees e
		 JOIN universities u ON e.university_id = u.id
		 WHERE e.id = $1`,
		id,
	).Scan(
		&employee.ID, &employee.FirstName, &employee.LastName, &employee.MiddleName,
		&employee.Phone, &employee.MaxID, &employee.INN, &employee.KPP,
		&employee.UniversityID, &employee.CreatedAt, &employee.UpdatedAt,
		&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	employee.University = university
	return employee, nil
}

func (r *EmployeePostgres) GetByPhone(phone string) (*domain.Employee, error) {
	employee := &domain.Employee{}
	university := &domain.University{}
	
	err := r.db.QueryRow(
		`SELECT e.id, e.first_name, e.last_name, e.middle_name, e.phone, e.max_id, e.inn, e.kpp, 
		        e.university_id, e.created_at, e.updated_at,
		        u.id, u.name, u.inn, u.kpp, u.created_at, u.updated_at
		 FROM employees e
		 JOIN universities u ON e.university_id = u.id
		 WHERE e.phone = $1`,
		phone,
	).Scan(
		&employee.ID, &employee.FirstName, &employee.LastName, &employee.MiddleName,
		&employee.Phone, &employee.MaxID, &employee.INN, &employee.KPP,
		&employee.UniversityID, &employee.CreatedAt, &employee.UpdatedAt,
		&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	employee.University = university
	return employee, nil
}

func (r *EmployeePostgres) GetByMaxID(maxID string) (*domain.Employee, error) {
	employee := &domain.Employee{}
	university := &domain.University{}
	
	err := r.db.QueryRow(
		`SELECT e.id, e.first_name, e.last_name, e.middle_name, e.phone, e.max_id, e.inn, e.kpp, 
		        e.university_id, e.created_at, e.updated_at,
		        u.id, u.name, u.inn, u.kpp, u.created_at, u.updated_at
		 FROM employees e
		 JOIN universities u ON e.university_id = u.id
		 WHERE e.max_id = $1`,
		maxID,
	).Scan(
		&employee.ID, &employee.FirstName, &employee.LastName, &employee.MiddleName,
		&employee.Phone, &employee.MaxID, &employee.INN, &employee.KPP,
		&employee.UniversityID, &employee.CreatedAt, &employee.UpdatedAt,
		&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	employee.University = university
	return employee, nil
}

func (r *EmployeePostgres) Search(query string, limit, offset int) ([]*domain.Employee, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return r.GetAll(limit, offset)
	}
	
	searchPattern := "%" + strings.ToLower(query) + "%"
	
	rows, err := r.db.Query(
		`SELECT e.id, e.first_name, e.last_name, e.middle_name, e.phone, e.max_id, e.inn, e.kpp, 
		        e.university_id, e.created_at, e.updated_at,
		        u.id, u.name, u.inn, u.kpp, u.created_at, u.updated_at
		 FROM employees e
		 JOIN universities u ON e.university_id = u.id
		 WHERE LOWER(e.first_name) LIKE $1 
		    OR LOWER(e.last_name) LIKE $1 
		    OR LOWER(e.middle_name) LIKE $1
		    OR LOWER(u.name) LIKE $1
		 ORDER BY e.last_name, e.first_name
		 LIMIT $2 OFFSET $3`,
		searchPattern, limit, offset,
	)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var employees []*domain.Employee
	for rows.Next() {
		employee := &domain.Employee{}
		university := &domain.University{}
		
		err := rows.Scan(
			&employee.ID, &employee.FirstName, &employee.LastName, &employee.MiddleName,
			&employee.Phone, &employee.MaxID, &employee.INN, &employee.KPP,
			&employee.UniversityID, &employee.CreatedAt, &employee.UpdatedAt,
			&university.ID, &university.Name, &university.INN, &university.KPP,
			&university.CreatedAt, &university.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		employee.University = university
		employees = append(employees, employee)
	}
	
	return employees, rows.Err()
}

func (r *EmployeePostgres) GetAll(limit, offset int) ([]*domain.Employee, error) {
	rows, err := r.db.Query(
		`SELECT e.id, e.first_name, e.last_name, e.middle_name, e.phone, e.max_id, e.inn, e.kpp, 
		        e.university_id, e.created_at, e.updated_at,
		        u.id, u.name, u.inn, u.kpp, u.created_at, u.updated_at
		 FROM employees e
		 JOIN universities u ON e.university_id = u.id
		 ORDER BY e.last_name, e.first_name
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var employees []*domain.Employee
	for rows.Next() {
		employee := &domain.Employee{}
		university := &domain.University{}
		
		err := rows.Scan(
			&employee.ID, &employee.FirstName, &employee.LastName, &employee.MiddleName,
			&employee.Phone, &employee.MaxID, &employee.INN, &employee.KPP,
			&employee.UniversityID, &employee.CreatedAt, &employee.UpdatedAt,
			&university.ID, &university.Name, &university.INN, &university.KPP,
			&university.CreatedAt, &university.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		employee.University = university
		employees = append(employees, employee)
	}
	
	return employees, rows.Err()
}

func (r *EmployeePostgres) Update(employee *domain.Employee) error {
	_, err := r.db.Exec(
		`UPDATE employees 
		 SET first_name = $1, last_name = $2, middle_name = $3, phone = $4, max_id = $5, 
		     inn = $6, kpp = $7, university_id = $8, updated_at = now()
		 WHERE id = $9`,
		employee.FirstName, employee.LastName, employee.MiddleName, employee.Phone,
		employee.MaxID, employee.INN, employee.KPP, employee.UniversityID, employee.ID,
	)
	return err
}

func (r *EmployeePostgres) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM employees WHERE id = $1`, id)
	return err
}

