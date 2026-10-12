package postgres

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
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
func (r *MessagesRepo) GetUsersFromChat(ctx context.Context, chatID string) ([]models.FeedChatMember, error) {
	sql := `SELECT 
	p.username, 
	cm.chat_id, 
	cm.user_id, 
	cm.role, 
	cm.joined_at FROM chat_members cm 
	LEFT JOIN profiles p ON p.user_id = cm.user_id WHERE chat_id = $1`
	var fcms []models.FeedChatMember
	rows, err := r.pool.Query(ctx, sql, chatID)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	for rows.Next() {
		var fcm models.FeedChatMember
		if err := rows.Scan(
			&fcm.UserName,
			&fcm.ChatID,
			&fcm.UserID,
			&fcm.Role,
			&fcm.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		fcms = append(fcms, fcm)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return fcms, nil
}
func (r *MessagesRepo) CheckBanUser(ctx context.Context, chatID, targetID string) (bool, error) {
	var banned bool
	sql := `
        SELECT EXISTS (
            SELECT 1 FROM chat_bans
            WHERE chat_id = $1 AND user_id = $2
        )`
	err := r.pool.QueryRow(ctx, sql, chatID, targetID).Scan(&banned)
	if err != nil {
		return false, fmt.Errorf("r.db.QueryRow: %w", err)
	}
	return banned, nil
}
