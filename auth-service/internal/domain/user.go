package domain

// Роли пользователей
const (
	RoleSuperAdmin = "super_admin" // Суперадмин (представитель VK): Полные права настройки сервиса
	RoleCurator    = "curator"     // Куратор (ответственный представитель от вуза): Имеет права на создание чатов, назначение операторов и управление чатами в масштабе всего вуза
	RoleOperator   = "operator"    // Оператор (представитель подразделения вуза, например, деканата): Управляет чатами в рамках своего подразделения
)

type User struct {
    ID       int64  `json:"id"`
    Phone    string `json:"phone"`    // Основной идентификатор (телефон)
    Email    string `json:"email,omitempty"` // Опциональный email
    Password string `json:"-"`
    Role     string `json:"role"`
}