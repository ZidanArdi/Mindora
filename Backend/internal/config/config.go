package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// DatabaseConfig holds all database connection and pool settings.
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DSN returns the Data Source Name string for PostgreSQL connection.
// It prioritizes URL (DATABASE_URL) if provided; otherwise builds from discrete fields.
func (d *DatabaseConfig) DSN() string {
	if d.URL != "" {
		return d.URL
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// JWTConfig holds configuration for access and refresh tokens.
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// GoogleOAuthConfig holds Google OAuth 2.0 / OpenID Connect credentials.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Enabled      bool
}

// TurnstileConfig holds Cloudflare Turnstile bot protection parameters.
type TurnstileConfig struct {
	Enabled          bool
	SecretKey        string
	ExpectedHostname string
	ExpectedAction   string
	VerifyURL        string
	AllowTestTokens  bool
}

// CookieConfig holds HTTP cookie parameters for refresh tokens.
type CookieConfig struct {
	Domain   string
	Secure   bool
	SameSite string
}

// ProxyConfig holds reverse proxy trust settings for client IP resolution and rate limiting.
type ProxyConfig struct {
	EnableTrustedProxyCheck bool
	TrustedProxies          []string
}

// GroqConfig holds LLM integration parameters for Groq Cloud API.
type GroqConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout time.Duration
}

// Config encapsulates entire backend configuration.
type Config struct {
	Port        string
	Env         string
	FrontendURL string
	BcryptCost  int
	Database    DatabaseConfig
	JWT         JWTConfig
	GoogleOAuth GoogleOAuthConfig
	Turnstile   TurnstileConfig
	Cookie      CookieConfig
	Proxy       ProxyConfig
	Groq        GroqConfig
}

