package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerRepository_GetServerByCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewServerRepository(db)
	ctx := context.Background()

	// 1. Success
	rows := sqlmock.NewRows([]string{"id", "code", "domain", "ip", "country", "users_count", "users_max"}).
		AddRow(1, "SG01", "sg01.foolvpn.me", "1.2.3.4", "SG", 10, 100)

	mock.ExpectQuery("SELECT (.+) FROM servers WHERE code = \\? LIMIT 1").
		WithArgs("SG01").
		WillReturnRows(rows)

	s, err := repo.GetServerByCode(ctx, "SG01")
	require.NoError(t, err)
	assert.Equal(t, "SG01", s.Code)
	assert.Equal(t, "sg01.foolvpn.me", s.Domain)
	assert.Equal(t, "1.2.3.4", s.IP)

	// 2. Not found
	mock.ExpectQuery("SELECT (.+) FROM servers WHERE code = \\? LIMIT 1").
		WithArgs("NON_EXIST").
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "domain", "ip", "country", "users_count", "users_max"}))

	s2, err := repo.GetServerByCode(ctx, "NON_EXIST")
	assert.ErrorIs(t, err, model.ErrServerNotFound)
	assert.Nil(t, s2)

	// 3. Empty code
	s3, err := repo.GetServerByCode(ctx, "")
	assert.ErrorIs(t, err, model.ErrServerNotFound)
	assert.Nil(t, s3)

	assert.NoError(t, mock.ExpectationsWereMet())
}
