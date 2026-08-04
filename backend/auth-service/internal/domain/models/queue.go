package models

import "time"

type RabbitQueue struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
	Email  string `json:"email"`

	TitleSubject string    `json:"subject"`
	TaskType     string    `json:"task_type"`
	DispatchDate time.Time `json:"dispatch_date"`
}
