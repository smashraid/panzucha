package domain

import "context"

// EventPublisher is the domain contract for publishing events.
// The service layer depends on this interface — never on publisher.Publisher
// or any messaging package directly.
//
// Domain-specific methods (OrderCreated, OrderFailed) are declared here so
// the service can call them without a type assertion. The mock in tests
// implements all methods — usually as no-ops.
type EventPublisher interface {
	// Publish is the low-level method used for any routing key and payload.
	Publish(ctx context.Context, routingKey string, event any) error

	// OrderCreated publishes an event after a successful order commit.
	// Non-fatal — publishing failure does not affect the committed order.
	OrderCreated(ctx context.Context, order *Order)

	// OrderFailed publishes an event when order processing fails.
	OrderFailed(ctx context.Context, orderID, reason string)

	// OrderPaid publishes an event after payment confirmation.
	OrderPaid(ctx context.Context, orderID, paymentID string)
}
