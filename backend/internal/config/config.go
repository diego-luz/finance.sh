package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	Env  string
	Port string

	DB  DBConfig
	JWT JWTConfig

	CORSOrigins  []string
	RateLimitRPM int

	SwaggerEnabled bool
	EncryptionKey  string // base64-encoded 32 bytes for AES-256-GCM
	FrontendURL    string
	RetentionDays  int
	TermsVersion   string
	// RegistrationOpen controls whether public self-service signup is allowed.
	// When false, POST /auth/register is rejected (registration_closed) unless it
	// is an invitation-accept flow. Defaults to true for development.
	RegistrationOpen bool
	// JobsInProcess controls whether the app process runs the background
	// scheduler (recurrence engine, bill notifications, retention purge) as a
	// goroutine. Defaults to true: the single app binary owns it. Set to false to
	// disable the scheduler entirely — e.g. running multiple app replicas where
	// only one should drive jobs.
	JobsInProcess bool
	// JobsIntervalSec is the scheduler tick interval (env WORKER_INTERVAL_SEC)
	// used when JobsInProcess is true; defaults to 3600 (1 hour).
	JobsIntervalSec int

	// Admin bootstrap: OPT-IN (default false). The secure default for first run is
	// the setup wizard — a fresh DB (no users) makes the SPA show a form where the
	// operator CREATES the first super-admin (no secret is shown). Set
	// BootstrapAdmin=true only for headless/automated deploys: then a fresh DB
	// auto-creates the super-admin + first org, with MustChangePassword=true and,
	// when AdminPassword is empty, a generated password logged once on boot.
	// AdminEmail/AdminOrgName name the bootstrap account/organization.
	BootstrapAdmin bool
	AdminEmail     string
	AdminPassword  string
	AdminOrgName   string

	Login   LoginConfig
	SMTP    SMTPConfig
	Storage StorageConfig
}

// StorageConfig holds the attachment-storage settings. Blobs live in Postgres
// BYTEA in Community Edition; only the per-file size cap is configurable.
type StorageConfig struct {
	MaxAttachmentMB int
}

// LoginConfig holds brute-force / lockout parameters.
type LoginConfig struct {
	MaxAttempts int
	LockoutMin  int
}

// SMTPConfig holds outbound email settings. When Host is empty the mailer only
// logs messages (no real delivery) — the safe development default.
type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// Load reads configuration from the environment. A .env file is loaded when
// present but is optional (e.g. inside Docker where vars are injected).
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Env:  getenv("APP_ENV", "development"),
		Port: getenv("APP_PORT", "8080"),
		DB: DBConfig{
			Host:     getenv("DB_HOST", "localhost"),
			Port:     getenv("DB_PORT", "5432"),
			User:     getenv("DB_USER", "finance-sh"),
			Password: getenv("DB_PASSWORD", "finance-sh"),
			Name:     getenv("DB_NAME", "finance-sh"),
			SSLMode:  getenv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			AccessSecret:  getenv("JWT_ACCESS_SECRET", "dev-access-secret-change-me"),
			RefreshSecret: getenv("JWT_REFRESH_SECRET", "dev-refresh-secret-change-me"),
			AccessTTL:     time.Duration(getenvInt("JWT_ACCESS_TTL_MIN", 15)) * time.Minute,
			RefreshTTL:    time.Duration(getenvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour,
		},
		CORSOrigins:  []string{getenv("CORS_ORIGINS", "http://localhost:5173")},
		RateLimitRPM: getenvInt("RATE_LIMIT_RPM", 120),

		SwaggerEnabled: getenvBool("SWAGGER_ENABLED", true),
		// Dev default is a fixed test key; pkg/crypto logs a WARNING when it is used.
		EncryptionKey:    getenv("ENCRYPTION_KEY", ""),
		FrontendURL:      getenv("FRONTEND_URL", "http://localhost:8090"),
		RetentionDays:    getenvInt("RETENTION_DAYS", 90),
		TermsVersion:     getenv("TERMS_VERSION", "1.0"),
		RegistrationOpen: getenvBool("REGISTRATION_OPEN", true),
		JobsInProcess:    getenvBool("JOBS_IN_PROCESS", true),
		JobsIntervalSec:  getenvInt("WORKER_INTERVAL_SEC", 3600),

		BootstrapAdmin: getenvBool("BOOTSTRAP_ADMIN", false),
		AdminEmail:     getenv("ADMIN_EMAIL", "admin@finance.sh"),
		AdminPassword:  getenv("ADMIN_PASSWORD", ""),
		AdminOrgName:   getenv("ADMIN_ORG_NAME", "Minha Organização"),

		Login: LoginConfig{
			MaxAttempts: getenvInt("LOGIN_MAX_ATTEMPTS", 5),
			LockoutMin:  getenvInt("LOGIN_LOCKOUT_MIN", 15),
		},
		SMTP: SMTPConfig{
			Host: getenv("SMTP_HOST", ""),
			Port: getenv("SMTP_PORT", "587"),
			User: getenv("SMTP_USER", ""),
			Pass: getenv("SMTP_PASS", ""),
			From: getenv("SMTP_FROM", "finance.sh <no-reply@finance.sh>"),
		},
		Storage: StorageConfig{
			MaxAttachmentMB: getenvInt("ATTACHMENT_MAX_MB", 10),
		},
	}
}

// DSN builds the PostgreSQL connection string for GORM.
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
