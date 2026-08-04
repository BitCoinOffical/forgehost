package rabbitmq

import (
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	maxReconnect  = 10
	RetryInterval = 5
	Heartbeat     = 1
)

type ResilientConnection struct {
	url             string
	conn            *amqp.Connection
	channel         *amqp.Channel
	mu              sync.RWMutex
	publishMu       sync.Mutex
	notifyClose     chan *amqp.Error
	notifyChanClose chan *amqp.Error
	isReady         bool
	done            chan struct{}
	reconnectDelay  time.Duration
	maxReconnect    int
	logger          *zap.Logger
}

type RabbitURL struct {
	User string
	Pass string
	Host string
	Port string
}

func NewResilientConnection(cfg *RabbitURL, logger *zap.Logger) *ResilientConnection {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Pass, cfg.Host, cfg.Port)

	rc := &ResilientConnection{
		url:            url,
		done:           make(chan struct{}),
		reconnectDelay: time.Second,
		maxReconnect:   maxReconnect,
		logger:         logger,
	}
	go rc.Reconnect()
	return rc
}

func (rc *ResilientConnection) Reconnect() {
	for {
		rc.mu.Lock()
		rc.isReady = false
		rc.mu.Unlock()

		rc.logger.Info("Attempting to connect to RabbitMQ...")
		conn, err := rc.connect()
		if err != nil {
			rc.logger.Error("Failed to connect", zap.Error(err))
			select {
			case <-rc.done:
				return
			case <-time.After(rc.reconnectDelay):
				continue
			}
		}

		rc.mu.Lock()
		rc.conn = conn
		rc.isReady = true
		rc.mu.Unlock()

		rc.logger.Info("Connected to RabbitMQ successfully")

		select {
		case <-rc.done:
			return
		case err := <-rc.notifyClose:
			if err != nil {
				rc.logger.Error("Connection closed Reconnecting...", zap.Error(err))
			}
		case err := <-rc.notifyChanClose:
			if err != nil {
				rc.logger.Error("Channel closed Reconnecting...", zap.Error(err))
			}
		}
	}
}

func (rc *ResilientConnection) connect() (*amqp.Connection, error) {
	conn, err := amqp.Dial(rc.url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	rc.notifyClose = make(chan *amqp.Error, 1)
	rc.notifyChanClose = make(chan *amqp.Error, 1)
	conn.NotifyClose(rc.notifyClose)
	ch.NotifyClose(rc.notifyChanClose)

	rc.mu.Lock()
	rc.channel = ch
	rc.mu.Unlock()

	return conn, nil
}

func (rc *ResilientConnection) NewChannel() (*amqp.Channel, error) {
	timeout := time.After(30 * time.Second)
	for {
		rc.mu.RLock()
		if rc.isReady && rc.conn != nil {
			conn := rc.conn
			rc.mu.RUnlock()
			return conn.Channel()
		}
		rc.mu.RUnlock()

		select {
		case <-timeout:
			return nil, amqp.ErrClosed
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}
}

func (rc *ResilientConnection) Close() {
	close(rc.done)

	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.channel != nil {
		rc.channel.Close()
	}
	if rc.conn != nil {
		rc.conn.Close()
	}
}
