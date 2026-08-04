package rabbitqueue

import (
	rabbitmq "BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	attemptTTL  = 10000 //10sec
	maxAttempts = 5
	codeQueue   = "code_queue"
	resetQueue  = "reset_queue"
)

type RabbitQueue struct {
	rc *rabbitmq.ResilientConnection
}

func NewQueue(rc *rabbitmq.ResilientConnection) *RabbitQueue {
	return &RabbitQueue{
		rc: rc,
	}
}

func (r *RabbitQueue) AddVerificationCodeQueue(ctx context.Context, body []byte) error {
	ch, err := r.rc.NewChannel()
	if err != nil {
		return fmt.Errorf("r.conn.Channel: %w", err)
	}
	defer func() {
		ch.Close()
	}()

	q, err := ch.QueueDeclare(
		codeQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg:  amqp.QueueTypeQuorum,
			"x-message-ttl":    attemptTTL,
			"x-delivery-limit": maxAttempts,
		},
	)
	if err != nil {
		return fmt.Errorf("ch.QueueDeclar: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("ch.PublishWithContext: %w", err)
	}
	return nil
}

func (r *RabbitQueue) AddResetCodeQueue(ctx context.Context, body []byte) error {
	ch, err := r.rc.NewChannel()
	if err != nil {
		return fmt.Errorf("r.conn.Channel: %w", err)
	}
	defer func() {
		ch.Close()
	}()

	q, err := ch.QueueDeclare(
		resetQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg:  amqp.QueueTypeQuorum,
			"x-message-ttl":    attemptTTL,
			"x-delivery-limit": maxAttempts,
		},
	)
	if err != nil {
		return fmt.Errorf("ch.QueueDeclar: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("ch.PublishWithContext: %w", err)
	}
	return nil
}
