package repo

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostsRepo struct {
	pool *pgxpool.Pool
}

func NewPostsRepo(pool *pgxpool.Pool) *PostsRepo {
	return &PostsRepo{pool: pool}
}

func (r *PostsRepo) GetPostById(ctx context.Context, id string) error {
	sql := `SELECT * FROM posts WHERE id = $1`
	r.pool.Query(ctx, sql, id)
	return nil
}

func (r *PostsRepo) ViewPost(ctx context.Context, id string) error {
	sql := `UPDATE posts SET views = views + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *PostsRepo) GetPostsList(ctx context.Context, id string) ([]models.FeedPost, []models.FeedPost, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("r.pool.Begin: %w", err)
	}
	sql := `
WITH target_users_ids AS (
  SELECT s.target_user_id
  FROM subscriptions s WHERE s.user_id = 18
)

SELECT 
p.username,
p.avatar_url,
ps.id AS post_id,
ps.topic_id,
ps.user_id,
ps.image_url,
ps.description,
ps.views,
(SELECT COUNT(*) FROM post_likes WHERE post_id = ps.id) AS like_count
FROM target_users_ids tu 
JOIN posts ps ON ps.user_id = tu.target_user_id
JOIN profiles p ON p.user_id = tu.target_user_id
ORDER BY ps.created_at DESC LIMIT 100;

	`
	rows, err := tx.Query(ctx, sql, id)
	if err != nil {
		return nil, nil, fmt.Errorf("tx.Query: %w", err)
	}

	var sublists []models.FeedPost
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
			return nil, nil, fmt.Errorf("rows.Scan: %w", err)
		}
		sublists = append(sublists, fd)
	}

	sql = `
WITH target_topics_ids AS (
  SELECT s.target_topic_id
  FROM subscriptions s WHERE s.user_id = 18
)

SELECT 
p.username,
p.avatar_url,
ps.id AS post_id,
ps.topic_id,
ps.user_id,
ps.image_url,
ps.description,
ps.views
FROM target_topics_ids tt 
JOIN posts ps ON ps.topic_id = tt.target_topic_id
JOIN profiles p ON ps.user_id = p.user_id
ORDER BY ps.created_at DESC LIMIT 100;
	`
	rows, err = tx.Query(ctx, sql, id)
	if err != nil {
		return nil, nil, fmt.Errorf("tx.Query: %w", err)
	}
	var subtopics []models.FeedPost
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
			return nil, nil, fmt.Errorf("rows.Scan: %w", err)
		}
		subtopics = append(subtopics, fd)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("tx.Commit: %w", err)
	}
	return sublists, subtopics, nil
}
