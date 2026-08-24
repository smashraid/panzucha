package messaging

import "context"

// Broker is the transport-agnostic interface for publishing raw messages.
// It hides whether the underlying broker is RabbitMQ, Kafka, SQS, or a
// no-op used in tests. Nothing outside the messaging package depends on
// amqp, kafka client, or any other broker SDK directly.
//
// Implementations:
//   - RabbitMQBroker  — production AMQP publisher
//   - NoopBroker      — silent drop for tests and local dev
type Broker interface {
	// Publish sends a raw JSON payload to the given routing key.
	// eventID is the canonical outbox EventID — implementations MUST carry it
	// as the message's MessageId so consumer-side inbox deduplication keys on
	// the same identifier the producer committed. The broker decides what
	// "routing key" means — topic in RabbitMQ, topic name in Kafka, queue URL
	// in SQS.
	Publish(ctx context.Context, routingKey string, eventID string, payload []byte) error

	// Close cleanly shuts down the broker connection.
	Close()
}
