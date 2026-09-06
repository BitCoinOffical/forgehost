package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID    uuid.UUID `db:"user_id"`
	UserName  *string   `db:"username"`
	Bio       *string   `db:"bio"`
	AvatarUrl *string   `db:"avatar_url"`
	IsBanned  bool      `db:"is_banned"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Subscriptions struct {
	UserId string `db:"user_id"`
}

type Subscribes struct {
	TargetId string `db:"target_id"`
}