// LoadConfig loads environment variables from .env and populates the Config struct.
func LoadConfig() *Config {
	// Traverse up directory tree to locate .env file (supports tests running from subdirectories)
	// Also check ./Backend/.env when started from the monorepo root.
	if currDir, err := os.Getwd(); err == nil {
		loaded := false
		dir := currDir
		for i := 0; i < 5; i++ {
			envPath := filepath.Join(dir, ".env")
			if info, err := os.Stat(envPath); err == nil && !info.IsDir() {
				_ = godotenv.Load(envPath)
				loaded = true
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if !loaded {
			backendEnv := filepath.Join(currDir, "Backend", ".env")
			if info, err := os.Stat(backendEnv); err == nil && !info.IsDir() {
				_ = godotenv.Load(backendEnv)
			}
		}
	}

	port := getEnv("APP_PORT", "8080")
	env := getEnv("APP_ENV", "development")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")
	bcryptCost := getEnvAsInt("BCRYPT_COST", 12)

	// Database configurations
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "mindora_user")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := getEnv("DB_NAME", "mindora")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")
	dbURL := os.Getenv("DATABASE_URL")

	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetime := getEnvAsDuration("DB_CONN_MAX_LIFETIME", 15*time.Minute)
	connMaxIdleTime := getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute)

	// JWT configurations
	jwtSecret := getEnv("JWT_SECRET", "mindora-dev-jwt-super-secret-key-32chars-minimum")
	jwtAccessExpiry := getEnvAsDuration("JWT_ACCESS_EXPIRY", 15*time.Minute)
	jwtRefreshExpiry := getEnvAsDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour)

	// Google OAuth configurations
	googleClientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	googleClientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	googleRedirectURL := strings.TrimSpace(getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"))
	googleOAuthEnabled := getEnvAsBool("GOOGLE_OAUTH_ENABLED", googleClientID != "" && googleClientSecret != "")

	// Turnstile configurations
	turnstileSecretKey := os.Getenv("TURNSTILE_SECRET_KEY")
	turnstileEnabled := getEnvAsBool("TURNSTILE_ENABLED", turnstileSecretKey != "")
	turnstileHostname := getEnv("TURNSTILE_EXPECTED_HOSTNAME", "localhost")
	turnstileAction := getEnv("TURNSTILE_EXPECTED_ACTION", "")
	turnstileVerifyURL := getEnv("TURNSTILE_VERIFY_URL", "https://challenges.cloudflare.com/turnstile/v0/siteverify")
	turnstileAllowTestTokens := getEnvAsBool("TURNSTILE_ALLOW_TEST_TOKENS", false)
	if env == "production" {
		turnstileAllowTestTokens = false // Fail closed in production: test tokens are strictly prohibited
	}

	// Cookie configurations
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	cookieSecure := getEnvAsBool("COOKIE_SECURE", env == "production")
	cookieSameSite := getEnv("COOKIE_SAMESITE", "Lax")

	// Reverse proxy configurations
	trustedProxiesRaw := getEnv("TRUSTED_PROXIES", "127.0.0.1,::1")
	var trustedProxies []string
	if trustedProxiesRaw != "" {
		for _, p := range strings.Split(trustedProxiesRaw, ",") {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				trustedProxies = append(trustedProxies, trimmed)
			}
		}
	}
	enableTrustedProxy := getEnvAsBool("ENABLE_TRUSTED_PROXY_CHECK", len(trustedProxies) > 0)

	// Groq LLM configurations
	groqAPIKey := os.Getenv("GROQ_API_KEY")
	groqModel := getEnv("GROQ_MODEL", "openai/gpt-oss-20b")
	groqBaseURL := getEnv("GROQ_BASE_URL", "https://api.groq.com/openai/v1")
	groqTimeout := getEnvAsDuration("GROQ_TIMEOUT", 30*time.Second)

	return &Config{
		Port:        port,
		Env:         env,
		FrontendURL: frontendURL,
		BcryptCost:  bcryptCost,
		Database: DatabaseConfig{
			Host:            dbHost,
			Port:            dbPort,
			User:            dbUser,
			Password:        dbPassword,
			Name:            dbName,
			SSLMode:         dbSSLMode,
			URL:             dbURL,
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdleTime,
		},
		JWT: JWTConfig{
			Secret:        jwtSecret,
			AccessExpiry:  jwtAccessExpiry,
			RefreshExpiry: jwtRefreshExpiry,
		},
		GoogleOAuth: GoogleOAuthConfig{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  googleRedirectURL,
			Enabled:      googleOAuthEnabled,
		},
		Turnstile: TurnstileConfig{
			Enabled:          turnstileEnabled,
			SecretKey:        turnstileSecretKey,
			ExpectedHostname: turnstileHostname,
			ExpectedAction:   turnstileAction,
			VerifyURL:        turnstileVerifyURL,
			AllowTestTokens:  turnstileAllowTestTokens,
		},
		Cookie: CookieConfig{
			Domain:   cookieDomain,
			Secure:   cookieSecure,
			SameSite: cookieSameSite,
		},
		Proxy: ProxyConfig{
			EnableTrustedProxyCheck: enableTrustedProxy,
			TrustedProxies:          trustedProxies,
		},
		Groq: GroqConfig{
			APIKey:  groqAPIKey,
			Model:   groqModel,
			BaseURL: groqBaseURL,
			Timeout: groqTimeout,
		},
	}
}

// Validate enforces configuration invariants, preventing silent failure in production.
func (c *Config) Validate() error {
	if c.Env == "production" {
		if !c.Turnstile.Enabled || strings.TrimSpace(c.Turnstile.SecretKey) == "" {
			return fmt.Errorf("Cloudflare Turnstile bot protection is required in production environment; TURNSTILE_ENABLED must be true and TURNSTILE_SECRET_KEY must be provided")
		}
		if c.Turnstile.AllowTestTokens {
			return fmt.Errorf("Cloudflare Turnstile test token bypass is strictly prohibited in production")
		}
		if c.JWT.Secret == "mindora-dev-jwt-super-secret-key-32chars-minimum" {
			return fmt.Errorf("insecure default JWT_SECRET is prohibited in production")
		}
		if c.Database.Password == "" && c.Database.URL == "" {
			return fmt.Errorf("database password or DATABASE_URL is required in production")
		}
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
