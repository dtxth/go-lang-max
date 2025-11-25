package domain

// StructureRepository определяет интерфейс для работы со структурой вуза
type StructureRepository interface {
	// University
	CreateUniversity(university *University) error
	GetUniversityByID(id int64) (*University, error)
	GetUniversityByINN(inn string) (*University, error)
	UpdateUniversity(university *University) error
	DeleteUniversity(id int64) error

	// Branch
	CreateBranch(branch *Branch) error
	GetBranchByID(id int64) (*Branch, error)
	GetBranchesByUniversityID(universityID int64) ([]*Branch, error)
	UpdateBranch(branch *Branch) error
	DeleteBranch(id int64) error

	// Faculty
	CreateFaculty(faculty *Faculty) error
	GetFacultyByID(id int64) (*Faculty, error)
	GetFacultiesByBranchID(branchID int64) ([]*Faculty, error)
	GetFacultiesByUniversityID(universityID int64) ([]*Faculty, error)
	UpdateFaculty(faculty *Faculty) error
	DeleteFaculty(id int64) error

	// Group
	CreateGroup(group *Group) error
	GetGroupByID(id int64) (*Group, error)
	GetGroupsByFacultyID(facultyID int64) ([]*Group, error)
	UpdateGroup(group *Group) error
	DeleteGroup(id int64) error

	// Получение полной структуры
	GetStructureByUniversityID(universityID int64) (*StructureNode, error)
	GetAllUniversities() ([]*University, error)
}

