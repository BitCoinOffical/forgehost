package services

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	kafkatc "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	cluster = "test-cluster"
)

func setupTestKafka(t *testing.T) *kgo.Client {
	t.Helper()

	container, err := kafkatc.Run(t.Context(), "confluentinc/confluent-local:7.5.0",
		kafkatc.WithClusterID(cluster),
	)

	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	brokers, err := container.Brokers(t.Context())
	require.NoError(t, err)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	adm := kadm.NewClient(client)
	_, err = adm.CreateTopic(t.Context(), 1, 1, nil, topic)
	require.NoError(t, err)

	return client
}
