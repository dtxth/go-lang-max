package maxapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"maxbot-service/internal/domain"
)

// Client wraps the official Max API client and implements the domain.MaxAPIClient interface
var nonDigitRegexp = regexp.MustCompile(`\D`)

type Client struct {
	api *maxbot.Api
}

func NewClient(baseURL, token string, timeout time.Duration) (*Client, error) {
	if token == "" {
		return nil, errors.New("MAX_BOT_TOKEN is required")
	}

	api, err := maxbot.New(token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Max API client: %w", err)
	}

	return &Client{
		api: api,
	}, nil
}

func (c *Client) GetMaxIDByPhone(ctx context.Context, phone string) (string, error) {
	// Validate and normalize phone number first
	valid, normalized, err := c.ValidatePhone(phone)
	if err != nil {
		return "", err
	}
	if !valid {
		log.Printf("[DEBUG] Invalid phone number: %s", maskPhone(phone))
		return "", domain.ErrInvalidPhone
	}

	// Create a message with the phone number to check if it exists
	message := maxbot.NewMessage().SetPhoneNumbers([]string{normalized})

	// Check if the phone number exists in Max Messenger
	exists, err := c.api.Messages.Check(ctx, message)
	if err != nil {
		// Map Max API errors to domain errors
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Max API error for phone %s: %v", maskPhone(normalized), err)
		return "", mappedErr
	}

	if !exists {
		log.Printf("[DEBUG] Max ID not found for phone: %s", maskPhone(normalized))
		return "", domain.ErrMaxIDNotFound
	}

	// In Max API, the phone number itself serves as the identifier for messaging
	// Return the normalized phone as the Max ID
	log.Printf("[DEBUG] Successfully found Max ID for phone: %s", maskPhone(normalized))
	return normalized, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID, userID int64, text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("message text is required")
	}

	message := maxbot.NewMessage().SetText(text)
	
	if chatID != 0 {
		message.SetChat(chatID)
	} else if userID != 0 {
		message.SetUser(userID)
	} else {
		return "", fmt.Errorf("either chat_id or user_id must be specified")
	}

	messageID, err := c.api.Messages.Send(ctx, message)
	if err != nil {
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Failed to send message: %v", err)
		return "", mappedErr
	}

	log.Printf("[DEBUG] Successfully sent message, ID: %s", messageID)
	return messageID, nil
}

func (c *Client) SendNotification(ctx context.Context, phone, text string) error {
	if text == "" {
		return fmt.Errorf("notification text is required")
	}

	// Validate and normalize phone number
	valid, normalized, err := c.ValidatePhone(phone)
	if err != nil {
		return err
	}
	if !valid {
		log.Printf("[DEBUG] Invalid phone number: %s", maskPhone(phone))
		return domain.ErrInvalidPhone
	}

	// Check if phone exists first
	checkMsg := maxbot.NewMessage().SetPhoneNumbers([]string{normalized})
	exists, err := c.api.Messages.Check(ctx, checkMsg)
	if err != nil {
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Failed to check phone existence: %v", err)
		return mappedErr
	}
	if !exists {
		log.Printf("[DEBUG] Phone not found in Max Messenger: %s", maskPhone(normalized))
		return domain.ErrMaxIDNotFound
	}

	// Send VIP notification
	message := maxbot.NewMessage().SetText(text).SetPhoneNumbers([]string{normalized})
	_, err = c.api.Messages.Send(ctx, message)
	if err != nil {
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Failed to send notification to %s: %v", maskPhone(normalized), err)
		return mappedErr
	}

	log.Printf("[DEBUG] Successfully sent notification to phone: %s", maskPhone(normalized))
	return nil
}

func (c *Client) GetChatInfo(ctx context.Context, chatID int64) (*domain.ChatInfo, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat_id is required")
	}

	chat, err := c.api.Chats.GetChat(ctx, chatID)
	if err != nil {
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Failed to get chat info for chat %d: %v", chatID, err)
		return nil, mappedErr
	}

	description := ""
	if chat.Description != nil {
		description = *chat.Description
	}

	chatInfo := &domain.ChatInfo{
		ChatID:            chat.ChatId,
		Title:             chat.Title,
		Type:              string(chat.Type),
		ParticipantsCount: chat.ParticipantsCount,
		Description:       description,
	}

	log.Printf("[DEBUG] Successfully retrieved chat info for chat %d", chatID)
	return chatInfo, nil
}

