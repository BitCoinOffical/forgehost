package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StreamRepo struct {
	pool *pgxpool.Pool
}

func NewStreamRepo(pool *pgxpool.Pool) *StreamRepo {
	return &StreamRepo{pool: pool}
}

func (r *StreamRepo) SaveChannel(ctx context.Context, channelID, userID string) error {
	sql := `UPDATE profiles SET channel_id = $1 WHERE user_id = $2`
	tag, err := r.pool.Exec(ctx, sql, channelID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return fmt.Errorf("channel already linked: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *StreamRepo) GetChannel(ctx context.Context, userID string) (string, error) {
	sql := `SELECT channel_id FROM profiles WHERE user_id = $1`
	var channelID *string
	if err := r.pool.QueryRow(ctx, sql, userID).Scan(&channelID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("r.pool.Exec: %w", err)
	}
	if channelID == nil {
		return "", domain.ErrChannelNotLinked
	}
	return *channelID, nil
}

func (r *StreamRepo) SaveStream(ctx context.Context, stream *models.Stream) (string, error) {
	sql := `INSERT INTO streams (user_id, video_id, title, status, embed_url, viewers, likes) 
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var id string
	err := r.pool.QueryRow(ctx, sql,
		&stream.UserID,
		&stream.VideoID,
		&stream.Title,
		&stream.Status,
		&stream.EmbedURL,
		&stream.Viewers,
		&stream.Likes).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return "", fmt.Errorf("stream exists: %w", domain.ErrAlreadyExists)
		}
		return "", fmt.Errorf("r.pool.QueryRow: %w", err)
	}
	return id, nil
}

func (r *StreamRepo) BuildUpdateStream(ctx context.Context, stream *models.Stream) (*models.Stream, error) {
	builder := squirrel.Update("streams").Where(squirrel.And{
		squirrel.Eq{"video_id": stream.VideoID},
		squirrel.Eq{"user_id": stream.UserID},
	}).Set("updated_at", squirrel.Expr("NOW()"))
	if stream.Title != nil {
		builder = builder.Set("title", stream.Title)
	}
	if stream.Status != nil {
		builder = builder.Set("status", stream.Status)
	}
	if stream.EmbedURL != nil {
		builder = builder.Set("embed_url", stream.EmbedURL)
	}
	if stream.Viewers != nil {
		builder = builder.Set("viewers", stream.Viewers)
	}
	if stream.Likes != nil {
		builder = builder.Set("likes", stream.Likes)
	}
	query, args, err := builder.Suffix("RETURNING id, user_id, video_id, title, status, embed_url, viewers, likes, created_at, updated_at").
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("builder.ToSql: %w", err)
	}

	var resp models.Stream
	if err := r.pool.QueryRow(ctx, query, args...).Scan(
		&resp.ID,
		&resp.UserID,
		&resp.VideoID,
		&resp.Title,
		&resp.Status,
		&resp.EmbedURL,
		&resp.Viewers,
		&resp.Likes,
		&resp.CreatedAt,
		&resp.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &resp, nil
}

func (r *StreamRepo) DeleteStream(ctx context.Context, videoID, userID string) error {
	sql := `DELETE FROM streams WHERE user_id = $1 AND video_id = $2`

	tag, err := r.pool.Exec(ctx, sql, userID, videoID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrStreamNotFound
	}
	return nil
}

func (r *StreamRepo) DeleteComplatedStreams(ctx context.Context) error {
	sql := `DELETE FROM streams WHERE status = 'complete'`

	_, err := r.pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *StreamRepo) GetStream(ctx context.Context, videoID, userID string) (*models.Stream, error) {
	sql := `SELECT id, user_id, video_id, title, status, embed_url, viewers, likes, created_at, updated_at
	FROM streams WHERE user_id = $1 AND video_id = $2`

	stream := &models.Stream{}
	err := r.pool.QueryRow(ctx, sql, userID, videoID).Scan(
		&stream.ID,
		&stream.UserID,
		&stream.VideoID,
		&stream.Title,
		&stream.Status,
		&stream.EmbedURL,
		&stream.Viewers,
		&stream.Likes,
		&stream.CreatedAt,
		&stream.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrStreamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}
	return stream, nil
}

func (r *StreamRepo) GetCursorStreams(ctx context.Context, streamID string, createdAt time.Time) ([]models.FeedStream, error) {
	sql := `
	SELECT 
	s.id, 
	p.username,
	p.avatar_url,
	s.user_id,
	s.video_id, 
	s.title, 
	s.status, 
	s.embed_url, 
	s.viewers, 
	s.likes, 
	s.created_at, 
	s.updated_at
	FROM streams s 
	LEFT JOIN profiles p ON s.user_id = p.user_id 
	WHERE ($1::timestamptz IS NULL AND $2::uuid IS NULL) OR (s.id, s.created_at) < ($1, $2) 
	ORDER BY s.created_at DESC, s.id DESC
	LIMIT 10
	`

	rows, err := r.pool.Query(ctx, sql, streamID, createdAt)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	var streams []models.FeedStream
	for rows.Next() {
		var stream models.FeedStream
		if err := rows.Scan(
			&stream.ID,
			&stream.UserName,
			&stream.AvatarURL,
			&stream.UserID,
			&stream.VideoID,
			&stream.Title,
			&stream.Status,
			&stream.EmbedURL,
			&stream.Viewers,
			&stream.Likes,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		streams = append(streams, stream)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return streams, nil
}
