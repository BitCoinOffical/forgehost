package dto

import "time"

type CursorDTO struct {
	PostId    *string    `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
}

type PostDTO struct {
	PostId string `json:"id"`
}

type CreatePostDTO struct {
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
	TopicId     *int   `json:"topic_id"`
}

type UpdatePostDTO struct {
	PostId      string `json:"id"`
	TopicId     *int   `json:"topic_id"`
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
}
