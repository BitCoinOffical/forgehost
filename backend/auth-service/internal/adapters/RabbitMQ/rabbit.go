package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitURL struct {
	User string
	Pass string
	Host string
	Port string
}

func NewRabbitMQ(cfg *RabbitURL) (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Pass, cfg.Host, cfg.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("amqp.Dial: %w", err)
	}
	return conn, nil
}

func CloseRabbitMQ(conn *amqp.Connection) error {
	if err := conn.Close(); err != nil {
		return fmt.Errorf("conn.Close: %w", err)
	}
	return nil
}
