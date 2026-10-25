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
	ImageUrl    string `json:"image_url" binding:"required_without=Description,omitempty,url,startswith=https://,max=2048"`
	Description string `json:"description" binding:"required_without=ImageUrl,omitempty,max=4096"`
	TopicID     *int   `json:"topic_id" binding:"omitnil,gt=0"`
}

type UpdatePostDTO struct {
	TopicID     *int    `json:"topic_id" binding:"omitnil,gt=0"`
	ImageUrl    *string `json:"image_url" binding:"omitnil,url,startswith=https://,max=2048"`
	Description *string `json:"description" binding:"omitnil,max=4096"`
}

type CreateTopicDTO struct {
	Title string `json:"title" binding:"required,min=1,max=100"`
}

type ReportTopicDTO struct {
	Cause string `json:"cause" binding:"required,min=1,max=255"`
}
