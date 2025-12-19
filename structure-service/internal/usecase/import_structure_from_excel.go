package usecase

import (
	"fmt"
	"structure-service/internal/domain"
	"structure-service/internal/infrastructure/database"
)

// ImportStructureFromExcelUseCase импортирует структуру из Excel файла
type ImportStructureFromExcelUseCase struct {
	repo domain.StructureRepository
	db   *database.DB
}

func NewImportStructureFromExcelUseCase(repo domain.StructureRepository, db *database.DB) *ImportStructureFromExcelUseCase {
	return &ImportStructureFromExcelUseCase{
		repo: repo,
		db:   db,
	}
}

// Execute выполняет импорт структуры из Excel
func (uc *ImportStructureFromExcelUseCase) Execute(rows []*domain.ExcelRow) (*domain.ImportResult, error) {
	result := &domain.ImportResult{
		Created: 0,
		Updated: 0,
		Failed:  0,
		Errors:  []string{},
	}

	// Начинаем транзакцию
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Кэш для уже обработанных сущностей
	universitiesCache := make(map[string]*domain.University)
	branchesCache := make(map[string]*domain.Branch)
	facultiesCache := make(map[string]*domain.Faculty)

	for i, row := range rows {
		// Валидация обязательных полей
		if row.INN == "" || row.Organization == "" || row.Faculty == "" || row.GroupNumber == "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: missing required fields (INN, Organization, Faculty, GroupNumber)", i+1))
			continue
		}

		// 1. Обработка University
		universityKey := row.INN + "|" + row.KPP
		university, exists := universitiesCache[universityKey]
		if !exists {
			// Пытаемся найти существующий вуз
			var err error
			if row.KPP != "" {
				university, err = uc.repo.GetUniversityByINNAndKPP(row.INN, row.KPP)
			} else {
				university, err = uc.repo.GetUniversityByINN(row.INN)
			}

			if err == domain.ErrUniversityNotFound {
				// Создаем новый вуз
				university = &domain.University{
					Name: row.Organization,
					INN:  row.INN,
					KPP:  row.KPP,
					FOIV: row.FOIV,
				}
				if err := uc.repo.CreateUniversity(university); err != nil {
					result.Failed++
					result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to create university: %v", i+1, err))
					continue
				}
				result.Created++
			} else if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to get university: %v", i+1, err))
				continue
			} else {
				// Обновляем существующий вуз, если данные изменились
				if university.Name != row.Organization || university.FOIV != row.FOIV {
					university.Name = row.Organization
					university.FOIV = row.FOIV
					if err := uc.repo.UpdateUniversity(university); err != nil {
						result.Failed++
						result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to update university: %v", i+1, err))
						continue
					}
					result.Updated++
				}
			}
			universitiesCache[universityKey] = university
		}

		// 2. Обработка Branch (если указан)
		var branch *domain.Branch
		if row.Branch != "" {
			branchKey := fmt.Sprintf("%d|%s", university.ID, row.Branch)
			branch, exists = branchesCache[branchKey]
			if !exists {
				// Пытаемся найти существующий филиал
				var err error
				branch, err = uc.repo.GetBranchByUniversityAndName(university.ID, row.Branch)
				if err == domain.ErrBranchNotFound {
					// Создаем новый филиал
					branch = &domain.Branch{
						UniversityID: university.ID,
						Name:         row.Branch,
					}
					if err := uc.repo.CreateBranch(branch); err != nil {
						result.Failed++
						result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to create branch: %v", i+1, err))
						continue
					}
					result.Created++
				} else if err != nil {
					result.Failed++
					result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to get branch: %v", i+1, err))
					continue
				}
				branchesCache[branchKey] = branch
			}
		}

		// 3. Обработка Faculty
		var facultyKey string
		if branch != nil {
			facultyKey = fmt.Sprintf("%d|%s", branch.ID, row.Faculty)
		} else {
			facultyKey = fmt.Sprintf("nil|%s", row.Faculty)
		}
		
		faculty, exists := facultiesCache[facultyKey]
		if !exists {
			// Пытаемся найти существующий факультет
			var branchIDPtr *int64
			if branch != nil {
				branchIDPtr = &branch.ID
			}
			
			var err error
			faculty, err = uc.repo.GetFacultyByBranchAndName(branchIDPtr, row.Faculty)
			if err == domain.ErrFacultyNotFound {
				// Создаем новый факультет
				faculty = &domain.Faculty{
					Name:     row.Faculty,
					BranchID: branchIDPtr,
				}
				if err := uc.repo.CreateFaculty(faculty); err != nil {
					result.Failed++
					result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to create faculty: %v", i+1, err))
					continue
				}
				result.Created++
			} else if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to get faculty: %v", i+1, err))
				continue
			}
			facultiesCache[facultyKey] = faculty
		}

		// 4. Обработка Group
		group, err := uc.repo.GetGroupByFacultyAndNumber(faculty.ID, row.Course, row.GroupNumber)
		if err == domain.ErrGroupNotFound {
			// Создаем новую группу
			group = &domain.Group{
				FacultyID: faculty.ID,
				Course:    row.Course,
				Number:    row.GroupNumber,
				ChatURL:   row.ChatURL,
				ChatName:  row.ChatName,
			}
			if err := uc.repo.CreateGroup(group); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to create group: %v", i+1, err))
				continue
			}
			result.Created++
		} else if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to get group: %v", i+1, err))
			continue
		} else {
			// Обновляем существующую группу, если данные изменились
			if group.ChatURL != row.ChatURL || group.ChatName != row.ChatName {
				group.ChatURL = row.ChatURL
				group.ChatName = row.ChatName
				if err := uc.repo.UpdateGroup(group); err != nil {
					result.Failed++
					result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to update group: %v", i+1, err))
					continue
				}
				result.Updated++
			}
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}
