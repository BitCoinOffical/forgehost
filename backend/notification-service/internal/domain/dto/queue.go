package dto

import "time"

type RabbitQueueDTO struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
	Email  string `json:"email"`

	TitleSubject string    `json:"subject"`
	DispatchDate time.Time `json:"dispatch_date"`
}
