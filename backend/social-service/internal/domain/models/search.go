package models

import "github.com/google/uuid"

type SearchPost struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Description *string   `json:"description"`
	ImageID     *string
	ImageURL    *string `json:"image_url"`
}

type SearchStream struct {
	ID     string    `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Title  string    `json:"title"`
}

type SearchUser struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
}

type SearchResponse struct {
	Posts   []SearchPost   `json:"posts,omitempty"`
	Streams []SearchStream `json:"streams,omitempty"`
	Users   []SearchUser   `json:"users,omitempty"`
}

type Suggestion struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Text string `json:"text"`
}
