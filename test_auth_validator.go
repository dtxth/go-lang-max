package main

import (
	"fmt"
	"log"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Импортируем наш валидатор
import "auth-service/internal/infrastructure/max"

func main() {
	fmt.Println("🧪 Тест MAX Auth Validator")
	
	// Создаем валидатор
	validator := max.NewAuthValidator()
	botToken := "test_bot_token_123"
	
	// Тестовые данные
	maxID := "18963527"
	firstName := "Андрей"
	lastName := "Тестов"
	username := "testuser"
	
	fmt.Printf("📋 Тестовые данные:\n")
	fmt.Printf("   MAX ID: %s\n", maxID)
	fmt.Printf("   Имя: %s\n", firstName)
	fmt.Printf("   Фамилия: %s\n", lastName)
	fmt.Printf("   Username: %s\n", username)
	fmt.Printf("   Bot Token: %s\n", botToken)
	fmt.Println()
	
	// Создаем валидную init_data
	fmt.Println("🔧 Создание валидной init_data...")
	initData := createValidInitData(maxID, firstName, lastName, username, botToken)
	fmt.Printf("✅ Init data создана (длина: %d)\n", len(initData))
	fmt.Println()
	
	// Тестируем валидатор
	fmt.Println("🔍 Тестирование валидатора...")
	userData, err := validator.ValidateInitData(initData, botToken)
	if err != nil {
		fmt.Printf("❌ Ошибка валидации: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Валидация успешна!\n")
	fmt.Printf("📊 Извлеченные данные:\n")
	fmt.Printf("   MAX ID: %d\n", userData.MaxID)
	fmt.Printf("   Имя: %s\n", userData.FirstName)
	fmt.Printf("   Фамилия: %s\n", userData.LastName)
	fmt.Printf("   Username: %s\n", userData.Username)
	fmt.Println()
	
	// Проверяем корректность данных
	fmt.Println("✅ Проверка корректности данных:")
	
	expectedMaxID := int64(18963527)
	if userData.MaxID == expectedMaxID {
		fmt.Printf("   ✅ MAX ID корректен: %d\n", userData.MaxID)
	} else {
		fmt.Printf("   ❌ MAX ID некорректен: ожидался %d, получен %d\n", expectedMaxID, userData.MaxID)
	}
	
	if userData.FirstName == firstName {
		fmt.Printf("   ✅ Имя корректно: %s\n", userData.FirstName)
	} else {
		fmt.Printf("   ❌ Имя некорректно: ожидалось %s, получено %s\n", firstName, userData.FirstName)
	}
	
	if userData.LastName == lastName {
		fmt.Printf("   ✅ Фамилия корректна: %s\n", userData.LastName)
	} else {
		fmt.Printf("   ❌ Фамилия некорректна: ожидалось %s, получено %s\n", lastName, userData.LastName)
	}
	
	if userData.Username == username {
		fmt.Printf("   ✅ Username корректен: %s\n", userData.Username)
	} else {
		fmt.Printf("   ❌ Username некорректен: ожидался %s, получен %s\n", username, userData.Username)
	}
	
	fmt.Println()
	fmt.Println("🎉 Тест валидатора завершен успешно!")
}

func createValidInitData(maxID, firstName, lastName, username, botToken string) string {
	// Создаем данные пользователя в JSON формате
	userJSON := fmt.Sprintf(`{"id":%s,"first_name":"%s","last_name":"%s","username":"%s","language_code":"ru","photo_url":"https://example.com/photo.jpg"}`, 
		maxID, firstName, lastName, username)
	
	// Создаем параметры
	authDate := fmt.Sprintf("%d", time.Now().Unix())
	chatJSON := `{"id":123809879,"type":"DIALOG"}`
	ip := "127.0.0.1"
	queryID := "test-query-id-12345"
	
	// Создаем параметры для подписи (сортированные по алфавиту)
	params := map[string]string{
		"auth_date": authDate,
		"chat":      chatJSON,
		"ip":        ip,
		"query_id":  queryID,
		"user":      userJSON,
	}
	
	// Сортируем ключи
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	// Создаем строку для подписи
	var signatureParams []string
	for _, key := range keys {
		signatureParams = append(signatureParams, fmt.Sprintf("%s=%s", key, params[key]))
	}
	dataCheckString := strings.Join(signatureParams, "\n")
	
	// Вычисляем подпись
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))
	
	// Создаем полную строку параметров
	values := url.Values{}
	values.Set("auth_date", authDate)
	values.Set("hash", hash)
	values.Set("chat", chatJSON)
	values.Set("ip", ip)
	values.Set("user", userJSON)
	values.Set("query_id", queryID)
	
	// Добавляем WebApp параметры
	fullParams := values.Encode() + "&WebAppPlatform=web&WebAppVersion=25.12.13"
	
	// URL кодируем всю строку
	return url.QueryEscape(fullParams)
}