package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	SupabaseURL           string
	SupabaseAnonKey       string
	AddisPayAPIKey        string
	AddisPayBaseURL       string
	AddisPayWebhookSecret string
	AddisPayRedirectURL   string
	AddisPayCancelURL     string
	AddisPaySuccessURL    string
	AddisPayErrorURL      string
	RedisURL              string
	Environment           string
	AllowedOrigins        []string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		SupabaseURL:           getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:       getEnv("SUPABASE_ANON_KEY", ""),
		AddisPayAPIKey:        getEnv("ADDISPAY_API_KEY", ""),
		AddisPayBaseURL:       getEnv("ADDISPAY_BASE_URL", "https://uat.api.addispay.et"),
		AddisPayWebhookSecret: getEnv("ADDISPAY_WEBHOOK_SECRET", ""),
		AddisPayRedirectURL:   getEnv("ADDISPAY_REDIRECT_URL", "http://localhost:5173"),
		AddisPayCancelURL:     getEnv("ADDISPAY_CANCEL_URL", "http://localhost:5173/tournaments"),
		AddisPaySuccessURL:    getEnv("ADDISPAY_SUCCESS_URL", "http://localhost:5173/me/dashboard"),
		AddisPayErrorURL:      getEnv("ADDISPAY_ERROR_URL", "http://localhost:5173/me/dashboard"),
		RedisURL:              getEnv("REDIS_URL", ""),
		Environment:           getEnv("ENVIRONMENT", "development"),
		AllowedOrigins:        splitOrigins(getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitOrigins(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
