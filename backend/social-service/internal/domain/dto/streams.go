package dto

import "time"

type StreamDTO struct {
	VideoID           string  `json:"videoId"`
	Title             string  `json:"title"`
	Status            *string `json:"lifeCycleStatus"`
	EmbedURL          string  `json:"embedUrl"`
	ConcurrentViewers *uint64 `json:"concurrentViewers"`
	LikeCount         *uint64 `json:"likeCount"`
}

type UpdateStreamDTO struct {
	Title             *string `json:"title"`
	Status            *string `json:"lifeCycleStatus"`
	EmbedURL          *string `json:"embedUrl"`
	ConcurrentViewers *uint64 `json:"concurrentViewers"`
	LikeCount         *uint64 `json:"likeCount"`
}

type CursorStreamDTO struct {
	StreamID  string    `json:"stream_id"`
	CreatedAt time.Time `json:"created_at"`
}
