package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port string

	// FrontendOrigin はCORSで許可するフロントエンドのオリジン。
	FrontendOrigin string

	DatabaseURL string

	JWTAccessSecret  string
	JWTRefreshSecret string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration

	GeminiAPIKey  string
	GeminiModel   string
	GeminiTimeout time.Duration

	AdminEmail string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	IMAPHost         string
	IMAPPort         int
	IMAPUsername     string
	IMAPPassword     string
	IMAPPollInterval time.Duration
	PasswordResetTTL time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		FrontendOrigin:   getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  14 * 24 * time.Hour,
		GeminiAPIKey:     os.Getenv("GEMINI_API_KEY"),
		GeminiModel:      getEnv("GEMINI_MODEL", "gemini-3.5-flash"),
		GeminiTimeout:    5 * time.Second,
		AdminEmail:       os.Getenv("ADMIN_EMAIL"),
		SMTPHost:         os.Getenv("SMTP_HOST"),
		SMTPUsername:     os.Getenv("SMTP_USERNAME"),
		SMTPPassword:     os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:         os.Getenv("SMTP_FROM"),
		IMAPHost:         os.Getenv("IMAP_HOST"),
		IMAPUsername:     os.Getenv("IMAP_USERNAME"),
		IMAPPassword:     os.Getenv("IMAP_PASSWORD"),
		PasswordResetTTL: 24 * time.Hour,
	}

	var err error
	if cfg.SMTPPort, err = getEnvInt("SMTP_PORT", 587); err != nil {
		return nil, err
	}
	if cfg.IMAPPort, err = getEnvInt("IMAP_PORT", 993); err != nil {
		return nil, err
	}
	pollSeconds, err := getEnvInt("IMAP_POLL_INTERVAL_SECONDS", 30)
	if err != nil {
		return nil, err
	}
	cfg.IMAPPollInterval = time.Duration(pollSeconds) * time.Second

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTAccessSecret == "" || cfg.JWTRefreshSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid value for %s: %w", key, err)
	}
	return n, nil
}
