package connect

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	attemptTTL    = 10000 //10sec
	maxAttempts   = 5
	prefetchCount = 1
	prefetchSize  = 0
	queueName     = "code_queue"
)

type ConsumerData struct {
	Ch   *amqp.Channel
	Msgs <-chan amqp.Delivery
}

type Consumer struct {
	conn   *amqp.Connection
	logger *zap.Logger
}

func NewRunConsumer(conn *amqp.Connection, logger *zap.Logger) *Consumer {
	return &Consumer{conn: conn, logger: logger}
}

func (c *Consumer) StartConsumer() (*ConsumerData, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("ch.QueueDeclar: %w", err)
	}

	err = ch.Qos(
		prefetchCount,
		prefetchSize,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("ch.Qos: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("ch.Consume: %w", err)
	}

	c.logger.Info("consumer started")
	return &ConsumerData{
		Ch:   ch,
		Msgs: msgs,
	}, nil
}

func (c *ConsumerData) CloseConsumerChannel() error {
	if err := c.Ch.Close(); err != nil {
		return fmt.Errorf("ch.Close(): %w", err)
	}
	return nil
}
