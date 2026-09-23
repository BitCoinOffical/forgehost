package services

import "context"

type RabbitQueue interface {
	AddEmailTaskQueue(ctx context.Context, body []byte) error
}
