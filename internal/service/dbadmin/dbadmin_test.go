package dbadmin

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockKVRepo struct {
	val string
	err error
}

func (m *mockKVRepo) GetAll(ctx context.Context) (map[string]any, error) {
	return nil, nil
}
func (m *mockKVRepo) GetValueByKey(ctx context.Context, key string) (string, error) {
	return m.val, m.err
}

func TestDBAdminService_ExecSQL_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	kvRepo := &mockKVRepo{val: "secret-token"}
	svc := NewDBAdminService(db, kvRepo)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET quota = 500 WHERE id = 1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM proxies WHERE id = 99").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	queries := []string{
		"UPDATE users SET quota = 500 WHERE id = 1",
		"DELETE FROM proxies WHERE id = 99",
	}

	results, err := svc.ExecSQL(context.Background(), "secret-token", queries)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 1}, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBAdminService_ExecSQL_Unauthorized(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	kvRepo := &mockKVRepo{val: "correct-token"}
	svc := NewDBAdminService(db, kvRepo)

	_, err = svc.ExecSQL(context.Background(), "wrong-token", []string{"SELECT 1"})
	assert.ErrorIs(t, err, model.ErrUnauthorized)
}
