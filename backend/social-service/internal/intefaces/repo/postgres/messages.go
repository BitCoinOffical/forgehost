package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MessagesRepo struct {
	pool *pgxpool.Pool
}

func NewMessagersRepo(pool *pgxpool.Pool) *MessagesRepo {
	return &MessagesRepo{pool: pool}
}

func (r *MessagesRepo) CheckBlock(ctx context.Context, userID, targetID string) (bool, error) {
	sql := `SELECT EXISTS (
	SELECT 1 FROM blocks 
	WHERE (user_id = $1 AND target_id = $2) 
	OR (user_id = $2 AND target_id = $1) 
	)`

	var blocked bool

	err := r.pool.QueryRow(ctx, sql, userID, targetID).Scan(&blocked)
	if err != nil {
		return false, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return blocked, nil
}
