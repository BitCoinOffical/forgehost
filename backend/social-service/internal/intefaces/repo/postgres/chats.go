package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatsRepo struct {
	pool *pgxpool.Pool
}

func NewChatsRepo(pool *pgxpool.Pool) *ChatsRepo {
	return &ChatsRepo{pool: pool}
}

func (r *ChatsRepo) CreateChat(ctx context.Context, ownerID string, chat *models.Chat) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("r.pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var chatID string
	if err := tx.QueryRow(ctx, `INSERT INTO chats (title, description, owner_id) VALUES ($1, $2, $3) RETURNING id`, chat.Title, chat.Description, ownerID).Scan(&chatID); err != nil {
		return "", fmt.Errorf("tx.Exec: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO chat_members (chat_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, chatID, ownerID)
	if err != nil {
		return "", fmt.Errorf("tx.Exec insert member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("tx.Commit: %w", err)
	}

	return chatID, nil
}

func (r *ChatsRepo) GetChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	sql := `SELECT title, description, owner_id, created_at FROM chats WHERE id = $1`
	var chat models.Chat
	if err := r.pool.QueryRow(ctx, sql, chatID).Scan(
		&chat.Title,
		&chat.Description,
		&chat.OwnerID,
		&chat.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &chat, nil
}

func (r *ChatsRepo) DeleteChat(ctx context.Context, chatID, ownerID string) error {
	sql := `DELETE FROM chats WHERE id = $1 AND owner_id = $2`
	tag, err := r.pool.Exec(ctx, sql, chatID, ownerID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ChatsRepo) LeaveChat(ctx context.Context, chatID, userID string) error {
	sql := `DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, sql, chatID, userID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ChatsRepo) BuildUpdateChat(ctx context.Context, chatID, userID string, chat *models.Chat) (*models.Chat, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)
	var role string
	err = tx.QueryRow(ctx, `
	SELECT role 
	FROM chat_members 
	WHERE chat_id = $1 AND user_id = $2 
	FOR SHARE`, chatID, userID).Scan(&role)
	if err != nil {
		return nil, fmt.Errorf("tx.QueryRow: %w", err)
	}

	if role != "owner" {
		return nil, domain.ErrForbidden
	}

	builder := squirrel.Update("chats").
		Where(squirrel.Eq{"chat_id": chatID}, squirrel.Eq{"owner_id": userID}).
		Set("updated_at", squirrel.Expr("NOW()"))

	if chat.Title != nil {
		builder = builder.Set("title", chat.Title)
	}

	if chat.Description != nil {
		builder = builder.Set("description", chat.Description)
	}

	query, args, err := builder.Suffix("RETURNING id, title, description, owner_id, created_at, updated_at").
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("builder.ToSql: %w", err)
	}

	var resp models.Chat
	if err := tx.QueryRow(ctx, query, args...).Scan(
		&resp.ID,
		&resp.Title,
		&resp.Description,
		&resp.OwnerID,
		&resp.CreatedAt,
		&resp.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("tx.Commit: %w", err)
	}

	return &resp, nil

}

func (r *ChatsRepo) SetUserRoleChat(ctx context.Context, ownerID string, members *models.ChatMember) (*models.ChatMember, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)
	var role string
	err = tx.QueryRow(ctx, `
	SELECT role 
	FROM chat_members 
	WHERE chat_id = $1 AND user_id = $2 
	FOR SHARE`, members.ChatID, ownerID).Scan(&role)
	if err != nil {
		return nil, fmt.Errorf("tx.QueryRow: %w", err)
	}

	if role != "owner" {
		return nil, domain.ErrForbidden
	}

	var resp models.ChatMember
	if err := tx.QueryRow(ctx, `UPDATE chat_member 
	SET role = $1 
	WHERE chat_id = $2 AND user_id = $3 
	RETURNING chat_id, user_id, role, joined_at`, members.Role, members.ChatID, members.UserID).Scan(
		&resp.ChatID,
		&resp.UserID,
		&resp.Role,
		&resp.JoinedAt,
	); err != nil {
		return nil, fmt.Errorf("tx.QueryRow: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("tx.Commit: %w", err)
	}

	return &resp, nil
}

func (r *ChatsRepo) AddUserInChat(ctx context.Context, chatID, userID string) error {
	sql := `INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)`
	_, err := r.pool.Exec(ctx, sql, chatID, userID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec")
	}
	return nil
}

func (r *ChatsRepo) KickUserFromChat(ctx context.Context, chatID, targetID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("r.pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)
	var role string
	err = tx.QueryRow(ctx, `
	SELECT role 
	FROM chat_members 
	WHERE chat_id = $1 AND user_id = $2 
	FOR SHARE`, chatID, userID).Scan(&role)
	if err != nil {
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	if role != "moderator" && role != "owner" {
		return domain.ErrForbidden
	}

	var targetRole string
	err = tx.QueryRow(ctx, `SELECT role FROM chat_members WHERE chat_id = $1 AND user_id = $2`, chatID, targetID).Scan(&targetRole)
	if err != nil {
		return fmt.Errorf("tx.QueryRow target: %w", err)
	}
	if targetRole == "owner" {
		return domain.ErrCannotKickOwner
	}

	_, err = tx.Exec(ctx, `DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`, chatID, targetID)
	if err != nil {
		return fmt.Errorf("tx.Exec: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *ChatsRepo) GetUsersFromChat(ctx context.Context, chatID string) ([]models.FeedChatMember, error) {
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

func (r *ChatsRepo) BanUser(ctx context.Context, chatID, targetID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("r.pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)
	var role string
	err = tx.QueryRow(ctx, `
	SELECT role 
	FROM chat_members 
	WHERE chat_id = $1 AND user_id = $2 
	FOR SHARE`, chatID, userID).Scan(&role)
	if err != nil {
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	if role != "moderator" && role != "owner" {
		return domain.ErrForbidden
	}

	var targetRole string
	err = tx.QueryRow(ctx, `SELECT role FROM chat_members WHERE chat_id = $1 AND user_id = $2`, chatID, targetID).Scan(&targetRole)
	if err != nil {
		return fmt.Errorf("tx.QueryRow target: %w", err)
	}
	if targetRole == "owner" {
		return domain.ErrCannotKickOwner
	}

	_, err = r.pool.Exec(ctx, `INSERT INTO chat_bans (chat_id, user_id, banned_by) VALUES ($1, $2, $3)`, chatID, targetID, userID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}
func (r *ChatsRepo) UnbanUser(ctx context.Context, chatID, targetID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("r.pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)
	var role string
	err = tx.QueryRow(ctx, `
	SELECT role 
	FROM chat_members 
	WHERE chat_id = $1 AND user_id = $2 
	FOR SHARE`, chatID, userID).Scan(&role)
	if err != nil {
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	if role != "moderator" && role != "owner" {
		return domain.ErrForbidden
	}

	_, err = r.pool.Exec(ctx, `DELETE FROM chat_bans WHERE chat_id = $1 AND user_id = $2`, chatID, targetID, userID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *ChatsRepo) CheckBanUser(ctx context.Context, chatID, targetID string) (bool, error) {
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
