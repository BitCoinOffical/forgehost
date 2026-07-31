package sender

import (
	"BitCoinOffical/forgehost/auth-service/internal/adapters/email"
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"BitCoinOffical/forgehost/auth-service/internal/worker/connect"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	connectionTimeout  = 10
	connectionAttempts = 5
)

type Worker struct {
	consumer *connect.Consumer
	logger   *zap.Logger
	rclient  *email.ResendClient
}

func NewWorker(logger *zap.Logger, rclient *email.ResendClient, consumer *connect.Consumer) *Worker {
	return &Worker{
		logger:   logger,
		rclient:  rclient,
		consumer: consumer,
	}
}

func (w *Worker) WorkerPool(worker int) <-chan error {
	wg := &sync.WaitGroup{}
	errs := make(chan error)

	for num := range worker {
		wg.Go(func() {
			for {
				consumer, err := w.consumer.StartConsumer()
				if err != nil {
					errs <- fmt.Errorf("StartConsumer: %w", err)
					time.Sleep(connectionTimeout * time.Second)
					continue
				}
				w.logger.Info("worker started", zap.Int("number", num+1))

				for msg := range consumer.Msgs {
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

					sentId, err := w.rclient.SendVerificationEmail([]string{body.Email}, body.Code)
					if err != nil {
						errs <- fmt.Errorf("w.rclient.SendVerificationEmail: %w", err)
						msg.Nack(false, true)
						continue
					}
					w.logger.Info("email sent successfully:", zap.Any("email id", sentId))
					msg.Ack(false)
				}
				w.logger.Info("worker stoped", zap.Int("number", num+1))
				if err := consumer.CloseConsumerChannel(); err != nil {
					errs <- fmt.Errorf("consumer.CloseConsumerChannel: %w worker number %d", err, num+1)
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
