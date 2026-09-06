package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReexportedModels(t *testing.T) {
	now := time.Now()

	activeUser := User{
		ID:         1,
		Token:      "valid-token",
		Quota:      100,
		Expired:    now.Add(24 * time.Hour),
		ServerCode: "SG01",
		VPN:        "trojan",
	}
	assert.True(t, activeUser.IsActive(now))

	expiredUser := User{
		ID:         2,
		Token:      "expired-token",
		Quota:      100,
		Expired:    now.Add(-24 * time.Hour),
		ServerCode: "SG01",
		VPN:        "trojan",
	}
	assert.False(t, expiredUser.IsActive(now))

	server := Server{
		Code:       "SG01",
		UsersCount: 100,
		UsersMax:   100,
	}
	assert.True(t, server.IsFull())

	serverNotFull := Server{
		Code:       "SG02",
		UsersCount: 50,
		UsersMax:   100,
	}
	assert.False(t, serverNotFull.IsFull())
}

func TestDomainErrors(t *testing.T) {
	assert.Equal(t, "user not found", ErrUserNotFound.Error())
	assert.Equal(t, "subscription expired", ErrSubscriptionExpired.Error())
	assert.Equal(t, "subscription quota exceeded", ErrQuotaExceeded.Error())
	assert.Equal(t, "invalid or missing token", ErrInvalidToken.Error())
	assert.Equal(t, "server not found", ErrServerNotFound.Error())
	assert.Equal(t, "unauthorized", ErrUnauthorized.Error())
}
