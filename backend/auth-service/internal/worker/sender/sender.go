package sender

import (
	rabbitmq "BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	"BitCoinOffical/forgehost/auth-service/internal/adapters/email"
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/dto"

	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"go.uber.org/zap"
)

const (
	attemptTTL    = 10000 //10sec
	maxAttempts   = 5
	prefetchCount = 1
	prefetchSize  = 0
)

type Worker struct {
	logger   *zap.Logger
	rclient  *email.ResendClient
	conn     *rabbitmq.ResilientConnection
	TaskType string
}

func NewWorker(logger *zap.Logger, rclient *email.ResendClient, conn *rabbitmq.ResilientConnection, TaskType string) *Worker {
	return &Worker{
		logger:   logger,
		rclient:  rclient,
		conn:     conn,
		TaskType: TaskType,
	}
}

func (w *Worker) WorkerPool(worker int) <-chan error {
	wg := &sync.WaitGroup{}
	errs := make(chan error)

	for num := range worker {
		wg.Go(func() {
			for {
				ch, err := w.conn.NewChannel()
				if err != nil {
					w.logger.Error("c.conn.NewChannel", zap.Error(err))
					time.Sleep(time.Second)
					continue
				}

				q, err := ch.QueueDeclare(
					w.TaskType,
					true,
					false,
					false,
					false,
					amqp.Table{
						amqp.QueueTypeArg:  amqp.QueueTypeQuorum,
						"x-message-ttl":    attemptTTL,
						"x-delivery-limit": maxAttempts,
					},
				)
				if err != nil {
					ch.Close()
					w.logger.Error("ch.QueueDeclare", zap.Error(err))
					time.Sleep(time.Second)
					continue
				}

				err = ch.Qos(
					prefetchCount,
					prefetchSize,
					false,
				)
				if err != nil {
					ch.Close()
					w.logger.Error("ch.Qos", zap.Error(err))
					time.Sleep(time.Second)
					continue
				}

				msgs, err := ch.Consume(
					q.Name,
					"",
					false,
					false,
					false,
					false,
					nil,
				)
				if err != nil {
					ch.Close()
					w.logger.Error("ch.Consume", zap.Error(err))
					time.Sleep(time.Second)
					continue
				}

				w.logger.Info("consumer started", zap.String("queue", w.TaskType))
				w.logger.Info("worker started", zap.Int("number", num+1), zap.String("task", w.TaskType))

				for msg := range msgs {
					var body dto.RabbitQueueDTO
					if err := json.Unmarshal(msg.Body, &body); err != nil {
						errs <- fmt.Errorf("json.Unmarshal: %w", err)
						msg.Nack(false, false)
						continue
					}

					if time.Now().After(body.DispatchDate) {
						w.logger.Debug("time after", zap.Any("dispatch date", body.DispatchDate), zap.Any("now", time.Now()))
						errs <- domain.ErrLimitExceeded
						msg.Nack(false, false)
						continue
					}

					sentId, err := w.rclient.SendCodeEmail([]string{body.Email}, body.Code, body.TitleSubject)
					if err != nil {
						errs <- fmt.Errorf("w.rclient.SendVerificationEmail: %w", err)
						msg.Nack(false, true)
						continue
					}
					w.logger.Info("email sent successfully:", zap.Any("email id", sentId))
					msg.Ack(false)
				}
				ch.Close()
				w.logger.Info("worker stoped", zap.Int("number", num+1), zap.String("task", w.TaskType))
			}
		})
	}

	go func() {
		wg.Wait()
		defer close(errs)
	}()

	return errs
}
