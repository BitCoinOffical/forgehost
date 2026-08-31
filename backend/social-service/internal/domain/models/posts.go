package models

import (
	"time"

	"github.com/google/uuid"
)

type FeedPost struct {
	Username    string  `db:"username"`
	AvatarURL   *string `db:"avatar_url"`
	ImageURL    *string `db:"image_url"`
	Description *string `db:"description"`
	PostID      int     `db:"post_id"`
	TopicID     *int    `db:"topic_id"`
	UserID      int     `db:"user_id"`
	Views       int     `db:"views"`
	LikeCount   int     `db:"like_count"`
}

type Post struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	ImageURL    string    `db:"image_url"`
	Description string    `db:"description"`
	Views       int       `db:"views"`
	IsDelete    bool      `db:"is_delete"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
