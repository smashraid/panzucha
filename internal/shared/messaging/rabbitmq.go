package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Compile-time guard: RabbitMQBroker must satisfy Broker.
var _ Broker = (*RabbitMQBroker)(nil)

// RabbitMQBroker implements Broker using amqp091.
// It owns the connection lifecycle including automatic reconnection.
// The exchange is declared once on Connect and reused for all publishes.
type RabbitMQBroker struct {
	url          string
	exchange     string
	conn         *amqp.Connection
	channel      *amqp.Channel
	mu           sync.RWMutex
	reconnecting bool
}

// NewRabbitMQBroker creates a RabbitMQBroker. Does not connect —
// call Connect() explicitly so main.go can handle startup failure.
func NewRabbitMQBroker(url, exchange string) *RabbitMQBroker {
	return &RabbitMQBroker{url: url, exchange: exchange}
}

// Connect establishes the AMQP connection, opens a channel, declares the
// topic exchange, and starts the reconnect watcher goroutine.
func (b *RabbitMQBroker) Connect() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.connect()
}

func (b *RabbitMQBroker) connect() error {
	conn, err := amqp.Dial(b.url)
	if err != nil {
		return fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("rabbitmq open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		b.exchange,
		"topic", // topic exchange: routing keys like "order.created", "order.*"
		true,    // durable — survives broker restart
		false,   // auto-delete
		false,   // internal
		false,   // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("rabbitmq declare exchange %q: %w", b.exchange, err)
	}

	connClose := make(chan *amqp.Error, 1)
	conn.NotifyClose(connClose)
	go b.handleReconnect(connClose)

	b.conn = conn
	b.channel = ch

	// Log exchange name only — never the URL which contains credentials.
	slog.Info("rabbitmq: connected", "exchange", b.exchange)
	return nil
}

// handleReconnect watches the connection close notification and retries
// with exponential backoff. Exits when Close() is called cleanly.
func (b *RabbitMQBroker) handleReconnect(connClose <-chan *amqp.Error) {
	amqpErr, ok := <-connClose
	if !ok {
		return // clean Close() — do not reconnect
	}

	slog.Warn("rabbitmq: connection lost, reconnecting",
		"reason", amqpErr.Reason,
		"code", amqpErr.Code,
	)

	b.mu.Lock()
	b.reconnecting = true
	b.mu.Unlock()

	backoff := 1 * time.Second
	const maxBackoff = 30 * time.Second

	for {
		time.Sleep(backoff)

		b.mu.Lock()
		err := b.connect()
		if err == nil {
			b.reconnecting = false
			b.mu.Unlock()
			slog.Info("rabbitmq: reconnected successfully")
			return
		}
		b.mu.Unlock()

		slog.Warn("rabbitmq: reconnect attempt failed",
			"err", err,
			"backoff", backoff,
		)

		backoff = min(backoff*2, maxBackoff)
	}
}

// Publish sends a pre-serialised JSON payload to the exchange.
// Returns an error immediately if the broker is reconnecting so the
// caller can decide to retry or accept the loss.
func (b *RabbitMQBroker) Publish(ctx context.Context, routingKey string, payload []byte) error {
	b.mu.RLock()
	ch := b.channel
	reconnecting := b.reconnecting
	b.mu.RUnlock()

	if reconnecting || ch == nil {
		return fmt.Errorf("rabbitmq: broker not ready (reconnecting)")
	}

	return ch.PublishWithContext(ctx,
		b.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,  // survives broker restart when queue is durable
			MessageId:    uuid.NewString(), // for consumer-side deduplication
			Timestamp:    time.Now().UTC(),
			Body:         payload,
		},
	)
}

// Close shuts down the channel and connection cleanly.
// The reconnect goroutine detects the clean close and exits without retrying.
func (b *RabbitMQBroker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.channel != nil {
		_ = b.channel.Close()
		b.channel = nil
	}
	if b.conn != nil {
		_ = b.conn.Close()
		b.conn = nil
	}
	slog.Info("rabbitmq: connection closed")
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
