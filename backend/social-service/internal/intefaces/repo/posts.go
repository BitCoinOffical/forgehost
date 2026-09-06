package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostsRepo struct {
	pool *pgxpool.Pool
}

func NewPostsRepo(pool *pgxpool.Pool) *PostsRepo {
	return &PostsRepo{pool: pool}
}

func (r *PostsRepo) GetPostById(ctx context.Context, id string) (*models.FeedPost, error) {
	sql := `SELECT 
		p.username,
		p.avatar_url,
		ps.id AS post_id,
		ps.topic_id,
		ps.user_id,
		ps.image_url,
		ps.description,
		ps.views,
	(SELECT title FROM topics WHERE p.topic_id = id) AS topic_title, 
	(SELECT COUNT(*) FROM post_likes WHERE post_id = ps.id) AS like_count
	FROM profiles p LEFT JOIN posts ps ON p.user_id = ps.user_id`
	var post models.FeedPost
	if err := r.pool.QueryRow(ctx, sql, id).Scan(
		&post.Username,
		&post.AvatarURL,
		&post.PostID,
		&post.TopicID,
		&post.UserID,
		&post.ImageURL,
		&post.Description,
		&post.Views,
		&post.TopicName,
		&post.LikeCount,
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
(SELECT title FROM topics WHERE ps.topic_id = id) AS topic_title, 
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
			&fd.TopicName,
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
(SELECT title FROM topics WHERE ps.topic_id = id) AS topic_title, 
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
			&fd.TopicName,
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
			ps.id AS post_id,
			ps.topic_id,
			ps.user_id, 
			ps.image_url, 
			ps.description, 
			ps.views,
			(SELECT COUNT(*) FROM post_likes WHERE post_id = ps.id) AS like_count,
			(SELECT title FROM topics WHERE p.topic_id = id) AS topic_title,
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
			&fd.TopicName,
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

func (r *PostsRepo) GetTopics(ctx context.Context) ([]models.Topics, error) {
	sql := `SELECT * FROM topics WHERE is_delete = false`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	var tpcs []models.Topics
	for rows.Next() {
		var tpc models.Topics
		if err := rows.Scan(&tpc.ID, &tpc.Title, &tpc.IsDeleted, &tpc.CreatedAt, &tpc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		tpcs = append(tpcs, tpc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return tpcs, nil
}

func (r *PostsRepo) GetTopicsByID(ctx context.Context) (*models.Topics, error) {
	sql := `SELECT * FROM topics WHERE is_delete = false AND id = $1`
	var tpc models.Topics
	if err := r.pool.QueryRow(ctx, sql).Scan(&tpc.ID, &tpc.Title, &tpc.IsDeleted, &tpc.CreatedAt, &tpc.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &tpc, nil
}

func (r *PostsRepo) CreatePost(ctx context.Context, post *models.Post) error {
	sql := `INSERT INTO posts (topic_id, user_id, image_url, description) VALUES $1, $2, $3, $4`
	_, err := r.pool.Exec(ctx, sql, post.TopicId, post.UserID, post.ImageURL, post.Description)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return fmt.Errorf("post already exists: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *PostsRepo) BuildUpdatePost(ctx context.Context, post *models.Post) (*models.Post, error) {
	builder := squirrel.Update("posts").
		Where(squirrel.Eq{"id": post.ID}, squirrel.Eq{"user_id": post.UserID}).
		PlaceholderFormat(squirrel.Dollar).
		Suffix("RETURNING topic_id, user_id, image_url, description")

	if post.ImageURL != nil {
		builder.Set("image_url", post.ImageURL)
	}

	if post.Description != nil {
		builder.Set("description", post.Description)
	}

	builder.Set("updated_at", squirrel.Expr("NOW()"))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builder.ToSql: %w", err)
	}

	var resp models.Post
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&post.TopicId, &post.UserID, &post.ImageURL, &post.Description); err != nil {
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &resp, nil

}

func (r *PostsRepo) DeletePost(ctx context.Context, postId string, userId string) error {
	sql := `UPDATE posts SET is_delete = true WHERE id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, sql, postId, userId)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostsRepo) ReportPost(ctx context.Context, report *models.PostReport) error {
	sql := `INSERT INTO post_reports (user_id, post_id, cause) VALUES $1, $2, $3`
	_, err := r.pool.Exec(ctx, sql, report.UserId, report.PostId, report.Cause)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return fmt.Errorf("report already exists: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("r.pool.Exec: %w", err)
	}

	return nil
}

func (r *PostsRepo) LikePost(ctx context.Context, userId, postId string) error {
	sql := `INSER INTO post_likes (user_id, post_id) VALUES ($1, $2)`
	_, err := r.pool.Exec(ctx, sql, userId, postId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation {
			return fmt.Errorf("like already exists: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
func (r *PostsRepo) UnlikePost(ctx context.Context, userId, postId string) error {
	sql := `DELETE FROM post_likes WHERE user_id = $1 AND post_id = $2`
	_, err := r.pool.Exec(ctx, sql, userId, postId)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
