package models

import "time"

type Stream struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	VideoID  string  `json:"video_id"`
	Title    *string `json:"title"`
	Status   *string `json:"status"`
	EmbedURL *string `json:"embed_url"`
	Viewers  *uint64 `json:"viewers"`
	Likes    *uint64 `json:"likes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type FeedStream struct {
	ID        string  `json:"id"`
	UserName  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	UserID    string  `json:"user_id"`
	VideoID   string  `json:"video_id"`
	Title     string  `json:"title"`
	Status    string  `json:"status"`
	EmbedURL  string  `json:"embed_url"`
	Viewers   *uint64 `json:"viewers"`
	Likes     *uint64 `json:"likes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
