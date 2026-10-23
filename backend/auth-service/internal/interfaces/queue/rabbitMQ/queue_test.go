package rabbitqueue

import (
	"context"
	"testing"
	"time"

	rabbitmq "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	rabbittc "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap/zaptest"
)

const (
	username = "test"
	password = "test"
)

func setupTestRabbit(t *testing.T) *rabbitmq.ResilientConnection {
	t.Helper()
	ctx := t.Context()
	container, err := rabbittc.Run(ctx, "rabbitmq:4.1-alpine",
		rabbittc.WithAdminUsername(username),
		rabbittc.WithAdminPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete"),
		))
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "5672/tcp")
	if err != nil {
		t.Fatal(err)
	}

	rc := rabbitmq.NewResilientConnection(&rabbitmq.RabbitURL{
		User: username,
		Pass: password,
		Host: host,
		Port: port.Port(),
	}, zaptest.NewLogger(t))

	t.Cleanup(rc.Close)

	return rc
}

func TestAddEmailTaskQueue(t *testing.T) {
	rc := setupTestRabbit(t)
	queue := NewQueue(rc)
	ch, err := rc.NewChannel()
	require.NoError(t, err)

	q, err := ch.QueueDeclare(nameQueue, true, false, false, false, amqp.Table{
		amqp.QueueTypeArg:  amqp.QueueTypeQuorum,
		"x-message-ttl":    attemptTTL,
		"x-delivery-limit": maxAttempts,
	})
	require.NoError(t, err)

	testCases := []struct {
		caseName string
		body     []byte
		succes   bool
	}{
		{
			caseName: "add in task queue",
			body:     []byte("task"),
			succes:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			err := queue.AddEmailTaskQueue(context.Background(), tc.body)
			require.NoError(t, err)

			var got amqp.Delivery
			require.Eventually(t, func() bool {
				msg, ok, err := ch.Get(q.Name, true)
				if err != nil || !ok {
					return false
				}
				got = msg
				return true
			}, 5*time.Second, 100*time.Millisecond)

			if tc.succes {
				require.NoError(t, err)
				assert.Equal(t, tc.body, got.Body)
				return
			}

			require.Error(t, err)
		})
	}
	ch.Close()
}
