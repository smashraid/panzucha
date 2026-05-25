package messaging

import (
	"fmt"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex
	url     string
}

// NewRabbitMQ creates a new RabbitMQ client (does not connect yet)
func NewRabbitMQ(url string) *RabbitMQ {
	return &RabbitMQ{url: url}
}

// Connect establishes the connection and channel
func (r *RabbitMQ) Connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.conn = conn
	r.channel = ch
	slog.Info("Connected to RabbitMQ", "url", r.url)
	return nil
}

// Close cleans up the connection
func (r *RabbitMQ) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

// GetChannel returns the AMQP channel (thread-safe)
func (r *RabbitMQ) GetChannel() *amqp.Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channel
}

// DeclareExchange ensures the exchange exists
func (r *RabbitMQ) DeclareExchange(name, kind string, durable bool) error {
	return r.channel.ExchangeDeclare(
		name,    // name
		kind,    // type: "direct", "topic", "fanout"
		durable, // durable
		false,   // auto-delete
		false,   // internal
		false,   // no-wait
		nil,     // args
	)
}

// DeclareQueue ensures the queue exists and binds it to an exchange
func (r *RabbitMQ) DeclareQueue(name, exchange, routingKey string, durable bool) error {
	_, err := r.channel.QueueDeclare(
		name,    // name
		durable, // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}
	return r.channel.QueueBind(name, routingKey, exchange, false, nil)
}
