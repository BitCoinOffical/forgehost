package dto

import "time"

type CursorDTO struct {
	PostId    string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type PostDTO struct {
	PostId string `json:"id"`
}
