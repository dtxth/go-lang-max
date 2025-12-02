package domain

// UniversityRepository определяет интерфейс для работы с вузами
type UniversityRepository interface {
	// Create создает новый вуз
	Create(university *University) error
	
	// GetByID получает вуз по ID
	GetByID(id int64) (*University, error)
	
	// GetByINN получает вуз по ИНН
	GetByINN(inn string) (*University, error)
	
	// GetByINNAndKPP получает вуз по ИНН и КПП
	GetByINNAndKPP(inn, kpp string) (*University, error)
	
	// SearchByName выполняет поиск вузов по названию
	SearchByName(query string) ([]*University, error)
	
	// GetAll получает все вузы
	GetAll() ([]*University, error)
}

