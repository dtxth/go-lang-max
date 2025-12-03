package usecase

import (
	"strconv"
	"structure-service/internal/domain"
)

type StructureService struct {
	repo domain.StructureRepository
}

func NewStructureService(repo domain.StructureRepository) *StructureService {
	return &StructureService{repo: repo}
}

// GetStructure получает полную иерархическую структуру вуза
func (s *StructureService) GetStructure(universityID int64) (*domain.StructureNode, error) {
	return s.repo.GetStructureByUniversityID(universityID)
}

// GetAllUniversities получает список всех вузов
func (s *StructureService) GetAllUniversities() ([]*domain.University, error) {
	return s.repo.GetAllUniversities()
}

// GetUniversity получает вуз по ID
func (s *StructureService) GetUniversity(id int64) (*domain.University, error) {
	return s.repo.GetUniversityByID(id)
}

// GetUniversityByINN получает вуз по ИНН
func (s *StructureService) GetUniversityByINN(inn string) (*domain.University, error) {
	return s.repo.GetUniversityByINN(inn)
}

// CreateUniversity создает новый вуз
func (s *StructureService) CreateUniversity(u *domain.University) error {
	return s.repo.CreateUniversity(u)
}

// UpdateUniversity обновляет информацию о вузе
func (s *StructureService) UpdateUniversity(u *domain.University) error {
	return s.repo.UpdateUniversity(u)
}

// DeleteUniversity удаляет вуз
func (s *StructureService) DeleteUniversity(id int64) error {
	return s.repo.DeleteUniversity(id)
}

// CreateBranch создает новый филиал
func (s *StructureService) CreateBranch(b *domain.Branch) error {
	return s.repo.CreateBranch(b)
}

// UpdateBranch обновляет информацию о филиале
func (s *StructureService) UpdateBranch(b *domain.Branch) error {
	return s.repo.UpdateBranch(b)
}

// DeleteBranch удаляет филиал
func (s *StructureService) DeleteBranch(id int64) error {
	return s.repo.DeleteBranch(id)
}

// CreateFaculty создает новый факультет
func (s *StructureService) CreateFaculty(f *domain.Faculty) error {
	return s.repo.CreateFaculty(f)
}

// UpdateFaculty обновляет информацию о факультете
func (s *StructureService) UpdateFaculty(f *domain.Faculty) error {
	return s.repo.UpdateFaculty(f)
}

// DeleteFaculty удаляет факультет
func (s *StructureService) DeleteFaculty(id int64) error {
	return s.repo.DeleteFaculty(id)
}

// CreateGroup создает новую группу
func (s *StructureService) CreateGroup(g *domain.Group) error {
	return s.repo.CreateGroup(g)
}

// GetGroupByID получает группу по ID
func (s *StructureService) GetGroupByID(id int64) (*domain.Group, error) {
	return s.repo.GetGroupByID(id)
}

// UpdateGroup обновляет информацию о группе
func (s *StructureService) UpdateGroup(g *domain.Group) error {
	return s.repo.UpdateGroup(g)
}

// DeleteGroup удаляет группу
func (s *StructureService) DeleteGroup(id int64) error {
	return s.repo.DeleteGroup(id)
}

// ImportFromExcel импортирует структуру из Excel файла
func (s *StructureService) ImportFromExcel(rows []*domain.ExcelRow) error {
	// Группируем по вузам (по ИНН)
	universitiesMap := make(map[string]*domain.University)
	branchesMap := make(map[string]*domain.Branch)
	facultiesMap := make(map[string]*domain.Faculty)
	groupsMap := make(map[string]*domain.Group)

	for _, row := range rows {
		// Создаем или получаем вуз
		university, exists := universitiesMap[row.INN]
		if !exists {
			university = &domain.University{
				Name: row.Organization,
				INN:  row.INN,
				KPP:  row.KPP,
				FOIV: row.FOIV,
			}
			// Проверяем, существует ли уже вуз с таким ИНН
			existing, err := s.repo.GetUniversityByINN(row.INN)
			if err == nil {
				university = existing
			} else {
				if err := s.repo.CreateUniversity(university); err != nil {
					return err
				}
			}
			universitiesMap[row.INN] = university
		}

		// Создаем или получаем филиал (если указан)
		var branch *domain.Branch
		if row.Branch != "" {
			branchKey := row.INN + "|" + row.Branch
			branch, exists = branchesMap[branchKey]
			if !exists {
				branch = &domain.Branch{
					UniversityID: university.ID,
					Name:         row.Branch,
				}
				if err := s.repo.CreateBranch(branch); err != nil {
					return err
				}
				branchesMap[branchKey] = branch
			}
		}

		// Создаем или получаем факультет
		facultyKey := row.INN
		if branch != nil {
			facultyKey += "|" + strconv.FormatInt(branch.ID, 10) + "|" + row.Faculty
		} else {
			facultyKey += "||" + row.Faculty
		}
		faculty, exists := facultiesMap[facultyKey]
		if !exists {
			faculty = &domain.Faculty{
				Name: row.Faculty,
			}
			if branch != nil {
				faculty.BranchID = &branch.ID
			}
			if err := s.repo.CreateFaculty(faculty); err != nil {
				return err
			}
			facultiesMap[facultyKey] = faculty
		}

		// Создаем группу
		groupKey := strconv.FormatInt(faculty.ID, 10) + "|" + row.GroupNumber + "|" + strconv.Itoa(row.Course)
		_, exists = groupsMap[groupKey]
		if !exists {
			group := &domain.Group{
				FacultyID: faculty.ID,
				Course:    row.Course,
				Number:    row.GroupNumber,
			}
			// Если есть ID чата, привязываем его
			if row.ChatID != "" {
				// Здесь можно добавить логику для получения chat_id из chat-service
				// Пока оставляем NULL
			}
			if err := s.repo.CreateGroup(group); err != nil {
				return err
			}
			groupsMap[groupKey] = group
		}
	}

	return nil
}

