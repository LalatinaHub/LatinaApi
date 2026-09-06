package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Clearenv()

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.False(t, cfg.IsProduction())
	assert.Equal(t, ":8080", cfg.Address())
	assert.Equal(t, "LatinaHub", cfg.DefaultSubscriptionTitle)
	assert.Equal(t, 5, cfg.CacheTTLMinutes)
	assert.Equal(t, 20.0, cfg.RateLimitRPS)
	assert.Equal(t, 50, cfg.RateLimitBurst)
}

func TestLoadConfig_Overrides(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("APP_ENV", "production")
	os.Setenv("TURSO_DATABASE_URL", "libsql://test.turso.io")
	os.Setenv("TURSO_AUTH_TOKEN", "secret-token")
	os.Setenv("CACHE_TTL_MINUTES", "10")
	os.Setenv("RATE_LIMIT_RPS", "50.5")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("TURSO_DATABASE_URL")
		os.Unsetenv("TURSO_AUTH_TOKEN")
		os.Unsetenv("CACHE_TTL_MINUTES")
		os.Unsetenv("RATE_LIMIT_RPS")
	}()

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "production", cfg.AppEnv)
	assert.True(t, cfg.IsProduction())
	assert.Equal(t, ":9090", cfg.Address())
	assert.Equal(t, "libsql://test.turso.io", cfg.TursoDatabaseURL)
	assert.Equal(t, "secret-token", cfg.TursoAuthToken)
	assert.Equal(t, 10, cfg.CacheTTLMinutes)
	assert.Equal(t, 50.5, cfg.RateLimitRPS)
}
