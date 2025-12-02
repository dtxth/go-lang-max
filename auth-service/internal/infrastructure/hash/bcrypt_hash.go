package hash

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher { return &BcryptHasher{} }

func (b *BcryptHasher) Hash(s string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
    return string(bytes), err
}

func (b *BcryptHasher) Compare(s, hashed string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(s)) == nil
}