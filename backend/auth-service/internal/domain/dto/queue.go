package dto

import "time"

type RabbitQueueDTO struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
	Email  string `json:"email"`

	DispatchDate time.Time `json:"dispatch_date"`
}
