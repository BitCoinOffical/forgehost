package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostsRepo struct {
	pool *pgxpool.Pool
}

func NewPostsRepo(pool *pgxpool.Pool) *PostsRepo {
	return &PostsRepo{pool: pool}
}

func (r *PostsRepo) GetPostById(ctx context.Context, id string) (*models.Post, error) {
	sql := `SELECT * FROM posts WHERE id = $1 AND is_delete = false`
	var post models.Post
	if err := r.pool.QueryRow(ctx, sql, id).Scan(
		&post.ID,
		&post.UserID,
		&post.ImageURL,
		&post.Description,
		&post.Views,
		&post.IsDelete,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}
	return &post, nil
}

func (r *PostsRepo) ViewPost(ctx context.Context, id string) error {
	sql := `UPDATE posts SET views = views + 1 WHERE id = $1 AND is_delete = false`
	_, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *PostsRepo) GetSubPosts(ctx context.Context, id string) ([]models.FeedPost, []models.FeedPost, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("r.pool.Begin: %w", err)
	}
	sql := `
WITH target_users_ids AS (
  SELECT s.target_user_id
  FROM subscriptions s WHERE s.user_id = $1
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
JOIN posts ps ON ps.user_id = tu.target_user_id AND ps.is_delete = false AND ps.created_at >= NOW() - INTERVAL '7 days'
JOIN profiles p ON p.user_id = tu.target_user_id
ORDER BY ps.created_at DESC LIMIT 10;

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
  FROM subscriptions s WHERE s.user_id = $1
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
JOIN posts ps ON ps.topic_id = tt.target_topic_id AND ps.is_delete = false AND ps.created_at >= NOW() - INTERVAL '7 days'
JOIN profiles p ON ps.user_id = p.user_id
ORDER BY ps.created_at DESC LIMIT 10;
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

func (r *PostsRepo) GetGlobalPosts(ctx context.Context) ([]models.FeedPost, error) {
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
