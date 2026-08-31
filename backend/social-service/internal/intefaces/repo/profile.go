package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepo struct {
	pool *pgxpool.Pool
}

func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

func (r *ProfileRepo) GetProfileByID(ctx context.Context, id string) (*models.Profile, error) {
	sql := `
		SELECT p.username, p.bio, p.avatar_url, p.is_banned, p.created_at, p.updated_at, 
		(
			SELECT COUNT(*) FROM subscriptions s 
			WHERE s.target_id = p.user_id
		) AS subscribers, 
			(
			SELECT COUNT(*) FROM subscriptions s 
			WHERE s.user_id = p.user_id
		) AS subscriptions,
			(
			SELECT COUNT(*) FROM posts 
			WHERE p.user_id = posts.user_id AND posts.is_delete = false
		) AS posts
			FROM profiles p WHERE user_id = $1 
    `

	var resp models.Profile

	if err := r.pool.QueryRow(ctx, sql, id).Scan(&resp.UserName, &resp.Bio, &resp.AvatarUrl, &resp.IsBanned, &resp.CreatedAt, &resp.UpdatedAt); err != nil {
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &resp, nil
}

func (r *ProfileRepo) SaveProfile(ctx context.Context, profile *models.Profile) error {
	sql := `INSERT INTO profiles (user_id, created_at, updated_at) VALUES ($1, NOW(), NOW())`
	if _, err := r.pool.Exec(ctx, sql, profile.UserID); err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *ProfileRepo) BuildUpdateProfile(ctx context.Context, profile *models.Profile) (*models.Profile, error) {
	builder := squirrel.Update("profiles").Where(squirrel.Eq{"id": profile.UserID}).PlaceholderFormat(squirrel.Dollar).
		Suffix("RETURNING user_id, username, bio, avatar_url, is_banned, created_at, updated_at")

	if profile.UserName != nil {
		builder.Set("username", profile.UserName)
	}

	if profile.Bio != nil {
		builder.Set("bio", profile.Bio)
	}

	if profile.AvatarUrl != nil {
		builder.Set("avatar_url", profile.AvatarUrl)
	}

	builder.Set("updated_at", squirrel.Expr("NOW()"))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builder.ToSql: %w", err)
	}

	var resp models.Profile

	err = r.pool.QueryRow(ctx, query, args...).Scan(&resp.UserID, &resp.UserName, &resp.Bio, &resp.AvatarUrl, &resp.IsBanned, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &resp, nil
}

func (r *ProfileRepo) GetSubscriptions(ctx context.Context, id string) ([]models.Subscriptions, error) {
	sql := `SELECT user_id FROM subscriptions WHERE target_id = 2`
	var subs []models.Subscriptions
	rows, err := r.pool.Query(ctx, sql, id)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	for rows.Next() {
		var sub models.Subscriptions
		if err := rows.Scan(&sub.UserId); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return subs, nil
}

func (r *ProfileRepo) GetSubscribers(ctx context.Context, id string) ([]models.Subscribes, error) {
	sql := `SELECT target_id FROM subscriptions WHERE user_id = 2`
	var subscr []models.Subscribes
	rows, err := r.pool.Query(ctx, sql, id)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	for rows.Next() {
		var subcr models.Subscribes
		if err := rows.Scan(&subcr.TargetId); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		subscr = append(subscr, subcr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return subscr, nil
}

func (r *ProfileRepo) Subscribe(ctx context.Context, userId, targetId string) error {
	sql := `INSERT INTO subscriptions (user_id, target_id, created_at) VALUES ($1, $2, NOW())`
	_, err := r.pool.Exec(ctx, sql, userId, targetId)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *ProfileRepo) UnSubscribe(ctx context.Context, userId, targetId string) error {
	sql := `DELETE FROM subscriptions WHERE user_id = $1 AND target_id = $2`
	_, err := r.pool.Exec(ctx, sql, userId, targetId)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *ProfileRepo) CreateProfileReport(ctx context.Context, userId, targetId, cause string) error {
	sql := `INSERT INTO profile_reports (user_id, target_id, cause, created_at, updated_at) VALUES $1, $2, $3, NOW(), NOW()`
	_, err := r.pool.Exec(ctx, sql, userId, targetId, cause)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
