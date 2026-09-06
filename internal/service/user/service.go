package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/LatinaApi/internal/repository"
)

// UserService manages user retrieval and auto-provisioning.
type UserService interface {
	GetUser(ctx context.Context, apiToken string, id int64) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	kvRepo   repository.KVRepository
}

// NewUserService returns a new UserService.
func NewUserService(userRepo repository.UserRepository, kvRepo repository.KVRepository) UserService {
	return &userService{
		userRepo: userRepo,
		kvRepo:   kvRepo,
	}
}

func (s *userService) GetUser(ctx context.Context, apiToken string, id int64) (*model.User, error) {
	// Verify API token if configured in database
	configuredToken, err := s.kvRepo.GetValueByKey(ctx, "apiToken")
	if err != nil {
		return nil, fmt.Errorf("failed to verify API token: %w", err)
	}

	if configuredToken != "" && apiToken != configuredToken {
		return nil, model.ErrUnauthorized
	}

	// Fetch existing user
	u, err := s.userRepo.GetUserByID(ctx, id)
	if err == nil {
		return u, nil
	}

	if errors.Is(err, model.ErrUserNotFound) {
		// Auto-provision new account with defaults
		return s.userRepo.CreateUserWithDefaults(ctx, id)
	}

	return nil, err
}