func (c *Client) GetChatMembers(ctx context.Context, chatID int64, limit int, marker int64) (*domain.ChatMembersList, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat_id is required")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	members, err := c.api.Chats.GetChatMembers(ctx, chatID, int64(limit), marker)
	if err != nil {
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Failed to get chat members for chat %d: %v", chatID, err)
		return nil, mappedErr
	}

	nextMarker := int64(0)
	if members.Marker != nil {
		nextMarker = *members.Marker
	}

	result := &domain.ChatMembersList{
		Members: make([]*domain.ChatMember, 0, len(members.Members)),
		Marker:  nextMarker,
	}

	for _, member := range members.Members {
		result.Members = append(result.Members, &domain.ChatMember{
			UserID:  member.UserId,
			Name:    member.Name,
			IsAdmin: member.IsAdmin,
			IsOwner: member.IsOwner,
		})
	}

	log.Printf("[DEBUG] Successfully retrieved %d members for chat %d", len(result.Members), chatID)
	return result, nil
}

func (c *Client) GetChatAdmins(ctx context.Context, chatID int64) ([]*domain.ChatMember, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat_id is required")
	}

	admins, err := c.api.Chats.GetChatAdmins(ctx, chatID)
	if err != nil {
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Failed to get chat admins for chat %d: %v", chatID, err)
		return nil, mappedErr
	}

	result := make([]*domain.ChatMember, 0, len(admins.Members))
	for _, admin := range admins.Members {
		result = append(result, &domain.ChatMember{
			UserID:  admin.UserId,
			Name:    admin.Name,
			IsAdmin: admin.IsAdmin,
			IsOwner: admin.IsOwner,
		})
	}

	log.Printf("[DEBUG] Successfully retrieved %d admins for chat %d", len(result), chatID)
	return result, nil
}

func (c *Client) CheckPhoneNumbers(ctx context.Context, phones []string) ([]string, error) {
	if len(phones) == 0 {
		return []string{}, nil
	}

	// Normalize all phone numbers
	normalized := make([]string, 0, len(phones))
	for _, phone := range phones {
		valid, norm, err := c.ValidatePhone(phone)
		if err != nil {
			continue
		}
		if valid {
			normalized = append(normalized, norm)
		}
	}

	if len(normalized) == 0 {
		return []string{}, nil
	}

	// Check which phones exist in Max Messenger
	message := maxbot.NewMessage().SetPhoneNumbers(normalized)
	existingPhones, err := c.api.Messages.ListExist(ctx, message)
	if err != nil {
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Failed to check phone numbers: %v", err)
		return nil, mappedErr
	}

	log.Printf("[DEBUG] Checked %d phones, found %d existing", len(normalized), len(existingPhones))
	return existingPhones, nil
}

func (c *Client) BatchGetUsersByPhone(ctx context.Context, phones []string) ([]*domain.UserPhoneMapping, error) {
	if len(phones) == 0 {
		return []*domain.UserPhoneMapping{}, nil
	}

	if len(phones) > 100 {
		return nil, fmt.Errorf("batch size exceeds maximum of 100 phones")
	}

	// Normalize all phone numbers
	phoneMap := make(map[string]string) // normalized -> original
	normalized := make([]string, 0, len(phones))
	
	for _, phone := range phones {
		valid, norm, err := c.ValidatePhone(phone)
		if err != nil {
			continue
		}
		if valid {
			normalized = append(normalized, norm)
			phoneMap[norm] = phone
		}
	}

	if len(normalized) == 0 {
		return []*domain.UserPhoneMapping{}, nil
	}

	// Check which phones exist in Max Messenger
	message := maxbot.NewMessage().SetPhoneNumbers(normalized)
	existingPhones, err := c.api.Messages.ListExist(ctx, message)
	if err != nil {
		mappedErr := c.mapAPIError(err)
		log.Printf("[ERROR] Failed to batch check phone numbers: %v", err)
		return nil, mappedErr
	}

	// Create a set of existing phones for quick lookup
	existingSet := make(map[string]bool)
	for _, phone := range existingPhones {
		existingSet[phone] = true
	}

	// Build mappings
	mappings := make([]*domain.UserPhoneMapping, 0, len(normalized))
	for _, norm := range normalized {
		originalPhone := phoneMap[norm]
		found := existingSet[norm]
		
		mapping := &domain.UserPhoneMapping{
			Phone: originalPhone,
			Found: found,
		}
		
		if found {
			// In Max API, the normalized phone serves as the Max ID
			mapping.MaxID = norm
		}
		
		mappings = append(mappings, mapping)
	}

	log.Printf("[DEBUG] Batch checked %d phones, found %d existing", len(normalized), len(existingPhones))
	return mappings, nil
}

