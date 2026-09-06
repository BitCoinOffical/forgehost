package models

import "time"

type Comments struct {
	ID        string    `db:"id"`
	PostID    string    `db:"post_id"`
	UserID    string    `db:"user_id"`
	ParentID  *string   `db:"parent_id"`
	Body      string    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
}
type FeedComments struct {
	ID            string    `db:"id"`
	PostID        string    `db:"post_id"`
	UserID        string    `db:"user_id"`
	ParentID      *string   `db:"parent_id"`
	Body          string    `db:"body"`
	Likes         int       `db:"likes"`
	CountComments int       `db:"count_comments"`
	CreatedAt     time.Time `db:"created_at"`
}
