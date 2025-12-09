package domain

type UserRepository interface {
    Create(user *User) error
    GetByPhone(phone string) (*User, error)
    GetByEmail(email string) (*User, error) // Для обратной совместимости
    GetByID(id int64) (*User, error)
}