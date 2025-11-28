package usecase

import (
	"chat-service/internal/domain"
	"testing"
)

// TestPaginationTotalCount tests that total count is included in response
func TestPaginationTotalCount(t *testing.T) {
	// Create 75 mock chats
	allChats := make([]*domain.Chat, 75)
	for i := 0; i < 75; i++ {
		allChats[i] = &domain.Chat{
			ID:   int64(i + 1),
			Name: "Chat " + string(rune('A'+(i%26))),
		}
	}

	repo := &MockChatRepository{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			total := len(allChats)
			
			// Apply pagination
			if offset >= total {
				return []*domain.Chat{}, total, nil
			}
			
			end := offset + limit
			if end > total {
				end = total
			}
			
			return allChats[offset:end], total, nil
		},
	}
	uc := NewListChatsWithRoleFilterUseCase(repo)

	filter := &domain.ChatFilter{
		Role:         "superadmin",
		UniversityID: nil,
	}

	// Get first page
	chats, total, err := uc.Execute("", 50, 0, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chats) != 50 {
		t.Errorf("expected 50 chats on first page, got %d", len(chats))
	}

	if total != 75 {
		t.Errorf("expected total count 75, got %d", total)
	}

	// Get second page
	chats, total, err = uc.Execute("", 50, 50, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chats) != 25 {
		t.Errorf("expected 25 chats on second page, got %d", len(chats))
	}

	if total != 75 {
		t.Errorf("expected total count 75 on second page, got %d", total)
	}
}

// TestPaginationOffsetExceedsTotal tests that empty array is returned when offset > total
func TestPaginationOffsetExceedsTotal(t *testing.T) {
	// Create 50 mock chats
	allChats := make([]*domain.Chat, 50)
	for i := 0; i < 50; i++ {
		allChats[i] = &domain.Chat{
			ID:   int64(i + 1),
			Name: "Chat " + string(rune('A'+(i%26))),
		}
	}

	repo := &MockChatRepository{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			total := len(allChats)
			
			// Apply pagination - return empty array when offset >= total
			if offset >= total {
				return []*domain.Chat{}, total, nil
			}
			
			end := offset + limit
			if end > total {
				end = total
			}
			
			return allChats[offset:end], total, nil
		},
	}
	uc := NewListChatsWithRoleFilterUseCase(repo)

	filter := &domain.ChatFilter{
		Role:         "superadmin",
		UniversityID: nil,
	}

	// Request with offset > total
	chats, total, err := uc.Execute("", 50, 100, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chats) != 0 {
		t.Errorf("expected empty array when offset > total, got %d chats", len(chats))
	}

	if total != 50 {
		t.Errorf("expected total count 50, got %d", total)
	}
}

// TestPaginationWithRoleFiltering tests pagination works with role-based filtering
func TestPaginationWithRoleFiltering(t *testing.T) {
	// Create chats for different universities
	allChats := make([]*domain.Chat, 0)
	universityID1 := int64(1)
	universityID2 := int64(2)

	// 60 chats for university 1
	for i := 0; i < 60; i++ {
		allChats = append(allChats, &domain.Chat{
			ID:           int64(i + 1),
			Name:         "Chat " + string(rune('A'+(i%26))),
			UniversityID: &universityID1,
		})
	}

	// 40 chats for university 2
	for i := 0; i < 40; i++ {
		allChats = append(allChats, &domain.Chat{
			ID:           int64(i + 61),
			Name:         "Chat " + string(rune('A'+(i%26))),
			UniversityID: &universityID2,
		})
	}

	repo := &MockChatRepository{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			// Apply role-based filtering
			filteredChats := make([]*domain.Chat, 0)
			for _, chat := range allChats {
				if filter != nil {
					if filter.IsSuperadmin() {
						// Superadmin sees all
						filteredChats = append(filteredChats, chat)
					} else if filter.IsCurator() && filter.UniversityID != nil {
						// Curator sees only their university
						if chat.UniversityID != nil && *chat.UniversityID == *filter.UniversityID {
							filteredChats = append(filteredChats, chat)
						}
					} else if filter.IsOperator() && filter.UniversityID != nil {
						// Operator sees only their university (simplified for test)
						if chat.UniversityID != nil && *chat.UniversityID == *filter.UniversityID {
							filteredChats = append(filteredChats, chat)
						}
					}
				}
			}

			total := len(filteredChats)

			// Apply pagination
			if offset >= total {
				return []*domain.Chat{}, total, nil
			}

			end := offset + limit
			if end > total {
				end = total
			}

			return filteredChats[offset:end], total, nil
		},
	}
	uc := NewListChatsWithRoleFilterUseCase(repo)

	// Curator for university 1 should see only 60 chats
	filter := &domain.ChatFilter{
		Role:         "curator",
		UniversityID: &universityID1,
	}

	chats, total, err := uc.Execute("", 50, 0, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chats) != 50 {
		t.Errorf("expected 50 chats on first page, got %d", len(chats))
	}

	if total != 60 {
		t.Errorf("expected total count 60 for curator's university, got %d", total)
	}

	// Second page should have 10 chats
	chats, total, err = uc.Execute("", 50, 50, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chats) != 10 {
		t.Errorf("expected 10 chats on second page, got %d", len(chats))
	}

	if total != 60 {
		t.Errorf("expected total count 60 on second page, got %d", total)
	}
}
