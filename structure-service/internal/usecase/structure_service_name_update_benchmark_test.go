package usecase

import (
	"testing"
	"time"

	"structure-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

// Benchmark tests for name update methods

func BenchmarkUpdateUniversityName(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	university := &domain.University{
		ID:        1,
		Name:      "Original University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
	mockRepo.On("UpdateUniversity", mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.UpdateUniversityName(1, "New University Name")
	}
}

func BenchmarkUpdateBranchName(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	branch := &domain.Branch{
		ID:           1,
		UniversityID: 1,
		Name:         "Original Branch Name",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockRepo.On("GetBranchByID", int64(1)).Return(branch, nil)
	mockRepo.On("UpdateBranch", mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.UpdateBranchName(1, "New Branch Name")
	}
}

func BenchmarkUpdateFacultyName(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	branchID := int64(1)
	faculty := &domain.Faculty{
		ID:        1,
		BranchID:  &branchID,
		Name:      "Original Faculty Name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetFacultyByID", int64(1)).Return(faculty, nil)
	mockRepo.On("UpdateFaculty", mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.UpdateFacultyName(1, "New Faculty Name")
	}
}

func BenchmarkUpdateGroupName(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	chatID := int64(123)
	group := &domain.Group{
		ID:        1,
		FacultyID: 1,
		Course:    2,
		Number:    "ВБ-21",
		ChatID:    &chatID,
		ChatURL:   "https://max.ru/join/test",
		ChatName:  "Test Chat",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetGroupByID", int64(1)).Return(group, nil)
	mockRepo.On("UpdateGroup", mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.UpdateGroupName(1, "ВБ-22")
	}
}

func BenchmarkGetBranchByID(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	branch := &domain.Branch{
		ID:           1,
		UniversityID: 1,
		Name:         "Test Branch",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockRepo.On("GetBranchByID", int64(1)).Return(branch, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetBranchByID(1)
	}
}

func BenchmarkGetFacultyByID(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	branchID := int64(1)
	faculty := &domain.Faculty{
		ID:        1,
		BranchID:  &branchID,
		Name:      "Test Faculty",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetFacultyByID", int64(1)).Return(faculty, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetFacultyByID(1)
	}
}

// Benchmark with different name lengths
func BenchmarkUpdateUniversityName_ShortName(b *testing.B) {
	benchmarkUpdateUniversityNameWithLength(b, "Short")
}

func BenchmarkUpdateUniversityName_MediumName(b *testing.B) {
	benchmarkUpdateUniversityNameWithLength(b, "Medium Length University Name")
}

func BenchmarkUpdateUniversityName_LongName(b *testing.B) {
	longName := "Very Long University Name That Contains Many Words And Characters To Test Performance With Longer Strings"
	benchmarkUpdateUniversityNameWithLength(b, longName)
}

func BenchmarkUpdateUniversityName_VeryLongName(b *testing.B) {
	veryLongName := ""
	for i := 0; i < 500; i++ {
		veryLongName += "A"
	}
	benchmarkUpdateUniversityNameWithLength(b, veryLongName)
}

func benchmarkUpdateUniversityNameWithLength(b *testing.B, newName string) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	university := &domain.University{
		ID:        1,
		Name:      "Original University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
	mockRepo.On("UpdateUniversity", mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.UpdateUniversityName(1, newName)
	}
}

// Benchmark with Unicode names
func BenchmarkUpdateUniversityName_UnicodeRussian(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	university := &domain.University{
		ID:        1,
		Name:      "Original University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	russianName := "Федеральное государственное бюджетное образовательное учреждение высшего образования"

	mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
	mockRepo.On("UpdateUniversity", mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.UpdateUniversityName(1, russianName)
	}
}

func BenchmarkUpdateUniversityName_UnicodeEmoji(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	university := &domain.University{
		ID:        1,
		Name:      "Original University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	emojiName := "University 🎓 with 📚 emojis 🏫"

	mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
	mockRepo.On("UpdateUniversity", mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.UpdateUniversityName(1, emojiName)
	}
}

// Benchmark concurrent access simulation
func BenchmarkUpdateUniversityName_Parallel(b *testing.B) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	university := &domain.University{
		ID:        1,
		Name:      "Original University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
	mockRepo.On("UpdateUniversity", mock.Anything).Return(nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = service.UpdateUniversityName(1, "New University Name")
		}
	})
}