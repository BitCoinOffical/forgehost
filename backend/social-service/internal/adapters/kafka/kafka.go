package kafkaread

import (
	"fmt"

	"github.com/segmentio/kafka-go"
)

const (
	topic   = "user.social"
	groupID = "social-service"
)

type KafkaConfig struct {
	Addr string
}

func NewKafkaReaвer(cfg *KafkaConfig) *kafka.Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Addr},
		Topic:   topic,
		GroupID: groupID,
	})

	return reader
}

func KafkaClose(reader *kafka.Reader) error {
	if err := reader.Close(); err != nil {
		return fmt.Errorf("reader.Close: %v", err)
	}
	return nil
}
