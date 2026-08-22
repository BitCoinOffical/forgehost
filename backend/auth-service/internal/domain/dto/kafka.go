package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserKafka struct {
	UserId    uuid.UUID `json:"user_id"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
