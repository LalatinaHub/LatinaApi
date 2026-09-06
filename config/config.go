package config

import (
	"bytes"
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

// LoadConfig loads configuration from environment variables and an optional .env file.
func LoadConfig(filenames ...string) (*Config, error) {
	loadDotEnv(filenames...)

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

// loadDotEnv loads environment variables from .env file, cleanly stripping UTF-8 BOM if present.
func loadDotEnv(filenames ...string) {
	target := ".env"
	if len(filenames) > 0 && filenames[0] != "" {
		target = filenames[0]
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return
	}

	// Strip UTF-8 Byte Order Mark (0xEF, 0xBB, 0xBF) if present
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	envMap, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return
	}

	for k, v := range envMap {
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}
