package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadRepo struct {
	pool *pgxpool.Pool
}

func NewUploadRepo(pool *pgxpool.Pool) *UploadRepo {
	return &UploadRepo{pool: pool}
}

func (r *UploadRepo) CreateUpload(ctx context.Context, imageID string, userID string) error {
	sql := `INSERT INTO uploads (image_id, user_id) VALUES ($1, $2)`
	_, err := r.pool.Exec(ctx, sql, imageID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return fmt.Errorf("upload already exists: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
