package repo

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerRepo struct {
	pool *pgxpool.Pool
}

func NewWorkerRepo(pool *pgxpool.Pool) *WorkerRepo {
	return &WorkerRepo{pool: pool}
}

func (r *WorkerRepo) GetCandidates(ctx context.Context) ([]models.FeedPost, error) {
	sql := `SELECT 
			p.username,
			p.avatar_url,
			ps.id, 
			ps.topic_id, 
			ps.user_id, 
			ps.image_url, 
			ps.description, 
			ps.views,
			(SELECT COUNT(*) FROM post_likes WHERE post_id = ps.id) AS like_count
			FROM posts ps JOIN profiles p ON ps.user_id = p.user_id
            ORDER BY ps.created_at DESC
			LIMIT 1000;
	`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	var fds []models.FeedPost
	for rows.Next() {
		var fd models.FeedPost
		if err := rows.Scan(
			&fd.Username,
			&fd.AvatarURL,
			&fd.ImageURL,
			&fd.Description,
			&fd.PostID,
			&fd.TopicID,
			&fd.UserID,
			&fd.Views,
			&fd.LikeCount,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		fds = append(fds, fd)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return fds, nil
}
