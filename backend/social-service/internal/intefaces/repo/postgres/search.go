package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchRepo struct {
	pool *pgxpool.Pool
}

func NewSearchRepo(pool *pgxpool.Pool) *SearchRepo {
	return &SearchRepo{pool: pool}
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (r *SearchRepo) SearchPosts(ctx context.Context, q string, limit int) ([]models.SearchPost, error) {
	sql := `SELECT id, user_id, description, image_id
	        FROM posts
	        WHERE is_delete = FALSE AND description ILIKE $1
	        ORDER BY created_at DESC
	        LIMIT $2`

	rows, err := r.pool.Query(ctx, sql, "%"+escapeLike(q)+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	posts := make([]models.SearchPost, 0)

	for rows.Next() {
		var p models.SearchPost

		if err := rows.Scan(&p.ID, &p.UserID, &p.Description, &p.ImageID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return posts, nil
}

func (r *SearchRepo) SearchStreams(ctx context.Context, q string, limit int) ([]models.SearchStream, error) {
	sql := `SELECT id, user_id, title
	        FROM streams
	        WHERE title ILIKE $1
	        ORDER BY title
	        LIMIT $2`

	rows, err := r.pool.Query(ctx, sql, "%"+escapeLike(q)+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	streams := make([]models.SearchStream, 0)

	for rows.Next() {
		var s models.SearchStream

		if err := rows.Scan(&s.ID, &s.UserID, &s.Title); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		streams = append(streams, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return streams, nil
}

func (r *SearchRepo) SearchUsers(ctx context.Context, q string, limit int) ([]models.SearchUser, error) {
	sql := `SELECT user_id, username, avatar_url
        FROM profiles
        WHERE username ILIKE $1
        AND is_banned = false
        ORDER BY username
        LIMIT $2`

	rows, err := r.pool.Query(ctx, sql, "%"+escapeLike(q)+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	users := make([]models.SearchUser, 0)

	for rows.Next() {
		var u models.SearchUser

		if err := rows.Scan(&u.UserID, &u.Username, &u.AvatarURL); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return users, nil
}

func (r *SearchRepo) SuggestUsers(ctx context.Context, userID uuid.UUID, prefix string, limit int) ([]models.Suggestion, error) {
	sql := `SELECT p.user_id, p.username
	        FROM profiles p
	        WHERE lower(p.username) LIKE lower($1) || '%'
	        AND p.user_id <> $3
	        AND NOT EXISTS (
	        SELECT 1
	        FROM blocks b
	        WHERE (b.user_id = $3 AND b.target_id = p.user_id)
	    OR (b.user_id = p.user_id AND b.target_id = $3)
	        )
	        ORDER BY p.username
	        LIMIT $2`

	rows, err := r.pool.Query(ctx, sql, escapeLike(prefix), limit, userID)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	suggestions := make([]models.Suggestion, 0)

	for rows.Next() {
		s := models.Suggestion{
			Type: "user",
		}

		if err := rows.Scan(&s.ID, &s.Text); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		suggestions = append(suggestions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return suggestions, nil
}

func (r *SearchRepo) SuggestStreams(ctx context.Context, prefix string, limit int) ([]models.Suggestion, error) {
	sql := `SELECT id, title
	        FROM streams
	        WHERE lower(title) LIKE lower($1) || '%'
	        ORDER BY title
	        LIMIT $2`

	rows, err := r.pool.Query(ctx, sql, escapeLike(prefix), limit)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	suggestions := make([]models.Suggestion, 0)

	for rows.Next() {
		s := models.Suggestion{
			Type: "stream",
		}

		if err := rows.Scan(&s.ID, &s.Text); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		suggestions = append(suggestions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return suggestions, nil
}
