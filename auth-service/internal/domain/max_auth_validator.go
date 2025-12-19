package domain

// MaxUserData represents user data extracted from MAX initData
type MaxUserData struct {
	MaxID     int64  `json:"max_id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
}

// MaxAuthValidator validates MAX Mini App initData
type MaxAuthValidator interface {
	ValidateInitData(initData string, botToken string) (*MaxUserData, error)
}