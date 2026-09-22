package kafkaread

import (
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	topic   = "user.social"
	groupID = "social-service"
)

type KafkaConfig struct {
	Addr string
}

func NewKafkaClient(cfg *KafkaConfig) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Addr),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		return nil, fmt.Errorf("kgo.NewClient: %w", err)
	}

	return client, nil
}

func KafkaClose(client *kgo.Client) {
	client.Close()
}
