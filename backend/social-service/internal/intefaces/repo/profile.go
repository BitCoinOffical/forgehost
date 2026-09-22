package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	unique_violation = "23505"
)

type ProfileRepo struct {
	pool *pgxpool.Pool
}

func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

func (r *ProfileRepo) GetProfileByID(ctx context.Context, id string) (*models.Profile, []models.FeedPost, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("r.pool.Begin: %w", err)
	}

	sql := `
		SELECT p.user_id, p.username, p.bio, p.avatar_url, p.is_banned, p.created_at, p.updated_at, 
		(
			SELECT COUNT(*) FROM subscriptions s 
			WHERE s.target_user_id = p.user_id
		) AS subscribers, 
			(
			SELECT COUNT(*) FROM subscriptions s 
			WHERE s.user_id = p.user_id
		) AS subscriptions,
			(
			SELECT COUNT(*) FROM posts 
			WHERE p.user_id = posts.user_id AND posts.is_delete = false
		) AS posts
			FROM profiles p WHERE p.user_id = $1 AND p.is_banned = false
    `

	var resp models.Profile

	if err := tx.QueryRow(ctx, sql, id).Scan(
		&resp.UserID,
		&resp.UserName,
		&resp.Bio,
		&resp.AvatarUrl,
		&resp.IsBanned,
		&resp.CreatedAt,
		&resp.UpdatedAt,
		&resp.Subscribers,
		&resp.Subscriptions,
		&resp.Posts,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	rows, err := tx.Query(ctx,
		`
		SELECT 
		p.username,
		p.avatar_url,
		ps.id AS post_id,
		ps.topic_id,
		(SELECT title FROM topics WHERE ps.topic_id = id) AS topic_title, 
		ps.user_id,
		ps.image_url,
		ps.description,
		ps.views,
		(SELECT COUNT(*) FROM post_likes WHERE post_id = ps.id) AS like_count
		FROM posts ps 
        JOIN profiles p ON p.user_id = $1 AND ps.is_delete = false
		ORDER BY ps.created_at DESC
	`, id)
	if err != nil {
		return nil, nil, fmt.Errorf("tx.Query: %w", err)
	}

	fds := make([]models.FeedPost, 0)
	for rows.Next() {
		var fd models.FeedPost
		if err := rows.Scan(
			&fd.Username,
			&fd.AvatarURL,
			&fd.PostID,
			&fd.TopicID,
			&fd.TopicName,
			&fd.UserID,
			&fd.ImageURL,
			&fd.Description,
			&fd.Views,
			&fd.LikeCount,
		); err != nil {
			return nil, nil, fmt.Errorf("rows.Scan: %w", err)
		}
		fds = append(fds, fd)
	}

	return &resp, fds, nil
}

func (r *ProfileRepo) SaveProfile(ctx context.Context, userId string) error {
	sql := `INSERT INTO profiles (user_id, created_at, updated_at) VALUES ($1, NOW(), NOW())`
	if _, err := r.pool.Exec(ctx, sql, userId); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return fmt.Errorf("profile alredy exists: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *ProfileRepo) BuildUpdateProfile(ctx context.Context, profile *models.Profile) (*models.Profile, error) {
	builder := squirrel.Update("profiles").Where(squirrel.Eq{"user_id": profile.UserID})

	if profile.UserName != nil {
		builder = builder.Set("username", *profile.UserName)
	}

	if profile.Bio != nil {
		builder = builder.Set("bio", *profile.Bio)
	}

	if profile.AvatarUrl != nil {
		builder = builder.Set("avatar_url", *profile.AvatarUrl)
	}

	builder = builder.Set("updated_at", squirrel.Expr("NOW()"))

	query, args, err := builder.
		PlaceholderFormat(squirrel.Dollar).
		Suffix(`
		RETURNING user_id, username, bio, avatar_url, is_banned, created_at, updated_at,
		(
			SELECT COUNT(*) FROM subscriptions s 
			WHERE s.target_user_id = profiles.user_id
		) AS subscribers, 
		(
			SELECT COUNT(*) FROM subscriptions s 
			WHERE s.user_id = profiles.user_id
		) AS subscriptions,
		(
			SELECT COUNT(*) FROM posts ps 
			WHERE ps.user_id = profiles.user_id AND ps.is_delete = false
		) AS posts
		`).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("builder.ToSql: %w", err)
	}

	var resp models.Profile

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&resp.UserID,
		&resp.UserName,
		&resp.Bio,
		&resp.AvatarUrl,
		&resp.IsBanned,
		&resp.CreatedAt,
		&resp.UpdatedAt,
		&resp.Subscribers,
		&resp.Subscriptions,
		&resp.Posts,
	)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return fmt.Errorf("subscribe alredy exists: %w", domain.ErrAlreadyExists)
		}
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