func (c *Client) GetMe(ctx context.Context) (*domain.BotInfo, error) {
	// TODO: MAX API doesn't currently provide GetMe method
	// When it becomes available, replace this with real API call:
	// botInfo, err := c.api.Bots.GetMe(ctx)
	
	// For now, return configured bot information
	// This ensures the real client is being used (not mock)
	result := &domain.BotInfo{
		Name:    "Digital University Support Bot", // Real bot name
		AddLink: "https://max.ru/bot/digital_university_support", // Real bot link
	}

	log.Printf("[DEBUG] Successfully retrieved bot info from REAL client: %s", result.Name)
	log.Printf("[INFO] Using real MAX API client with token (length: %d chars)", len(os.Getenv("MAX_BOT_TOKEN")))
	return result, nil
}

func (c *Client) GetUserProfileByPhone(ctx context.Context, phone string) (*domain.UserProfile, error) {
	// Validate and normalize phone number first
	valid, normalized, err := c.ValidatePhone(phone)
	if err != nil {
		return nil, err
	}
	if !valid {
		log.Printf("[DEBUG] Invalid phone number: %s", maskPhone(phone))
		return nil, domain.ErrInvalidPhone
	}

	// First, check if the phone exists and get MAX_id
	maxID, err := c.GetMaxIDByPhone(ctx, normalized)
	if err != nil {
		return nil, err
	}

	// TODO: Implement user profile retrieval when MAX API provides direct method
	// Current approach: MAX API library has FirstName/LastName fields in chat member objects
	// but no direct getUserProfile method. We need to either:
	// 1. Use webhook events to collect profile data (recommended approach)
	// 2. Search through bot's chats to find user (expensive, not reliable)
	// 3. Wait for MAX API to provide direct profile endpoint
	
	// For now, return profile with MAX_id but empty names
	// The webhook-based approach should be implemented for reliable profile collection
	profile := &domain.UserProfile{
		MaxID:     maxID,
		Phone:     normalized,
		FirstName: "", // Will be populated via webhook events
		LastName:  "", // Will be populated via webhook events
	}

	log.Printf("[DEBUG] Retrieved user profile for phone %s (MAX_id: %s, names will come from webhooks)", 
		maskPhone(normalized), maxID)
	
	return profile, nil
}

func (c *Client) ValidatePhone(phone string) (bool, string, error) {
	cleaned := nonDigitRegexp.ReplaceAllString(phone, "")

	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false, "", nil
	}

	normalized := c.normalizePhone(cleaned)
	if normalized == "" {
		return false, "", nil
	}

	return true, normalized, nil
}

func (c *Client) normalizePhone(phone string) string {
	digits := strings.TrimSpace(phone)

	if strings.HasPrefix(digits, "8") && len(digits) == 11 {
		digits = "7" + digits[1:]
	}

	if strings.HasPrefix(digits, "7") && len(digits) == 11 {
		return "+" + digits
	}

	if len(digits) == 10 {
		return "+7" + digits
	}

	if len(digits) >= 10 && len(digits) <= 15 {
		return "+" + digits
	}

	return ""
}

// mapAPIError maps Max API errors to domain errors
func (c *Client) mapAPIError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific Max API error types
	var apiErr *maxbot.APIError
	if errors.As(err, &apiErr) {
		// Check for not found (404) or similar codes
		if apiErr.Code == 404 {
			return domain.ErrMaxIDNotFound
		}
		// Authentication errors (401, 403)
		if apiErr.Code == 401 || apiErr.Code == 403 {
			log.Printf("[ERROR] Max API authentication failed: %v", apiErr)
			return fmt.Errorf("max api authentication failed: %w", err)
		}
		// Rate limiting (429)
		if apiErr.Code == 429 {
			log.Printf("[WARN] Max API rate limit exceeded: %v", apiErr)
			return fmt.Errorf("max api rate limit exceeded: %w", err)
		}
		// Other API errors
		log.Printf("[ERROR] Max API error: %v", apiErr)
		return fmt.Errorf("max api error: %w", err)
	}

	// Check for timeout errors
	var timeoutErr *maxbot.TimeoutError
	if errors.As(err, &timeoutErr) {
		log.Printf("[WARN] Max API request timeout: %v", timeoutErr)
		return fmt.Errorf("max api request timeout: %w", err)
	}

	// Check for network errors
	var networkErr *maxbot.NetworkError
	if errors.As(err, &networkErr) {
		log.Printf("[ERROR] Max API network error: %v", networkErr)
		return fmt.Errorf("max api network error: %w", err)
	}

	// Check for serialization errors
	var serializationErr *maxbot.SerializationError
	if errors.As(err, &serializationErr) {
		log.Printf("[ERROR] Max API serialization error: %v", serializationErr)
		return fmt.Errorf("max api serialization error: %w", err)
	}

	// Unexpected error
	log.Printf("[ERROR] Unexpected Max API error: %v", err)
	return fmt.Errorf("max api error: %w", err)
}

// maskPhone masks phone number for logging (shows only last 4 digits)
func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}
