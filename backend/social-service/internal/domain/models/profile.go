package models

import (
	"time"
)

type Profile struct {
	UserID        string    `db:"user_id"`
	UserName      *string   `db:"username"`
	Bio           *string   `db:"bio"`
	AvatarUrl     *string   `db:"avatar_url"`
	Subscribers   int       `db:"subscribers"`
	Subscriptions int       `db:"subscriptions"`
	Posts         int       `db:"posts"`
	IsBanned      bool      `db:"is_banned"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type Subscriptions struct {
	UserId string `db:"user_id"`
}

type Subscribes struct {
	TargetId string `db:"target_id"`
}

type ProfileResponse struct {
	Profile Profile    `json:"profile"`
	Posts   []FeedPost `json:"posts"`
}
