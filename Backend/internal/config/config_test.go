package config_test

import (
	"os"
	"testing"
	"time"

	"mindora/backend/internal/config"
)

func TestConfigLoadDefaults(t *testing.T) {
	oldPort := os.Getenv("APP_PORT")
	oldEnv := os.Getenv("APP_ENV")
	oldFrontend := os.Getenv("FRONTEND_URL")
	defer func() {
		if oldPort != "" {
			_ = os.Setenv("APP_PORT", oldPort)
		}
		if oldEnv != "" {
			_ = os.Setenv("APP_ENV", oldEnv)
		}
		if oldFrontend != "" {
			_ = os.Setenv("FRONTEND_URL", oldFrontend)
		}
	}()

	_ = os.Unsetenv("APP_PORT")
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("FRONTEND_URL")

	cfg := config.LoadConfig()

	if cfg.Port == "" {
		t.Errorf("expected default Port, got empty")
	}
	if cfg.Env == "" {
		t.Errorf("expected default Env, got empty")
	}
	if cfg.FrontendURL == "" {
		t.Errorf("expected default FrontendURL, got empty")
	}
	if cfg.Database.MaxOpenConns <= 0 {
		t.Errorf("expected positive MaxOpenConns, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.ConnMaxLifetime <= 0 {
		t.Errorf("expected positive ConnMaxLifetime, got %v", cfg.Database.ConnMaxLifetime)
	}
}

func TestDatabaseDSN(t *testing.T) {
	dbCfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5433",
		User:     "test_user",
		Password: "test_password",
		Name:     "test_db",
		SSLMode:  "disable",
	}

	dsn := dbCfg.DSN()
	expected := "host=localhost port=5433 user=test_user password=test_password dbname=test_db sslmode=disable"
	if dsn != expected {
		t.Errorf("expected DSN %q, got %q", expected, dsn)
	}

	// URL priority test
	dbCfg.URL = "postgres://custom:pass@remote:5432/customdb?sslmode=require"
	if dbCfg.DSN() != dbCfg.URL {
		t.Errorf("expected DSN to match URL %q, got %q", dbCfg.URL, dbCfg.DSN())
	}
}

func TestDurationHelper(t *testing.T) {
	oldVal := os.Getenv("DB_CONN_MAX_LIFETIME")
	defer func() {
		if oldVal != "" {
			_ = os.Setenv("DB_CONN_MAX_LIFETIME", oldVal)
		} else {
			_ = os.Unsetenv("DB_CONN_MAX_LIFETIME")
		}
	}()

	_ = os.Setenv("DB_CONN_MAX_LIFETIME", "30m")

	cfg := config.LoadConfig()
	if cfg.Database.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("expected ConnMaxLifetime 30m, got %v", cfg.Database.ConnMaxLifetime)
	}
}

func TestConfigValidation_ProductionSafety(t *testing.T) {
	// 1. Development configuration without Turnstile should pass validation
	devCfg := &config.Config{
		Env: "development",
	}
	if err := devCfg.Validate(); err != nil {
		t.Errorf("expected development config to pass validation, got: %v", err)
	}

	// 2. Production configuration without Turnstile should fail validation
	prodCfgNoTurnstile := &config.Config{
		Env: "production",
		Turnstile: config.TurnstileConfig{
			Enabled:   false,
			SecretKey: "",
		},
	}
	if err := prodCfgNoTurnstile.Validate(); err == nil {
		t.Errorf("SECURITY VULNERABILITY: production config passed without Turnstile enabled")
	}

	// 3. Production configuration with default JWT secret should fail validation
	prodCfgDefaultSecret := &config.Config{
		Env: "production",
		Turnstile: config.TurnstileConfig{
			Enabled:   true,
			SecretKey: "real-secret-key",
		},
		JWT: config.JWTConfig{
			Secret: "mindora-dev-jwt-super-secret-key-32chars-minimum",
		},
	}
	if err := prodCfgDefaultSecret.Validate(); err == nil {
		t.Errorf("SECURITY VULNERABILITY: production config passed with default JWT secret")
	}

	// 4. Production configuration with valid credentials should pass validation
	prodCfgValid := &config.Config{
		Env: "production",
		Turnstile: config.TurnstileConfig{
			Enabled:   true,
			SecretKey: "real-turnstile-secret",
		},
		JWT: config.JWTConfig{
			Secret: "production-high-entropy-jwt-secret-key-12345",
		},
		Database: config.DatabaseConfig{
			Password: "real-production-db-password",
		},
	}
	if err := prodCfgValid.Validate(); err != nil {
		t.Errorf("expected valid production config to pass, got: %v", err)
	}
}

func TestConfigProxySettings(t *testing.T) {
	cfg := config.LoadConfig()
	if !cfg.Proxy.EnableTrustedProxyCheck {
		t.Errorf("expected EnableTrustedProxyCheck to be enabled by default")
	}
	if len(cfg.Proxy.TrustedProxies) == 0 {
		t.Errorf("expected default TrustedProxies to be non-empty")
	}
}

func TestConfigGroqSettings(t *testing.T) {
	oldModel := os.Getenv("GROQ_MODEL")
	defer func() {
		if oldModel != "" {
			_ = os.Setenv("GROQ_MODEL", oldModel)
		} else {
			_ = os.Unsetenv("GROQ_MODEL")
		}
	}()

	_ = os.Setenv("GROQ_MODEL", "test-model-variant")
	cfg := config.LoadConfig()

	if cfg.Groq.Model != "test-model-variant" {
		t.Errorf("expected GROQ_MODEL 'test-model-variant', got %q", cfg.Groq.Model)
	}
	if cfg.Groq.BaseURL == "" {
		t.Errorf("expected default Groq.BaseURL, got empty")
	}
	if cfg.Groq.Timeout <= 0 {
		t.Errorf("expected positive Groq.Timeout, got %v", cfg.Groq.Timeout)
	}
}
