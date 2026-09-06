package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	unique_violation = "23505"
)

type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User

	sql := `SELECT email, email_verified, password_hash FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, sql, id).Scan(&user.Email, &user.EmailVerified, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &user, nil
}

func (r *AuthRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	sql := `SELECT id, email_verified, password_hash, updated_at, created_at FROM users WHERE email = $1`

	err := r.pool.QueryRow(ctx, sql, email).Scan(&user.ID, &user.EmailVerified, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &user, nil
}

func (r *AuthRepo) SaveUser(ctx context.Context, req *models.User) (uuid.UUID, error) {
	sql := `INSERT INTO users (email, password_hash, email_verified) VALUES ($1, $2, true) RETURNING id`

	var id uuid.UUID

	if err := r.pool.QueryRow(ctx, sql, req.Email, req.PasswordHash).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return uuid.Nil, fmt.Errorf("email alredy exists: %w", domain.ErrEmailAlreadyExists)
		}
		return uuid.Nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return id, nil
}

func (r *AuthRepo) SaveGoogleUser(ctx context.Context, req *models.User, oathReq *models.OAuthAccount) (uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("r.pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID uuid.UUID
	err = tx.QueryRow(ctx, "SELECT user_id FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2", oathReq.Provider, oathReq.ProviderUserID).Scan(&userID)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return uuid.Nil, fmt.Errorf("tx.Commit: %w", err)
		}
		return userID, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("tx.QueryRow: %w", err)
	}

	err = tx.QueryRow(ctx, `SELECT id FROM users WHERE email=$1`, req.Email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.QueryRow(ctx, `INSERT INTO users (name, email, picture, email_verified) VALUES ($1, $2, $3, $4) RETURNING id`, req.Name, req.Email, req.Picture, req.EmailVerified).Scan(&userID); err != nil {
			return uuid.Nil, fmt.Errorf("tx.QueryRow: %w", err)
		}
	} else if err != nil {
		return uuid.Nil, fmt.Errorf("check user by email: %w", err)
	} else {
		tag, err := tx.Exec(ctx, `UPDATE users SET email_verified=true WHERE id=$1 AND email_verified=false`, userID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("update email_verified: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return uuid.Nil, domain.ErrNotFound
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO oauth_accounts 
		(user_id, provider, provider_user_id, given_name, family_name)
		VALUES ($1, $2, $3, $4, $5)
		`, userID, oathReq.Provider, oathReq.ProviderUserID, oathReq.GivenName, oathReq.FamilyName); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, fmt.Errorf("oauth account already linked: %w", domain.ErrEmailAlreadyExists)
		}
		return uuid.Nil, fmt.Errorf("tx.Exec: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("tx.Commit: %w", err)
	}
	return userID, nil
}

func (r *AuthRepo) UpdateUserPassword(ctx context.Context, user *models.User) error {
	sql := `UPDATE users SET password_hash = $1 WHERE id = $2 AND email = $3`
	tag, err := r.pool.Exec(ctx, sql, user.PasswordHash, user.ID, user.Email)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
