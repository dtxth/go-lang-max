package repository

import (
	"database/sql"
	"structure-service/internal/domain"
	"structure-service/internal/infrastructure/database"
	"time"
)

type DepartmentManagerPostgres struct {
	db  *database.DB
	dsn string
}

func NewDepartmentManagerPostgres(db *database.DB) domain.DepartmentManagerRepository {
	return &DepartmentManagerPostgres{db: db}
}

func NewDepartmentManagerPostgresWithDSN(db *database.DB, dsn string) domain.DepartmentManagerRepository {
	return &DepartmentManagerPostgres{db: db, dsn: dsn}
}

// getDB returns a working database connection
func (r *DepartmentManagerPostgres) getDB() *database.DB {
	return r.db
}

func (r *DepartmentManagerPostgres) CreateDepartmentManager(dm *domain.DepartmentManager) error {
	query := `INSERT INTO department_managers (employee_id, branch_id, faculty_id, assigned_by, assigned_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	db := r.getDB()
	err := db.QueryRow(query, dm.EmployeeID, dm.BranchID, dm.FacultyID, dm.AssignedBy, time.Now()).Scan(&dm.ID)
	return err
}

func (r *DepartmentManagerPostgres) GetDepartmentManagerByID(id int64) (*domain.DepartmentManager, error) {
	dm := &domain.DepartmentManager{}
	query := `SELECT id, employee_id, branch_id, faculty_id, assigned_by, assigned_at 
			  FROM department_managers WHERE id = $1`
	
	var branchID, facultyID, assignedBy sql.NullInt64
	db := r.getDB()
	err := db.QueryRow(query, id).Scan(&dm.ID, &dm.EmployeeID, &branchID, &facultyID, &assignedBy, &dm.AssignedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrDepartmentManagerNotFound
	}
	if err != nil {
		return nil, err
	}
	
	if branchID.Valid {
		dm.BranchID = &branchID.Int64
	}
	if facultyID.Valid {
		dm.FacultyID = &facultyID.Int64
	}
	if assignedBy.Valid {
		dm.AssignedBy = &assignedBy.Int64
	}
	
	return dm, nil
}

func (r *DepartmentManagerPostgres) GetDepartmentManagersByEmployeeID(employeeID int64) ([]*domain.DepartmentManager, error) {
	query := `SELECT id, employee_id, branch_id, faculty_id, assigned_by, assigned_at 
			  FROM department_managers WHERE employee_id = $1 ORDER BY assigned_at DESC`
	db := r.getDB()
	rows, err := db.Query(query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var managers []*domain.DepartmentManager
	for rows.Next() {
		dm := &domain.DepartmentManager{}
		var branchID, facultyID, assignedBy sql.NullInt64
		if err := rows.Scan(&dm.ID, &dm.EmployeeID, &branchID, &facultyID, &assignedBy, &dm.AssignedAt); err != nil {
			return nil, err
		}
		
		if branchID.Valid {
			dm.BranchID = &branchID.Int64
		}
		if facultyID.Valid {
			dm.FacultyID = &facultyID.Int64
		}
		if assignedBy.Valid {
			dm.AssignedBy = &assignedBy.Int64
		}
		
		managers = append(managers, dm)
	}
	return managers, rows.Err()
}

func (r *DepartmentManagerPostgres) GetDepartmentManagersByBranchID(branchID int64) ([]*domain.DepartmentManager, error) {
	query := `SELECT id, employee_id, branch_id, faculty_id, assigned_by, assigned_at 
			  FROM department_managers WHERE branch_id = $1 ORDER BY assigned_at DESC`
	db := r.getDB()
	rows, err := db.Query(query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var managers []*domain.DepartmentManager
	for rows.Next() {
		dm := &domain.DepartmentManager{}
		var branchID, facultyID, assignedBy sql.NullInt64
		if err := rows.Scan(&dm.ID, &dm.EmployeeID, &branchID, &facultyID, &assignedBy, &dm.AssignedAt); err != nil {
			return nil, err
		}
		
		if branchID.Valid {
			dm.BranchID = &branchID.Int64
		}
		if facultyID.Valid {
			dm.FacultyID = &facultyID.Int64
		}
		if assignedBy.Valid {
			dm.AssignedBy = &assignedBy.Int64
		}
		
		managers = append(managers, dm)
	}
	return managers, rows.Err()
}

func (r *DepartmentManagerPostgres) GetDepartmentManagersByFacultyID(facultyID int64) ([]*domain.DepartmentManager, error) {
	query := `SELECT id, employee_id, branch_id, faculty_id, assigned_by, assigned_at 
			  FROM department_managers WHERE faculty_id = $1 ORDER BY assigned_at DESC`
	db := r.getDB()
	rows, err := db.Query(query, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var managers []*domain.DepartmentManager
	for rows.Next() {
		dm := &domain.DepartmentManager{}
		var branchID, facultyID, assignedBy sql.NullInt64
		if err := rows.Scan(&dm.ID, &dm.EmployeeID, &branchID, &facultyID, &assignedBy, &dm.AssignedAt); err != nil {
			return nil, err
		}
		
		if branchID.Valid {
			dm.BranchID = &branchID.Int64
		}
		if facultyID.Valid {
			dm.FacultyID = &facultyID.Int64
		}
		if assignedBy.Valid {
			dm.AssignedBy = &assignedBy.Int64
		}
		
		managers = append(managers, dm)
	}
	return managers, rows.Err()
}

func (r *DepartmentManagerPostgres) GetAllDepartmentManagers() ([]*domain.DepartmentManager, error) {
	query := `SELECT id, employee_id, branch_id, faculty_id, assigned_by, assigned_at 
			  FROM department_managers ORDER BY assigned_at DESC`
	db := r.getDB()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var managers []*domain.DepartmentManager
	for rows.Next() {
		dm := &domain.DepartmentManager{}
		var branchID, facultyID, assignedBy sql.NullInt64
		if err := rows.Scan(&dm.ID, &dm.EmployeeID, &branchID, &facultyID, &assignedBy, &dm.AssignedAt); err != nil {
			return nil, err
		}
		
		if branchID.Valid {
			dm.BranchID = &branchID.Int64
		}
		if facultyID.Valid {
			dm.FacultyID = &facultyID.Int64
		}
		if assignedBy.Valid {
			dm.AssignedBy = &assignedBy.Int64
		}
		
		managers = append(managers, dm)
	}
	return managers, rows.Err()
}

func (r *DepartmentManagerPostgres) DeleteDepartmentManager(id int64) error {
	db := r.getDB()
	result, err := db.Exec("DELETE FROM department_managers WHERE id = $1", id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return domain.ErrDepartmentManagerNotFound
	}
	
	return nil
}
