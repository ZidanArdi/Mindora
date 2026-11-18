package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken   = errors.New("invalid or malformed token")
	ErrExpiredToken   = errors.New("token has expired")
	ErrInvalidClaims  = errors.New("token claims validation failed")
	ErrInvalidIssuer  = errors.New("invalid token issuer")
	ErrInvalidAudience = errors.New("invalid token audience")
)

// Claims defines the standard Mindora JWT access token payload.
type Claims struct {
	UserID uuid.UUID `json:"sub"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager handles JWT generation, validation, and refresh token hashing.
type TokenManager struct {
	secretKey     []byte
	issuer        string
	audience      string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewTokenManager creates a new instance of TokenManager.
func NewTokenManager(secretKey string, accessExpiry, refreshExpiry time.Duration) *TokenManager {
	return &TokenManager{
		secretKey:     []byte(secretKey),
		issuer:        "mindora-api",
		audience:      "mindora-client",
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken signs a new short-lived JWT access token for the given user.
func (m *TokenManager) GenerateAccessToken(userID uuid.UUID, email, role string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.accessExpiry)

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateAccessToken parses and validates a JWT access token string.
func (m *TokenManager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok || !parsedToken.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// GenerateRefreshToken generates a cryptographically random raw refresh token
// and returns its raw form and its SHA-256 hash.
func (m *TokenManager) GenerateRefreshToken() (string, string, time.Time, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	rawToken := hex.EncodeToString(randomBytes)
	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().UTC().Add(m.refreshExpiry)

	return rawToken, tokenHash, expiresAt, nil
}

// HashToken produces a deterministic SHA-256 hex string of the raw token.
func HashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

// AccessExpiry returns the access token duration.
func (m *TokenManager) AccessExpiry() time.Duration {
	return m.accessExpiry
}

// RefreshExpiry returns the refresh token duration.
func (m *TokenManager) RefreshExpiry() time.Duration {
	return m.refreshExpiry
}
