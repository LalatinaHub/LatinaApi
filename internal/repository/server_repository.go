package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/common/repository"
)

type serverRepo struct {
	repository.ServerRepository
	db *sql.DB
}

// NewServerRepository returns an extended ServerRepository backed by *sql.DB.
func NewServerRepository(db *sql.DB) ServerRepository {
	return &serverRepo{
		ServerRepository: repository.NewServerRepository(db),
		db:               db,
	}
}

func (r *serverRepo) GetServerByCode(ctx context.Context, code string) (*model.Server, error) {
	if code == "" {
		return nil, model.ErrServerNotFound
	}

	query := "SELECT id, code, domain, ip, country, users_count, users_max FROM servers WHERE code = ? LIMIT 1;"
	row := r.db.QueryRowContext(ctx, query, code)

	var s model.Server
	err := row.Scan(
		&s.ID,
		&s.Code,
		&s.Domain,
		&s.IP,
		&s.Country,
		&s.UsersCount,
		&s.UsersMax,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrServerNotFound
		}
		return nil, fmt.Errorf("failed to query server %s: %w", code, err)
	}

	return &s, nil
}
