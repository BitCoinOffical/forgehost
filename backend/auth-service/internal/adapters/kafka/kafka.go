package kafkaconn

import (
	"fmt"

	"github.com/segmentio/kafka-go"
)

const (
	topic = "user.social"
)

type KafkaConn struct {
	Addr string
}

func NewKafkaConn(cfg *KafkaConn) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Addr),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return writer
}

func KafkaClose(writer *kafka.Writer) error {
	if err := writer.Close(); err != nil {
		return fmt.Errorf("writer.Close: %w", err)
	}
	return nil
}
