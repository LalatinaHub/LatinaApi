package repository

import (
	"context"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/common/repository"
)

// UserRepository extends common/repository.UserRepository with API-specific lookups.
type UserRepository interface {
	repository.UserRepository

	// GetUserByToken fetches a user by their unique subscription token.
	GetUserByToken(ctx context.Context, token string) (*model.User, error)

	// GetUserByID fetches a user by their unique primary ID.
	GetUserByID(ctx context.Context, id int64) (*model.User, error)

	// CreateUserWithDefaults auto-provisions a new user with default parameters.
	CreateUserWithDefaults(ctx context.Context, id int64) (*model.User, error)
}

// ServerRepository extends common/repository.ServerRepository with code-based queries.
type ServerRepository interface {
	repository.ServerRepository

	// GetServerByCode fetches a single server by its unique code (e.g. "SG01").
	GetServerByCode(ctx context.Context, code string) (*model.Server, error)
}

// ProxyFilter defines query parameters for dynamic proxy node filtering.
type ProxyFilter struct {
	VPN         string
	CountryCode string
	Region      string
	Transport   string
	ConnMode    string
	TLS         *bool
	Include     string
	Exclude     string
	Limit       int
}

// ProxyRepository extends common/repository.ProxyRepository with dynamic filter queries.
type ProxyRepository interface {
	repository.ProxyRepository

	// GetProxiesByFilter fetches proxy nodes matching dynamic criteria.
	GetProxiesByFilter(ctx context.Context, filter ProxyFilter) ([]model.ProxyNode, error)
}

// KVRepository extends common/repository.KVRepository with single key lookups.
type KVRepository interface {
	repository.KVRepository

	// GetValueByKey retrieves a single string configuration value by key.
	GetValueByKey(ctx context.Context, key string) (string, error)
}
