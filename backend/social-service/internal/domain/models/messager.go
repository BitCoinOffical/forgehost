package models

import "time"

type Message struct {
	ChatID    string    `bson:"chat_id"`
	UserID    string    `bson:"user_id"`
	Text      string    `bson:"text"`
	CreatedAt time.Time `bson:"created_at"`
}
