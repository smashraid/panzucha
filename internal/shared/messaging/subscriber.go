package messaging

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Subscriber is the consume-only counterpart to Broker (publish-only).
// Kept separate following interface segregation — a component that only
// publishes should not depend on subscription methods, and vice versa.
//
// Implementations:
//   - RabbitMQBroker  — production AMQP consumer
type Subscriber interface {
	// Subscribe opens a consumer channel with the given prefetch, declares
	// the queue durably, and returns a stream of deliveries that must be
	// acknowledged manually (AutoAck=false).
	Subscribe(ctx context.Context, queue string, prefetch int) (<-chan amqp.Delivery, error)
}
