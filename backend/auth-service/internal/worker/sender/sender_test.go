package sender

import (
	"context"
	v2 "encoding/json/v2"

	"testing"
	"time"

	notificationv1 "github.com/BitCoinOffical/forgehost-proto/notification/v1"
	rabbitmq "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	sender_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/worker/sender/mock"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	rabbittc "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
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

func TestWorkerPool(t *testing.T) {
	rc := setupTestRabbit(t)
	ctrl := gomock.NewController(t)
	client := sender_mocks.NewMockNotificationClient(ctrl)
	called := make(chan struct{})
	client.EXPECT().SendEmail(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *notificationv1.SendEmailRequest) (*notificationv1.SendEmailResponse, error) {
			called <- struct{}{}
			close(called)
			return &notificationv1.SendEmailResponse{}, nil
		}).Times(1)
	worker := NewWorker(zaptest.NewLogger(t), client, rc)

	ch, err := rc.NewChannel()
	require.NoError(t, err)

	q, err := ch.QueueDeclare(nameQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg:  amqp.QueueTypeQuorum,
			"x-message-ttl":    attemptTTL,
			"x-delivery-limit": maxAttempts,
		})
	require.NoError(t, err)

	body := models.RabbitQueue{
		Code:  "1234",
		Email: "user@gmail.com",

		TitleSubject: "title",
		DispatchDate: time.Now().Add(time.Minute),
	}
	b, err := v2.Marshal(&body)
	require.NoError(t, err)

	testCases := []struct {
		caseName string
		worker   int
		succes   bool
	}{
		{
			caseName: "get task",
			worker:   1,
			succes:   true,
		},
	}

	err = ch.PublishWithContext(
		t.Context(),
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         b,
		},
	)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			t.Cleanup(func() {
				cancel()
			})
			errs := worker.WorkerPool(ctx, tc.worker)
			select {
			case <-called:
				if tc.succes {
					require.NoError(t, <-errs)
					return
				}
				require.Error(t, <-errs)
				return
			case <-time.After(5 * time.Second):
				t.Fatal("time out")
			}
		})
	}
}
