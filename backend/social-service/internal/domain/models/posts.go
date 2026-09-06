package models

import (
	"time"
)

type FeedPost struct {
	Username    string    `db:"username"`
	TopicName   string    `db:"topic_title"`
	AvatarURL   *string   `db:"avatar_url"`
	ImageURL    *string   `db:"image_url"`
	Description *string   `db:"description"`
	PostID      int       `db:"post_id"`
	TopicID     *int      `db:"topic_id"`
	UserID      int       `db:"user_id"`
	Views       int       `db:"views"`
	LikeCount   int       `db:"like_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Post struct {
	ID          string    `db:"id"`
	TopicId     string    `db:"topic_id"`
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
	Title     string    `db:"title"`
	IsDeleted bool      `db:"is_delete"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CommentReport struct {
	ID        string `db:"id"`
	UserId    string `db:"user_id"`
	CommentId string `db:"comment_id"`
	Cause     string `db:"cause"`
}

type PostReport struct {
	ID     string `db:"id"`
	UserId string `db:"user_id"`
	PostId string `db:"post_id"`
	Cause  string `db:"cause"`
}
