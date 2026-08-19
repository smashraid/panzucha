package messaging

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// QueueSpec describes the full topology of one consumer queue: the primary
// queue, its topic routing key, and an optional dead-letter exchange/queue.
// Dead-lettered messages are re-routed to the DLX under the same routing key.
type QueueSpec struct {
	Name       string // primary queue, e.g. "order.created"
	RoutingKey string // topic binding key, e.g. "order.created"
	DLX        string // dead-letter exchange, e.g. "order.events.dlx"; empty = no DLQ
	DLQ        string // dead-letter queue, e.g. "order.created.dlq"; empty = no DLQ
}

// Subscriber is the consume-only counterpart to Broker (publish-only).
// Kept separate following interface segregation — a component that only
// publishes should not depend on subscription methods, and vice versa.
//
// Implementations:
//   - RabbitMQBroker  — production AMQP consumer
type Subscriber interface {
	// Subscribe ensures the queue topology (primary queue + DLQ + bindings)
	// exists, opens a consumer channel with the given prefetch, and returns a
	// stream of deliveries that must be acknowledged manually (AutoAck=false).
	Subscribe(ctx context.Context, q QueueSpec, prefetch int) (<-chan amqp.Delivery, error)
}
