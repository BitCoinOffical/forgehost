package rabbitqueue

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	rabbittc "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestRabbit(t *testing.T) *amqp.Connection {
	t.Helper()
	ctx := context.Background()
	container, err := rabbittc.Run(ctx, "rabbitmq:4.1-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete"),
		))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	strURL, err := container.AmqpURL(ctx)

	conn, err := amqp.Dial(strURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

// func TestAddEmailTaskQueue(t *testing.T) {
// 	conn := setupTestRabbit(t)
// 	queue := NewQueue()

// 	testCases := []struct {
// 		caseName string
// 		body     []byte
// 		succes   bool
// 	}{
// 		{
// 			caseName: "add in task queue",
// 			body:     []byte("task"),
// 			succes:   true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.caseName, func(t *testing.T) {
// 			err := queue.AddEmailTaskQueue(context.Background(), tc.body)
// 			if tc.succes {
// 				require.NoError(t, err)
// 				return
// 			}
// 			require.Error(t, err)
// 		})
// 	}

// }
