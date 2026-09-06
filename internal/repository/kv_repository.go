package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/LalatinaHub/common/repository"
)

type kvRepo struct {
	repository.KVRepository
	db *sql.DB
}

// NewKVRepository returns an extended KVRepository backed by *sql.DB.
func NewKVRepository(db *sql.DB) KVRepository {
	return &kvRepo{
		KVRepository: repository.NewKVRepository(db),
		db:           db,
	}
}

func (r *kvRepo) GetValueByKey(ctx context.Context, key string) (string, error) {
	query := "SELECT value FROM kv WHERE key = ? LIMIT 1;"
	row := r.db.QueryRowContext(ctx, query, key)

	var val string
	err := row.Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get kv key %s: %w", key, err)
	}

	return val, nil
}
