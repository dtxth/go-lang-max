package domain

// StructureServiceInterface определяет интерфейс для сервиса структуры
type StructureServiceInterface interface {
	// GetStructure получает полную иерархическую структуру вуза
	GetStructure(universityID int64) (*StructureNode, error)
	
	// GetAllUniversities получает список всех вузов
	GetAllUniversities() ([]*University, error)
	
	// GetAllUniversitiesWithSortingAndSearch получает список всех вузов с пагинацией, сортировкой и поиском
	GetAllUniversitiesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*University, int, error)
	
	// GetUniversity получает вуз по ID
	GetUniversity(id int64) (*University, error)
	
	// GetUniversityByINN получает вуз по ИНН
	GetUniversityByINN(inn string) (*University, error)
	
	// CreateUniversity создает новый вуз
	CreateUniversity(u *University) error
	
	// UpdateUniversity обновляет информацию о вузе
	UpdateUniversity(u *University) error
	
	// DeleteUniversity удаляет вуз
	DeleteUniversity(id int64) error
	
	// CreateBranch создает новый филиал
	CreateBranch(b *Branch) error
	
	// UpdateBranch обновляет информацию о филиале
	UpdateBranch(b *Branch) error
	
	// DeleteBranch удаляет филиал
	DeleteBranch(id int64) error
	
	// CreateFaculty создает новый факультет
	CreateFaculty(f *Faculty) error
	
	// UpdateFaculty обновляет информацию о факультете
	UpdateFaculty(f *Faculty) error
	
	// DeleteFaculty удаляет факультет
	DeleteFaculty(id int64) error
	
	// CreateGroup создает новую группу
	CreateGroup(g *Group) error
	
	// GetGroupByID получает группу по ID
	GetGroupByID(id int64) (*Group, error)
	
	// UpdateGroup обновляет информацию о группе
	UpdateGroup(g *Group) error
	
	// DeleteGroup удаляет группу
	DeleteGroup(id int64) error
	
	// ImportFromExcel импортирует структуру из Excel файла
	ImportFromExcel(rows []*ExcelRow) error
}