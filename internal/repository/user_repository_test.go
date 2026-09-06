package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetUserByToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// 1. Success case
	rows := sqlmock.NewRows([]string{"id", "token", "password", "expired", "server_code", "quota", "relay", "adblock", "vpn"}).
		AddRow(1, "tok-123", "pass", "2026-12-31", "SG01", 1000, "", 1, "trojan")

	mock.ExpectQuery("SELECT (.+) FROM users WHERE token = \\? LIMIT 1").
		WithArgs("tok-123").
		WillReturnRows(rows)

	u, err := repo.GetUserByToken(ctx, "tok-123")
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, "tok-123", u.Token)
	assert.Equal(t, "SG01", u.ServerCode)
	assert.True(t, u.Adblock)

	// 2. Not found case
	mock.ExpectQuery("SELECT (.+) FROM users WHERE token = \\? LIMIT 1").
		WithArgs("non-existent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "token", "password", "expired", "server_code", "quota", "relay", "adblock", "vpn"}))

	u2, err := repo.GetUserByToken(ctx, "non-existent")
	assert.ErrorIs(t, err, model.ErrUserNotFound)
	assert.Nil(t, u2)

	// 3. Empty token
	u3, err := repo.GetUserByToken(ctx, "")
	assert.ErrorIs(t, err, model.ErrInvalidToken)
	assert.Nil(t, u3)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "token", "password", "expired", "server_code", "quota", "relay", "adblock", "vpn"}).
		AddRow(42, "tok-42", "pass", "2026-12-31", "SG01", 1000, "", false, "vless")

	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = \\? LIMIT 1").
		WithArgs(42).
		WillReturnRows(rows)

	u, err := repo.GetUserByID(ctx, 42)
	require.NoError(t, err)
	assert.Equal(t, int64(42), u.ID)
	assert.Equal(t, "vless", u.VPN)
	assert.False(t, u.Adblock)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CreateUserWithDefaults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "", int64(1000), "", 0, "trojan").
		WillReturnResult(sqlmock.NewResult(100, 1))

	u, err := repo.CreateUserWithDefaults(ctx, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(100), u.ID)
	assert.NotEmpty(t, u.Token)
	assert.NotEmpty(t, u.Password)
	assert.Equal(t, int64(1000), u.Quota)
	assert.WithinDuration(t, time.Now(), u.Expired, 25*time.Hour)

	assert.NoError(t, mock.ExpectationsWereMet())
}
