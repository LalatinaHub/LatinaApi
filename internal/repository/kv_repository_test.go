package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKVRepository_GetValueByKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewKVRepository(db)
	ctx := context.Background()

	// 1. Found
	mock.ExpectQuery("SELECT value FROM kv WHERE key = \\? LIMIT 1").
		WithArgs("apiToken").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("secret-12345"))

	val, err := repo.GetValueByKey(ctx, "apiToken")
	require.NoError(t, err)
	assert.Equal(t, "secret-12345", val)

	// 2. Not found returns empty string without error
	mock.ExpectQuery("SELECT value FROM kv WHERE key = \\? LIMIT 1").
		WithArgs("nonExistent").
		WillReturnRows(sqlmock.NewRows([]string{"value"}))

	val2, err := repo.GetValueByKey(ctx, "nonExistent")
	require.NoError(t, err)
	assert.Empty(t, val2)

	assert.NoError(t, mock.ExpectationsWereMet())
}
