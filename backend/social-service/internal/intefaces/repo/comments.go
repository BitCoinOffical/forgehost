package repo

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentsRepo struct {
	pool *pgxpool.Pool
}

func NewCommentsRepo(pool *pgxpool.Pool) *CommentsRepo {
	return &CommentsRepo{pool: pool}
}

func (r *CommentsRepo) ListComments(ctx context.Context, postId string) ([]models.FeedComments, error) {
	sql := `SELECT 
	c.*,
	(SELECT COUNT(*) FROM comment_likes WHERE post_id = c.id) AS likes,
	(SELECT COUNT(*) FROM comments WHERE parent_id = c.id) AS count_comments
	FROM comments c WHERE c.post_id = $1`
	rows, err := r.pool.Query(ctx, sql, postId)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	var coments []models.FeedComments
	for rows.Next() {
		var coment models.FeedComments
		if err := rows.Scan(&coment.PostID, &coment.UserID, &coment.ParentID, &coment.Body, &coment.Likes, &coment.CountComments); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		coments = append(coments, coment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return coments, nil
}

func (r *CommentsRepo) CreateComment(ctx context.Context, comment *models.Comments) error {
	columns := []string{"post_id", "user_id"}
	values := []any{comment.PostID, comment.UserID}

	if comment.ParentID != nil {
		columns = append(columns, "parent_id")
		values = append(values, *comment.ParentID)
	}

	columns = append(columns, "body")
	values = append(values, comment.Body)

	query, args, err := squirrel.Insert("comments").Columns(columns...).Values(values...).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("squirrel.Insert: %w", err)
	}

	if _, err := r.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}

	return nil
}

func (r *CommentsRepo) UpdateComment(ctx context.Context, comment *models.Comments) (*models.Comments, error) {
	sql := `UPDATE comments SET body = $1 WHERE user_id = $2 AND post_id = $3 AND comment_id = $4 
	RETURNING id, post_id, user_id, parent_id, body`
	var cmt models.Comments
	if err := r.pool.QueryRow(ctx, sql, comment.Body, comment.UserID, comment.PostID, comment.ID).Scan(
		&cmt.ID,
		&cmt.PostID,
		&cmt.UserID,
		&cmt.ParentID,
		&cmt.Body,
	); err != nil {
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}
	return &cmt, nil
}

func (r *CommentsRepo) DeleteComment(ctx context.Context, comment *models.Comments) error {
	sql := `DELETE FROM comments WHERE user_id = $1 AND post_id = $2 AND comment_id = $3`
	_, err := r.pool.Exec(ctx, sql, comment.UserID, comment.PostID, comment.ID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *CommentsRepo) ReportComment(ctx context.Context, comment *models.CommentReport) error {
	sql := `INSERT INTO comment_reports (user_id, comment_id, cause) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, sql, comment.UserId, comment.CommentId, comment.Cause)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}

func (r *CommentsRepo) LikeComment(ctx context.Context, userId, comentId string) error {
	sql := `INSERT INTO comment_likes (user_id, comment_id) VALUES ($1, $2)`
	_, err := r.pool.Exec(ctx, sql, userId, comentId)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
func (r *CommentsRepo) UnlikeComment(ctx context.Context, userId, comentId string) error {
	sql := `DELETE FROM comment_likes WHERE user_id = $1 AND comment_id = $2`
	_, err := r.pool.Exec(ctx, sql, userId, comentId)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
