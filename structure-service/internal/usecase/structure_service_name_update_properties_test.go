package usecase

import (
	"testing"
	"time"
	"unicode/utf8"

	"structure-service/internal/domain"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/mock"
)

// Property-based tests for name update methods

func TestProperty_UpdateUniversityName_PreservesOtherFields(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("UpdateUniversityName preserves all fields except name", prop.ForAll(
		func(id int64, originalName, newName, inn, kpp, foiv string) bool {
			if id <= 0 {
				return true // Skip invalid IDs
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			originalUniversity := &domain.University{
				ID:        id,
				Name:      originalName,
				INN:       inn,
				KPP:       kpp,
				FOIV:      foiv,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Mock expects the university with updated name but same other fields
			mockRepo.On("GetUniversityByID", id).Return(originalUniversity, nil)
			mockRepo.On("UpdateUniversity", mock.MatchedBy(func(u *domain.University) bool {
				return u.ID == id &&
					u.Name == newName &&
					u.INN == inn &&
					u.KPP == kpp &&
					u.FOIV == foiv &&
					u.CreatedAt.Equal(originalUniversity.CreatedAt) &&
					u.UpdatedAt.Equal(originalUniversity.UpdatedAt)
			})).Return(nil)

			err := service.UpdateUniversityName(id, newName)
			return err == nil
		},
		gen.Int64Range(1, 1000),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.RegexMatch("[0-9]{10}"),
		gen.RegexMatch("[0-9]{9}"),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestProperty_UpdateBranchName_PreservesOtherFields(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("UpdateBranchName preserves all fields except name", prop.ForAll(
		func(id, universityID int64, originalName, newName string) bool {
			if id <= 0 || universityID <= 0 {
				return true // Skip invalid IDs
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			originalBranch := &domain.Branch{
				ID:           id,
				UniversityID: universityID,
				Name:         originalName,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			mockRepo.On("GetBranchByID", id).Return(originalBranch, nil)
			mockRepo.On("UpdateBranch", mock.MatchedBy(func(b *domain.Branch) bool {
				return b.ID == id &&
					b.UniversityID == universityID &&
					b.Name == newName &&
					b.CreatedAt.Equal(originalBranch.CreatedAt) &&
					b.UpdatedAt.Equal(originalBranch.UpdatedAt)
			})).Return(nil)

			err := service.UpdateBranchName(id, newName)
			return err == nil
		},
		gen.Int64Range(1, 1000),
		gen.Int64Range(1, 100),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestProperty_UpdateFacultyName_PreservesOtherFields(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("UpdateFacultyName preserves all fields except name", prop.ForAll(
		func(id int64, branchID *int64, originalName, newName string) bool {
			if id <= 0 {
				return true // Skip invalid IDs
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			originalFaculty := &domain.Faculty{
				ID:        id,
				BranchID:  branchID,
				Name:      originalName,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockRepo.On("GetFacultyByID", id).Return(originalFaculty, nil)
			mockRepo.On("UpdateFaculty", mock.MatchedBy(func(f *domain.Faculty) bool {
				branchIDsEqual := (f.BranchID == nil && branchID == nil) ||
					(f.BranchID != nil && branchID != nil && *f.BranchID == *branchID)

				return f.ID == id &&
					branchIDsEqual &&
					f.Name == newName &&
					f.CreatedAt.Equal(originalFaculty.CreatedAt) &&
					f.UpdatedAt.Equal(originalFaculty.UpdatedAt)
			})).Return(nil)

			err := service.UpdateFacultyName(id, newName)
			return err == nil
		},
		gen.Int64Range(1, 1000),
		gen.PtrOf(gen.Int64Range(1, 100)),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestProperty_UpdateGroupName_PreservesOtherFields(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("UpdateGroupName preserves all fields except number", prop.ForAll(
		func(id, facultyID int64, course int, originalNumber, newNumber, chatURL, chatName string, chatID *int64) bool {
			if id <= 0 || facultyID <= 0 || course < 0 || course > 10 {
				return true // Skip invalid values
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			originalGroup := &domain.Group{
				ID:        id,
				FacultyID: facultyID,
				Course:    course,
				Number:    originalNumber,
				ChatID:    chatID,
				ChatURL:   chatURL,
				ChatName:  chatName,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockRepo.On("GetGroupByID", id).Return(originalGroup, nil)
			mockRepo.On("UpdateGroup", mock.MatchedBy(func(g *domain.Group) bool {
				chatIDsEqual := (g.ChatID == nil && chatID == nil) ||
					(g.ChatID != nil && chatID != nil && *g.ChatID == *chatID)

				return g.ID == id &&
					g.FacultyID == facultyID &&
					g.Course == course &&
					g.Number == newNumber &&
					chatIDsEqual &&
					g.ChatURL == chatURL &&
					g.ChatName == chatName &&
					g.CreatedAt.Equal(originalGroup.CreatedAt) &&
					g.UpdatedAt.Equal(originalGroup.UpdatedAt)
			})).Return(nil)

			err := service.UpdateGroupName(id, newNumber)
			return err == nil
		},
		gen.Int64Range(1, 1000),
		gen.Int64Range(1, 100),
		gen.IntRange(1, 6),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.PtrOf(gen.Int64Range(1, 10000)),
	))

	properties.TestingRun(t)
}

func TestProperty_NameUpdate_HandlesUnicodeCorrectly(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Name updates handle Unicode strings correctly", prop.ForAll(
		func(unicodeName string) bool {
			// Skip invalid UTF-8 strings
			if !utf8.ValidString(unicodeName) {
				return true
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			university := &domain.University{
				ID:        1,
				Name:      "Old Name",
				INN:       "1234567890",
				KPP:       "123456789",
				FOIV:      "Test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
			mockRepo.On("UpdateUniversity", mock.MatchedBy(func(u *domain.University) bool {
				return u.Name == unicodeName
			})).Return(nil)

			err := service.UpdateUniversityName(1, unicodeName)
			return err == nil
		},
		gen.RegexMatch("[\u0400-\u04FF\u0020-\u007F]{1,100}"), // Cyrillic and ASCII
	))

	properties.TestingRun(t)
}

func TestProperty_NameUpdate_HandlesEmptyAndWhitespace(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Name updates handle empty and whitespace strings", prop.ForAll(
		func(whitespaceCount int) bool {
			if whitespaceCount < 0 || whitespaceCount > 100 {
				return true
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			// Create string with only whitespace
			whitespaceString := ""
			for i := 0; i < whitespaceCount; i++ {
				whitespaceString += " "
			}

			university := &domain.University{
				ID:        1,
				Name:      "Old Name",
				INN:       "1234567890",
				KPP:       "123456789",
				FOIV:      "Test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
			mockRepo.On("UpdateUniversity", mock.MatchedBy(func(u *domain.University) bool {
				return u.Name == whitespaceString
			})).Return(nil)

			err := service.UpdateUniversityName(1, whitespaceString)
			return err == nil
		},
		gen.IntRange(0, 50),
	))

	properties.TestingRun(t)
}

func TestProperty_NameUpdate_IdempotentOperation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Updating name to same value is idempotent", prop.ForAll(
		func(id int64, name string) bool {
			if id <= 0 {
				return true
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			university := &domain.University{
				ID:        id,
				Name:      name,
				INN:       "1234567890",
				KPP:       "123456789",
				FOIV:      "Test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Should still call update even if name is the same
			mockRepo.On("GetUniversityByID", id).Return(university, nil)
			mockRepo.On("UpdateUniversity", mock.MatchedBy(func(u *domain.University) bool {
				return u.Name == name
			})).Return(nil)

			err := service.UpdateUniversityName(id, name)
			return err == nil
		},
		gen.Int64Range(1, 1000),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestProperty_NameUpdate_ErrorPropagation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Repository errors are properly propagated", prop.ForAll(
		func(id int64, name string) bool {
			if id <= 0 {
				return true
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			// Test GetByID error propagation
			mockRepo.On("GetUniversityByID", id).Return(nil, domain.ErrUniversityNotFound)

			err := service.UpdateUniversityName(id, name)
			return err == domain.ErrUniversityNotFound
		},
		gen.Int64Range(1, 1000),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestProperty_NameUpdate_LongNames(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Name updates handle very long names", prop.ForAll(
		func(nameLength int) bool {
			if nameLength <= 0 || nameLength > 1000 {
				return true
			}

			mockRepo := new(MockStructureRepository)
			service := NewStructureService(mockRepo)

			// Create long name
			longName := ""
			for i := 0; i < nameLength; i++ {
				longName += "a"
			}

			university := &domain.University{
				ID:        1,
				Name:      "Old Name",
				INN:       "1234567890",
				KPP:       "123456789",
				FOIV:      "Test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
			mockRepo.On("UpdateUniversity", mock.MatchedBy(func(u *domain.University) bool {
				return u.Name == longName && len(u.Name) == nameLength
			})).Return(nil)

			err := service.UpdateUniversityName(1, longName)
			return err == nil
		},
		gen.IntRange(1, 500),
	))

	properties.TestingRun(t)
}