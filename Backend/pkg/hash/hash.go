package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Hasher defines password hashing and verification contract.
type Hasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hash, password string) error
}

// BcryptHasher implements Hasher using bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher initializes a BcryptHasher with the specified cost factor.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// HashPassword creates a secure one-way salted bcrypt hash of the given plain password.
func (b *BcryptHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// ComparePassword compares a bcrypt hashed password with its possible plaintext equivalent.
func (b *BcryptHasher) ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
