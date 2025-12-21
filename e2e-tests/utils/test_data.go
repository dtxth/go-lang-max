package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// TestUser представляет тестового пользователя
type TestUser struct {
	ID       string `json:"id,omitempty"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
}

// TestUniversity представляет тестовый университет
type TestUniversity struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
}

// TestEmployee представляет тестового сотрудника
type TestEmployee struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	MaxID    string `json:"max_id,omitempty"`
	Position string `json:"position,omitempty"`
}

// TestChat представляет тестовый чат
type TestChat struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
}

// GenerateTestUser создает тестового пользователя с уникальными данными
func GenerateTestUser() TestUser {
	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(10000)
	
	return TestUser{
		Email:    fmt.Sprintf("test%d@example.com", id),
		Phone:    fmt.Sprintf("+7900%07d", id),
		Password: "TestPassword123!",
		Role:     "operator",
	}
}

// GenerateTestUniversity создает тестовый университет
func GenerateTestUniversity() TestUniversity {
	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(1000)
	
	return TestUniversity{
		Name: fmt.Sprintf("Test University %d", id),
	}
}

// GenerateTestEmployee создает тестового сотрудника
func GenerateTestEmployee() TestEmployee {
	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(10000)
	
	return TestEmployee{
		Name:     fmt.Sprintf("Test Employee %d", id),
		Email:    fmt.Sprintf("employee%d@test.com", id),
		Phone:    fmt.Sprintf("+7900%07d", id),
		Position: "Test Position",
	}
}

// GenerateTestChat создает тестовый чат
func GenerateTestChat() TestChat {
	return TestChat{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("Test Chat %d", rand.Intn(1000)),
		Description: "Test chat description",
		Type:        "group",
	}
}

// GenerateMAXInitData создает тестовые данные для MAX авторизации
func GenerateMAXInitData() map[string]interface{} {
	return map[string]interface{}{
		"initData": "user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22Test%22%2C%22last_name%22%3A%22User%22%2C%22username%22%3A%22testuser%22%7D&auth_date=1640995200&hash=test_hash",
	}
}