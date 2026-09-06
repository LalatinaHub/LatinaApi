package model

import "errors"

var (
	// ErrUserNotFound is returned when a user account cannot be located in the database.
	ErrUserNotFound = errors.New("user not found")

	// ErrSubscriptionExpired is returned when a user's subscription date is in the past.
	ErrSubscriptionExpired = errors.New("subscription expired")

	// ErrQuotaExceeded is returned when a user has no remaining byte quota.
	ErrQuotaExceeded = errors.New("subscription quota exceeded")

	// ErrInvalidToken is returned when an authentication token is malformed or invalid.
	ErrInvalidToken = errors.New("invalid or missing token")

	// ErrServerNotFound is returned when a designated edge node server is not registered.
	ErrServerNotFound = errors.New("server not found")

	// ErrUnauthorized is returned when an admin or API token check fails.
	ErrUnauthorized = errors.New("unauthorized")
)
