package sender

import (
	"context"

	notificationv1 "github.com/BitCoinOffical/forgehost-proto/notification/v1"
	rabbitmq "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/notification"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"

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
	nameQueue     = "auth.email_tasks"
)

type Worker struct {
	logger *zap.Logger
	client *notification.NotificationClient
	conn   *rabbitmq.ResilientConnection
}

func NewWorker(logger *zap.Logger, client *notification.NotificationClient, conn *rabbitmq.ResilientConnection) *Worker {
	return &Worker{
		logger: logger,
		conn:   conn,
		client: client,
	}
}

func (w *Worker) WorkerPool(ctx context.Context, worker int) <-chan error {
	wg := &sync.WaitGroup{}
	errs := make(chan error)

	for num := range worker {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					w.logger.Info("worker shut down", zap.Int("num", num+1))
					return
				default:
					ch, err := w.conn.NewChannel()
					if err != nil {
						w.logger.Error("c.conn.NewChannel", zap.Error(err))
						time.Sleep(time.Second)
						continue
					}

					q, err := ch.QueueDeclare(
						nameQueue,
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

					w.logger.Info("worker started", zap.Int("number", num+1))

					for msg := range msgs {
						var body dto.RabbitQueueDTO
						if err := json.Unmarshal(msg.Body, &body); err != nil {
							errs <- fmt.Errorf("json.Unmarshal: %w", err)
							if nackErr := msg.Nack(false, false); nackErr != nil {
								w.logger.Debug("worker: nack failed", zap.Error(nackErr))
							}
							continue
						}

						if time.Now().After(body.DispatchDate) {
							w.logger.Debug("time after", zap.Time("dispatch_date", body.DispatchDate), zap.Time("now", time.Now()), zap.String("user_id", body.UserID))
							errs <- domain.ErrLimitExceeded
							if nackErr := msg.Nack(false, false); nackErr != nil {
								w.logger.Debug("worker: nack failed", zap.String("user_id", body.UserID), zap.Error(nackErr))
							}
							continue
						}

						res, err := w.client.SendEmail(context.Background(), &notificationv1.SendEmailRequest{
							UserId:       body.UserID,
							Email:        body.Email,
							Code:         body.Code,
							TitleSubject: body.TitleSubject,

							DispatchDate: timestamppb.New(body.DispatchDate),
						})
						if err != nil {
							errs <- fmt.Errorf("w.client.SendEmail: %w", err)
							if nackErr := msg.Nack(false, true); nackErr != nil {
								w.logger.Debug("worker: nack failed", zap.String("user_id", body.UserID), zap.Error(nackErr))
							}
							continue
						}
						w.logger.Info("email send", zap.String("email_id", res.GetEmailId()), zap.String("status", res.GetStatus()), zap.String("user_id", body.UserID))
						if ackErr := msg.Ack(false); ackErr != nil {
							w.logger.Debug("worker: ack failed", zap.String("user_id", body.UserID), zap.Error(ackErr))
						}
					}
					ch.Close()
					w.logger.Info("worker stoped", zap.Int("number", num+1))
				}
			}
		})
	}

	go func() {
		wg.Wait()
		defer close(errs)
	}()

	return errs
}
