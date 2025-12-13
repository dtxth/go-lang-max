package maxapi

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"maxbot-service/internal/domain"
)

// MockClient implements domain.MaxAPIClient for development/testing
type MockClient struct {
	phoneRegex *regexp.Regexp
}

func NewMockClient() *MockClient {
	return &MockClient{
		phoneRegex: regexp.MustCompile(`^\+7\d{10}$`),
	}
}

func (c *MockClient) GetMaxIDByPhone(ctx context.Context, phone string) (string, error) {
	// Validate and normalize phone number first
	valid, normalized, err := c.ValidatePhone(phone)
	if err != nil {
		return "", err
	}
	if !valid {
		log.Printf("[DEBUG] [MOCK] Invalid phone number: %s", maskPhone(phone))
		return "", domain.ErrInvalidPhone
	}

	// Mock: always return success for valid phones
	log.Printf("[DEBUG] [MOCK] Successfully found Max ID for phone: %s", maskPhone(normalized))
	return normalized, nil
}

func (c *MockClient) SendMessage(ctx context.Context, chatID, userID int64, text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("message text is required")
	}

	if chatID == 0 && userID == 0 {
		return "", fmt.Errorf("either chat_id or user_id must be specified")
	}

	// Mock: generate fake message ID
	messageID := fmt.Sprintf("mock_msg_%d", time.Now().Unix())
	log.Printf("[DEBUG] [MOCK] Successfully sent message, ID: %s", messageID)
	return messageID, nil
}

func (c *MockClient) SendNotification(ctx context.Context, phone, text string) error {
	if text == "" {
		return fmt.Errorf("notification text is required")
	}

	// Validate and normalize phone number
	valid, normalized, err := c.ValidatePhone(phone)
	if err != nil {
		return err
	}
	if !valid {
		log.Printf("[DEBUG] [MOCK] Invalid phone number: %s", maskPhone(phone))
		return domain.ErrInvalidPhone
	}

	// Mock: always succeed for valid phones
	log.Printf("[DEBUG] [MOCK] Successfully sent notification to phone: %s", maskPhone(normalized))
	return nil
}

func (c *MockClient) GetChatInfo(ctx context.Context, chatID int64) (*domain.ChatInfo, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat_id is required")
	}

	// Mock: return fake chat info
	chatInfo := &domain.ChatInfo{
		ChatID:            chatID,
		Title:             fmt.Sprintf("Mock Chat %d", chatID),
		Type:              "group",
		ParticipantsCount: 42,
		Description:       "Mock chat for development",
	}

	log.Printf("[DEBUG] [MOCK] Successfully retrieved chat info for chat %d", chatID)
	return chatInfo, nil
}

func (c *MockClient) GetChatMembers(ctx context.Context, chatID int64, limit int, marker int64) (*domain.ChatMembersList, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat_id is required")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Mock: return fake members
	members := make([]*domain.ChatMember, 0, limit)
	for i := 0; i < limit && i < 10; i++ { // Max 10 mock members
		members = append(members, &domain.ChatMember{
			UserID:  int64(1000 + i),
			Name:    fmt.Sprintf("Mock User %d", i+1),
			IsAdmin: i == 0, // First user is admin
			IsOwner: i == 0, // First user is owner
		})
	}

	result := &domain.ChatMembersList{
		Members: members,
		Marker:  0, // No pagination in mock
	}

	log.Printf("[DEBUG] [MOCK] Successfully retrieved %d members for chat %d", len(result.Members), chatID)
	return result, nil
}

func (c *MockClient) GetChatAdmins(ctx context.Context, chatID int64) ([]*domain.ChatMember, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat_id is required")
	}

	// Mock: return fake admins
	admins := []*domain.ChatMember{
		{
			UserID:  1000,
			Name:    "Mock Admin 1",
			IsAdmin: true,
			IsOwner: true,
		},
		{
			UserID:  1001,
			Name:    "Mock Admin 2",
			IsAdmin: true,
			IsOwner: false,
		},
	}

	log.Printf("[DEBUG] [MOCK] Successfully retrieved %d admins for chat %d", len(admins), chatID)
	return admins, nil
}

func (c *MockClient) CheckPhoneNumbers(ctx context.Context, phones []string) ([]string, error) {
	if len(phones) == 0 {
		return []string{}, nil
	}

	// Mock: return all valid phones as existing
	existing := make([]string, 0, len(phones))
	for _, phone := range phones {
		valid, normalized, err := c.ValidatePhone(phone)
		if err != nil {
			continue
		}
		if valid {
			existing = append(existing, normalized)
		}
	}

	log.Printf("[DEBUG] [MOCK] Checked %d phones, found %d existing", len(phones), len(existing))
	return existing, nil
}

func (c *MockClient) BatchGetUsersByPhone(ctx context.Context, phones []string) ([]*domain.UserPhoneMapping, error) {
	if len(phones) == 0 {
		return []*domain.UserPhoneMapping{}, nil
	}

	if len(phones) > 100 {
		return nil, fmt.Errorf("batch size exceeds maximum of 100 phones")
	}

	// Mock: return mappings for all valid phones
	mappings := make([]*domain.UserPhoneMapping, 0, len(phones))
	for _, phone := range phones {
		valid, normalized, err := c.ValidatePhone(phone)
		if err != nil {
			continue
		}

		mapping := &domain.UserPhoneMapping{
			Phone: phone,
			Found: valid,
		}

		if valid {
			mapping.MaxID = normalized
		}

		mappings = append(mappings, mapping)
	}

	log.Printf("[DEBUG] [MOCK] Batch checked %d phones, found %d existing", len(phones), len(mappings))
	return mappings, nil
}

func (c *MockClient) ValidatePhone(phone string) (bool, string, error) {
	// Remove all non-digit characters
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false, "", nil
	}

	normalized := c.normalizePhone(cleaned)
	if normalized == "" {
		return false, "", nil
	}

	return true, normalized, nil
}

func (c *MockClient) normalizePhone(phone string) string {
	digits := strings.TrimSpace(phone)

	// Convert 8XXXXXXXXXX to 7XXXXXXXXXX
	if strings.HasPrefix(digits, "8") && len(digits) == 11 {
		digits = "7" + digits[1:]
	}

	// Add + prefix for Russian numbers
	if strings.HasPrefix(digits, "7") && len(digits) == 11 {
		return "+" + digits
	}

	// Add +7 prefix for 10-digit numbers
	if len(digits) == 10 {
		return "+7" + digits
	}

	// For other international numbers, just add +
	if len(digits) >= 10 && len(digits) <= 15 {
		return "+" + digits
	}

	return ""
}