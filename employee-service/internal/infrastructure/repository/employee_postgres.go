package repository

import (
	"employee-service/internal/domain"
	"employee-service/internal/infrastructure/database"
	"strconv"
	"strings"
)

type EmployeePostgres struct {
	db  *database.DB
	dsn string
}

func NewEmployeePostgres(db *database.DB) *EmployeePostgres {
	return &EmployeePostgres{db: db}
}

func NewEmployeePostgresWithDSN(db *database.DB, dsn string) *EmployeePostgres {
	return &EmployeePostgres{db: db, dsn: dsn}
}

// getDB returns a working database connection
func (r *EmployeePostgres) getDB() *database.DB {
	return r.db
}

// scanEmployeeWithUniversity сканирует строку результата в Employee с University
func (r *EmployeePostgres) scanEmployeeWithUniversity(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Employee, error) {
	employee := &domain.Employee{}
	university := &domain.University{}
	
	err := scanner.Scan(
		&employee.ID, &employee.FirstName, &employee.LastName, &employee.MiddleName,
		&employee.Phone, &employee.MaxID, &employee.INN, &employee.KPP,
		&employee.UniversityID, &employee.Role, &employee.UserID, &employee.MaxIDUpdatedAt,
		&employee.ProfileSource, &employee.ProfileLastUpdated,
		&employee.CreatedAt, &employee.UpdatedAt,
		&university.ID, &university.Name, &university.INN, &university.KPP,
		&university.CreatedAt, &university.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	employee.University = university
	return employee, nil
}

// employeeSelectQuery возвращает стандартный SELECT запрос для сотрудников
func (r *EmployeePostgres) employeeSelectQuery() string {
	return `SELECT e.id, e.first_name, e.last_name, e.middle_name, e.phone, e.max_id, e.inn, e.kpp, 
		        e.university_id, e.role, e.user_id, e.max_id_updated_at, e.profile_source, e.profile_last_updated, e.created_at, e.updated_at,
		        u.id, u.name, u.inn, u.kpp, u.created_at, u.updated_at
		 FROM employees e
		 JOIN universities u ON e.university_id = u.id`
}

func (r *EmployeePostgres) Create(employee *domain.Employee) error {
	db := r.getDB()
	err := db.QueryRow(
		`INSERT INTO employees (first_name, last_name, middle_name, phone, max_id, inn, kpp, university_id, role, user_id, max_id_updated_at, profile_source, profile_last_updated) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id, created_at, updated_at`,
		employee.FirstName, employee.LastName, employee.MiddleName, employee.Phone,
		employee.MaxID, employee.INN, employee.KPP, employee.UniversityID,
		employee.Role, employee.UserID, employee.MaxIDUpdatedAt,
		employee.ProfileSource, employee.ProfileLastUpdated,
	).Scan(&employee.ID, &employee.CreatedAt, &employee.UpdatedAt)
	return err
}

func (r *EmployeePostgres) GetByID(id int64) (*domain.Employee, error) {
	query := r.employeeSelectQuery() + " WHERE e.id = $1"
	db := r.getDB()
	row := db.QueryRow(query, id)
	return r.scanEmployeeWithUniversity(row)
}

func (r *EmployeePostgres) GetByPhone(phone string) (*domain.Employee, error) {
	query := r.employeeSelectQuery() + " WHERE e.phone = $1"
	db := r.getDB()
	row := db.QueryRow(query, phone)
	return r.scanEmployeeWithUniversity(row)
}

func (r *EmployeePostgres) GetByMaxID(maxID string) (*domain.Employee, error) {
	query := r.employeeSelectQuery() + " WHERE e.max_id = $1"
	db := r.getDB()
	row := db.QueryRow(query, maxID)
	return r.scanEmployeeWithUniversity(row)
}

func (r *EmployeePostgres) Search(query string, limit, offset int) ([]*domain.Employee, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return r.GetAll(limit, offset)
	}
	
	searchPattern := "%" + strings.ToLower(query) + "%"
	
	sqlQuery := r.employeeSelectQuery() + `
		 WHERE LOWER(e.first_name) LIKE $1 
		    OR LOWER(e.last_name) LIKE $1 
		    OR LOWER(e.middle_name) LIKE $1
		    OR LOWER(u.name) LIKE $1
		 ORDER BY e.last_name, e.first_name
		 LIMIT $2 OFFSET $3`
	
	db := r.getDB()
	rows, err := db.Query(sqlQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var employees []*domain.Employee
	for rows.Next() {
		employee, err := r.scanEmployeeWithUniversity(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}
	
	return employees, rows.Err()
}

func (r *EmployeePostgres) GetAll(limit, offset int) ([]*domain.Employee, error) {
	sqlQuery := r.employeeSelectQuery() + `
		 ORDER BY e.last_name, e.first_name
		 LIMIT $1 OFFSET $2`
	
	db := r.getDB()
	rows, err := db.Query(sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var employees []*domain.Employee
	for rows.Next() {
		employee, err := r.scanEmployeeWithUniversity(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}
	
	return employees, rows.Err()
}

func (r *EmployeePostgres) Update(employee *domain.Employee) error {
	db := r.getDB()
	_, err := db.Exec(
		`UPDATE employees 
		 SET first_name = $1, last_name = $2, middle_name = $3, phone = $4, max_id = $5, 
		     inn = $6, kpp = $7, university_id = $8, role = $9, user_id = $10, max_id_updated_at = $11,
		     profile_source = $12, profile_last_updated = $13, updated_at = now()
		 WHERE id = $14`,
		employee.FirstName, employee.LastName, employee.MiddleName, employee.Phone,
		employee.MaxID, employee.INN, employee.KPP, employee.UniversityID,
		employee.Role, employee.UserID, employee.MaxIDUpdatedAt,
		employee.ProfileSource, employee.ProfileLastUpdated, employee.ID,
	)
	return err
}

func (r *EmployeePostgres) Delete(id int64) error {
	db := r.getDB()
	_, err := db.Exec(`DELETE FROM employees WHERE id = $1`, id)
	return err
}

func (r *EmployeePostgres) GetEmployeesWithoutMaxID(limit, offset int) ([]*domain.Employee, error) {
	sqlQuery := r.employeeSelectQuery() + `
		 WHERE e.max_id IS NULL OR e.max_id = ''
		 ORDER BY e.last_name, e.first_name
		 LIMIT $1 OFFSET $2`
	
	db := r.getDB()
	rows, err := db.Query(sqlQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var employees []*domain.Employee
	for rows.Next() {
		employee, err := r.scanEmployeeWithUniversity(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}
	
	return employees, rows.Err()
}

func (r *EmployeePostgres) CountEmployeesWithoutMaxID() (int, error) {
	var count int
	db := r.getDB()
	err := db.QueryRow(
		`SELECT COUNT(*) FROM employees WHERE max_id IS NULL OR max_id = ''`,
	).Scan(&count)
	return count, err
}

func (r *EmployeePostgres) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.Employee, error) {
	// Валидация параметров сортировки
	validSortFields := map[string]string{
		"first_name":    "e.first_name",
		"last_name":     "e.last_name",
		"phone":         "e.phone",
		"max_id":        "e.max_id",
		"university":    "u.name",
		"created_at":    "e.created_at",
		"profile_source": "e.profile_source",
	}
	
	sortField, ok := validSortFields[sortBy]
	if !ok {
		sortField = "e.last_name"
	}
	
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "ASC"
	}
	
	// Построение WHERE условия для поиска
	var whereClause string
	var args []interface{}
	argIndex := 0
	
	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		whereClause = `WHERE LOWER(e.first_name) LIKE $1 
		                  OR LOWER(e.last_name) LIKE $1 
		                  OR LOWER(e.middle_name) LIKE $1
		                  OR LOWER(u.name) LIKE $1`
		args = append(args, searchPattern)
		argIndex++
	}
	
	// Добавляем LIMIT и OFFSET
	limitArg := "$" + strconv.Itoa(argIndex+1)
	offsetArg := "$" + strconv.Itoa(argIndex+2)
	args = append(args, limit, offset)
	
	query := r.employeeSelectQuery()
	if whereClause != "" {
		query += " " + whereClause
	}
	query += " ORDER BY " + sortField + " " + sortOrder + ", e.first_name ASC"
	query += " LIMIT " + limitArg + " OFFSET " + offsetArg
	
	db := r.getDB()
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var employees []*domain.Employee
	for rows.Next() {
		employee, err := r.scanEmployeeWithUniversity(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}
	
	return employees, rows.Err()
}

func (r *EmployeePostgres) CountAllWithSearch(search string) (int, error) {
	var count int
	var err error
	db := r.getDB()
	
	if search == "" {
		err = db.QueryRow(`SELECT COUNT(*) FROM employees`).Scan(&count)
	} else {
		searchPattern := "%" + strings.ToLower(search) + "%"
		err = db.QueryRow(
			`SELECT COUNT(*) FROM employees e
			 JOIN universities u ON e.university_id = u.id
			 WHERE LOWER(e.first_name) LIKE $1 
			    OR LOWER(e.last_name) LIKE $1 
			    OR LOWER(e.middle_name) LIKE $1
			    OR LOWER(u.name) LIKE $1`,
			searchPattern,
		).Scan(&count)
	}
	
	return count, err
}