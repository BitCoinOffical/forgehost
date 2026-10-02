package models

import "time"

type Message struct {
	ChatID    string    `bson:"chat_id"`
	UserID    string    `bson:"user_id"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"created_at"`
}
