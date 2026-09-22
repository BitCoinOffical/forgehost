package kafkaconn

import (
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaConn struct {
	Addr string
}

func NewKafkaClient(cfg *KafkaConn) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Addr),
	)
	if err != nil {
		return nil, fmt.Errorf("kgo.NewClient: %w", err)
	}
	return client, nil
}

func KafkaClose(client *kgo.Client) {
	client.Close()
}
