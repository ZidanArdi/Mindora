package hash_test

import (
	"testing"
	"time"

	"mindora/backend/pkg/hash"
)

func TestBcryptHasher(t *testing.T) {
	hasher := hash.NewBcryptHasher(10)

	password := "SecurePassword123!"
	hashed, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}

	if hashed == password {
		t.Errorf("hashed password must not match plaintext")
	}

	// Verify correct password
	if err := hasher.ComparePassword(hashed, password); err != nil {
		t.Errorf("expected password match, got error: %v", err)
	}

	// Verify incorrect password
	if err := hasher.ComparePassword(hashed, "WrongPassword"); err == nil {
		t.Errorf("expected error for incorrect password, got nil")
	}
}

func BenchmarkBcryptCost(b *testing.B) {
	costs := []int{10, 11, 12}
	password := "BenchmarkPassword123!"

	for _, cost := range costs {
		hasher := hash.NewBcryptHasher(cost)
		start := time.Now()
		_, _ = hasher.HashPassword(password)
		duration := time.Since(start)
		b.Logf("Bcrypt cost %d took %v", cost, duration)
	}
}
