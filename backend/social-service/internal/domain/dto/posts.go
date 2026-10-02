package dto

import "time"

type CursorDTO struct {
	PostID    *string    `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
}

type PostDTO struct {
	PostID string `json:"id"`
}

type CreatePostDTO struct {
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
	TopicID     *int   `json:"topic_id"`
}

type UpdatePostDTO struct {
	PostID      string `json:"id"`
	TopicID     *int   `json:"topic_id"`
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
}

type CreateTopicDTO struct {
	Title string `json:"title"`
}

type ReportTopicDTO struct {
	Cause string `json:"cause"`
}
