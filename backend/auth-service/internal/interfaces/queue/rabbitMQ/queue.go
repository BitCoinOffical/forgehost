package rabbitqueue

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	attemptTTL  = 10000 //10sec
	maxAttempts = 5
)

type RabbitQueue struct {
	conn *amqp.Connection
}

func NewQueue(conn *amqp.Connection) *RabbitQueue {
	return &RabbitQueue{
		conn: conn,
	}
}

func (r *RabbitQueue) AddQueue(ctx context.Context, body []byte) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("r.conn.Channel: %w", err)
	}
	defer func() {
		ch.Close()
	}()

	q, err := ch.QueueDeclare(
		"code_queue",
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
