package models

import (
	"time"

	"github.com/google/uuid"
)

type FeedPost struct {
	Username    *string   `db:"username"`
	TopicName   *string   `db:"topic_title"`
	AvatarURL   *string   `db:"avatar_url"`
	ImageURL    *string   `db:"image_url"`
	Description *string   `db:"description"`
	PostID      string    `db:"post_id"`
	TopicID     *int      `db:"topic_id"`
	UserID      uuid.UUID `db:"user_id"`
	Views       int       `db:"views"`
	LikeCount   int       `db:"like_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Post struct {
	ID          string    `db:"id"`
	TopicID     *int      `db:"topic_id"`
	UserID      string    `db:"user_id"`
	ImageURL    *string   `db:"image_url"`
	Description *string   `db:"description"`
	Views       int       `db:"views"`
	IsDelete    bool      `db:"is_delete"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Topics struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Title     string    `db:"title"`
	IsDeleted bool      `db:"is_delete"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type TopicsReport struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	TopicID   string    `db:"topic_id"`
	Cause     string    `db:"cause"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CommentReport struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	CommentID string `db:"comment_id"`
	Cause     string `db:"cause"`
}

type PostReport struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
	PostID string `db:"post_id"`
	Cause  string `db:"cause"`
}
