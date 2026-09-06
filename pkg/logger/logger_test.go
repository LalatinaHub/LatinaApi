package logger

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogger(t *testing.T) {
	SetupLogger("debug", false)
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
	assert.NotNil(t, Logger())

	SetupLogger("warn", true)
	assert.Equal(t, zerolog.WarnLevel, zerolog.GlobalLevel())

	SetupLogger("unknown", false)
	assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())
}

func TestLoggerEvents(t *testing.T) {
	SetupLogger("debug", false)
	assert.NotNil(t, Info())
	assert.NotNil(t, Error())
	assert.NotNil(t, Debug())
	assert.NotNil(t, Warn())

	sub := WithContext(map[string]any{"req_id": "123"})
	assert.NotNil(t, sub)
}
