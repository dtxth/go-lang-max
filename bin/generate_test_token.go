package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Секрет из docker-compose.yml для auth-service
	accessSecret := []byte("your-super-secret-jwt-key-change-in-production")
	
	// Параметры токена
	userID := int64(1)
	email := "test@example.com"
	role := "superadmin" // superadmin, curator, operator
	
	// Контекст (опционально)
	var universityID *int64
	var branchID *int64
	var facultyID *int64
	
	// Для curator или operator можно указать university_id
	// uniID := int64(1)
	// universityID = &uniID
	
	now := time.Now()
	accessTTL := 15 * time.Minute
	
	// Создаем claims
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", userID),
		"email": email,
		"role":  role,
		"exp":   now.Add(accessTTL).Unix(),
		"iat":   now.Unix(),
	}
	
	// Добавляем контекстную информацию, если есть
	if universityID != nil {
		claims["university_id"] = *universityID
	}
	if branchID != nil {
		claims["branch_id"] = *branchID
	}
	if facultyID != nil {
		claims["faculty_id"] = *facultyID
	}
	
	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(accessSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("=== Test Token Generated ===")
	fmt.Printf("Role: %s\n", role)
	fmt.Printf("User ID: %d\n", userID)
	fmt.Printf("Email: %s\n", email)
	if universityID != nil {
		fmt.Printf("University ID: %d\n", *universityID)
	}
	fmt.Printf("Expires: %s\n", now.Add(accessTTL).Format(time.RFC3339))
	fmt.Println("\nToken:")
	fmt.Println(tokenString)
	fmt.Println("\n=== Usage ===")
	fmt.Println("curl -H \"Authorization: Bearer " + tokenString + "\" http://localhost:8082/chats/all")
}
