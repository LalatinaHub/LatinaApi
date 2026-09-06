package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/common/repository"
	"github.com/google/uuid"
)

type userRepo struct {
	repository.UserRepository
	db *sql.DB
}

// NewUserRepository returns an extended UserRepository backed by *sql.DB.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{
		UserRepository: repository.NewUserRepository(db),
		db:             db,
	}
}

func (r *userRepo) GetUserByToken(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, model.ErrInvalidToken
	}

	query := "SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users WHERE token = ? LIMIT 1;"
	row := r.db.QueryRowContext(ctx, query, token)

	return scanUser(row)
}

func (r *userRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	query := "SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users WHERE id = ? LIMIT 1;"
	row := r.db.QueryRowContext(ctx, query, id)

	return scanUser(row)
}

func (r *userRepo) CreateUserWithDefaults(ctx context.Context, id int64) (*model.User, error) {
	now := time.Now()
	u := &model.User{
		ID:         id,
		Token:      uuid.New().String(),
		Password:   uuid.New().String(), // UUIDv4 password
		Expired:    time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()),
		ServerCode: "",
		Quota:      1000, // 1000 MB
		Relay:      "",
		Adblock:    false,
		VPN:        "trojan",
	}

	_, err := r.CreateUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-provision user: %w", err)
	}

	return u, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner rowScanner) (*model.User, error) {
	var (
		u          model.User
		expiredStr string
		adblockVal any
	)

	err := scanner.Scan(
		&u.ID,
		&u.Token,
		&u.Password,
		&expiredStr,
		&u.ServerCode,
		&u.Quota,
		&u.Relay,
		&adblockVal,
		&u.VPN,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	switch v := adblockVal.(type) {
	case bool:
		u.Adblock = v
	case int64:
		u.Adblock = v > 0
	}

	if parsedExpired, err := time.Parse("2006-01-02", expiredStr); err == nil {
		u.Expired = parsedExpired
	} else if parsedExpired, err := time.Parse(time.RFC3339, expiredStr); err == nil {
		u.Expired = parsedExpired
	}

	return &u, nil
}
