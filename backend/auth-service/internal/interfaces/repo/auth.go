package repo

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	sql := `SELECT id, password_hash FROM users WHERE email = $1`

	err := r.pool.QueryRow(ctx, sql, email).Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &user, nil
}

func (r *AuthRepo) SaveUser(ctx context.Context, req *models.User) (uuid.UUID, error) {
	sql := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`

	var id uuid.UUID

	if err := r.pool.QueryRow(ctx, sql, req.Email, req.PasswordHash).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, fmt.Errorf("email alredy exists: %w", domain.ErrEmailAlreadyExists)
		}
		return uuid.Nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return id, nil
}
