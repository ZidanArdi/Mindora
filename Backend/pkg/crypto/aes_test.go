package crypto_test

import (
	"bytes"
	"testing"

	"mindora/backend/pkg/crypto"
)

func TestAESGCM_EncryptDecrypt(t *testing.T) {
	key := []byte("super-secret-mindora-key-32bytes")
	plaintext := []byte(`{"access_token":"token.123","refresh_token":"raw_refresh_token"}`)

	ciphertext, err := crypto.EncryptAESGCM(plaintext, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if ciphertext == "" {
		t.Fatal("expected non-empty ciphertext")
	}

	decrypted, err := crypto.DecryptAESGCM(ciphertext, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("expected plaintext %s, got %s", string(plaintext), string(decrypted))
	}
}

func TestAESGCM_TamperDetection(t *testing.T) {
	key := []byte("super-secret-mindora-key-32bytes")
	plaintext := []byte("sensitive-session-data")

	ciphertext, err := crypto.EncryptAESGCM(plaintext, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Tamper with key
	wrongKey := []byte("wrong-key-for-decryption-test!!")
	_, err = crypto.DecryptAESGCM(ciphertext, wrongKey)
	if err == nil {
		t.Fatal("expected decryption error with wrong key, got nil")
	}

	// Corrupted ciphertext
	corrupted := ciphertext[:len(ciphertext)-5] + "AAAAA"
	_, err = crypto.DecryptAESGCM(corrupted, key)
	if err == nil {
		t.Fatal("expected decryption error on corrupted ciphertext, got nil")
	}
}
