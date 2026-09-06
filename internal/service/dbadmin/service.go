package dbadmin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/LatinaApi/internal/repository"
)

// DBAdminService handles administrative SQL executions securely.
type DBAdminService interface {
	ExecSQL(ctx context.Context, apiToken string, queries []string) ([]int64, error)
}

type dbAdminService struct {
	db     *sql.DB
	kvRepo repository.KVRepository
}

// NewDBAdminService returns a new DBAdminService.
func NewDBAdminService(db *sql.DB, kvRepo repository.KVRepository) DBAdminService {
	return &dbAdminService{
		db:     db,
		kvRepo: kvRepo,
	}
}

func (s *dbAdminService) ExecSQL(ctx context.Context, apiToken string, queries []string) ([]int64, error) {
	// 1. Verify API Token
	configuredToken, err := s.kvRepo.GetValueByKey(ctx, "apiToken")
	if err != nil {
		return nil, fmt.Errorf("failed to verify API token: %w", err)
	}

	if configuredToken != "" && apiToken != configuredToken {
		return nil, model.ErrUnauthorized
	}

	if len(queries) == 0 {
		return nil, nil
	}

	// 2. Begin atomic transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start database transaction: %w", err)
	}
	defer tx.Rollback()

	var results []int64
	for i, q := range queries {
		trimmed := strings.TrimSpace(q)
		if trimmed == "" {
			continue
		}

		res, err := tx.ExecContext(ctx, trimmed)
		if err != nil {
			return nil, fmt.Errorf("query %d failed: %w", i+1, err)
		}

		rows, _ := res.RowsAffected()
		results = append(results, rows)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit database transaction: %w", err)
	}

	return results, nil
}
