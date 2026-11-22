package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrDecryptionFailed  = errors.New("decryption failed or invalid authentication tag")
	ErrInvalidCiphertext = errors.New("ciphertext payload is invalid or truncated")
)

// EncryptAESGCM encrypts plaintext using AES-256-GCM authenticated encryption.
// The key is deterministically derived to 32 bytes using SHA-256 if not already 32 bytes.
// A unique 12-byte nonce is prepended to the ciphertext.
func EncryptAESGCM(plaintext []byte, key []byte) (string, error) {
	if len(plaintext) == 0 {
		return "", errors.New("plaintext cannot be empty")
	}

	key32 := DeriveKey(key)

	block, err := aes.NewCipher(key32)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to initialize GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate secure nonce: %w", err)
	}

	// Seal appends the ciphertext and 16-byte auth tag to the nonce slice
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptAESGCM decrypts a base64-encoded AES-256-GCM ciphertext and validates its authentication tag.
func DecryptAESGCM(ciphertextB64 string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	key32 := DeriveKey(key)

	block, err := aes.NewCipher(key32)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// DeriveKey deterministically derives a 32-byte (256-bit) key from arbitrary input using SHA-256.
func DeriveKey(key []byte) []byte {
	if len(key) == 32 {
		return key
	}
	hash := sha256.Sum256(key)
	return hash[:]
}
