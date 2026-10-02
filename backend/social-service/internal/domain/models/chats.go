package models

import "time"

type MemberRole string

const (
	RoleMember MemberRole = "member"
	RoleAdmin  MemberRole = "moderator"
)

type Chat struct {
	ID          string  `db:"id"`
	Title       *string `db:"title"`
	Description *string `db:"description"`
	OwnerID     string  `db:"owner_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ChatMember struct {
	ChatID   string     `db:"chat_id"`
	UserID   string     `db:"user_id"`
	Role     MemberRole `db:"role"`
	JoinedAt time.Time  `db:"joined_at"`
}

type FeedChatMember struct {
	UserName *string    `db:"username"`
	ChatID   string     `db:"chat_id"`
	UserID   string     `db:"user_id"`
	Role     MemberRole `db:"role"`
	JoinedAt time.Time  `db:"joined_at"`
}
