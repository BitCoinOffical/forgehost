package dto

import "time"

type StreamDTO struct {
	VideoID           string  `json:"videoId" binding:"required"`
	Title             string  `json:"title" binding:"required,min=1,max=100"`
	Status            *string `json:"lifeCycleStatus" binding:"required"`
	EmbedURL          string  `json:"embedUrl" binding:"required,url,startswith=https://www.youtube.com/embed/"`
	ConcurrentViewers *uint64 `json:"concurrentViewers" binding:"omitnil"`
	LikeCount         *uint64 `json:"likeCount" binding:"omitnil"`
}

type UpdateStreamDTO struct {
	Title             *string `json:"title" binding:"omitnil,min=1,max=100"`
	Status            *string `json:"lifeCycleStatus" binding:"omitnil"`
	EmbedURL          *string `json:"embedUrl" binding:"omitnil,url,startswith=https://www.youtube.com/embed/"`
	ConcurrentViewers *uint64 `json:"concurrentViewers" binding:"omitnil"`
	LikeCount         *uint64 `json:"likeCount" binding:"omitnil"`
}

type CursorStreamDTO struct {
	StreamID  string    `json:"stream_id"`
	CreatedAt time.Time `json:"created_at"`
}
