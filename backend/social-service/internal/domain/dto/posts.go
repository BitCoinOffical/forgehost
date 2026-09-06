package dto

import "time"

type CursorDTO struct {
	PostId    string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type PostDTO struct {
	PostId string `json:"id"`
}

type CreatePostDTO struct {
	TopicId     string `json:"topic_id"`
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
}

type UpdatePostDTO struct {
	PostId      string `json:"id"`
	TopicId     string `json:"topic_id"`
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
}
