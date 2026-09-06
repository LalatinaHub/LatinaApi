package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config represents all application configuration parameters.
type Config struct {
	Port                     string
	AppEnv                   string
	TursoDatabaseURL         string
	TursoAuthToken           string
	DefaultSubscriptionTitle string
	CacheTTLMinutes          int
	RateLimitRPS             float64
	RateLimitBurst           int
}

// LoadConfig loads configuration from environment variables and .env file.
func LoadConfig() (*Config, error) {
	// Attempt loading .env file; ignore if missing
	_ = godotenv.Load()

	cfg := &Config{
		Port:                     getEnv("PORT", "8080"),
		AppEnv:                   getEnv("APP_ENV", "development"),
		TursoDatabaseURL:         os.Getenv("TURSO_DATABASE_URL"),
		TursoAuthToken:           os.Getenv("TURSO_AUTH_TOKEN"),
		DefaultSubscriptionTitle: getEnv("DEFAULT_SUBSCRIPTION_TITLE", "LatinaHub"),
		CacheTTLMinutes:          getEnvInt("CACHE_TTL_MINUTES", 5),
		RateLimitRPS:             getEnvFloat("RATE_LIMIT_RPS", 20.0),
		RateLimitBurst:           getEnvInt("RATE_LIMIT_BURST", 50),
	}

	return cfg, nil
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production" || c.AppEnv == "prod"
}

// Address returns the listen address (e.g. ":8080").
func (c *Config) Address() string {
	if c.Port == "" {
		return ":8080"
	}
	if c.Port[0] == ':' {
		return c.Port
	}
	return fmt.Sprintf(":%s", c.Port)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}
