package rabbitqueue

import (
	"context"
	"testing"

	rabbitmq "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
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
			if tc.succes {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
		})
	}

}
