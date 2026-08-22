package dto

import (
	"github.com/google/uuid"
)

type UserProfileDTO struct {
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"username"`
	Bio       string    `json:"bio"`
	AvatarUrl string    `json:"avatar_url"`
}

type UpdateProfileDTO struct {
	UserName  string `json:"username"`
	Bio       string `json:"bio"`
	AvatarUrl string `json:"avatar_url"`
}
