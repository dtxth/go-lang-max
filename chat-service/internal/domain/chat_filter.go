package domain

// ChatFilter представляет фильтр для чатов на основе роли и контекста пользователя
type ChatFilter struct {
	Role         string
	UniversityID *int64
	BranchID     *int64
	FacultyID    *int64
}

// NewChatFilter создает новый фильтр чатов из информации о токене
func NewChatFilter(tokenInfo *TokenInfo) *ChatFilter {
	if tokenInfo == nil {
		return nil
	}

	return &ChatFilter{
		Role:         tokenInfo.Role,
		UniversityID: tokenInfo.UniversityID,
		BranchID:     tokenInfo.BranchID,
		FacultyID:    tokenInfo.FacultyID,
	}
}

// IsSuperadmin проверяет, является ли пользователь суперадмином
func (f *ChatFilter) IsSuperadmin() bool {
	return f.Role == "superadmin"
}

// IsCurator проверяет, является ли пользователь куратором
func (f *ChatFilter) IsCurator() bool {
	return f.Role == "curator"
}

// IsOperator проверяет, является ли пользователь оператором
func (f *ChatFilter) IsOperator() bool {
	return f.Role == "operator"
}
