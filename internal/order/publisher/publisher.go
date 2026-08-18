package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"panzucha/internal/order/domain"
	"panzucha/internal/shared/messaging"
)

// Publisher implements domain.EventPublisher.
// It is the adapter between the domain layer and the messaging transport.
// Single responsibility: translate domain entities into event payloads
// and dispatch them through the Broker.
//
// Dependency direction:
//
//	domain.EventPublisher  ← service depends on this interface
//	      ↑
//	publisher.Publisher    ← this struct implements it
//	      ↓
//	messaging.Broker       ← publisher delegates raw sending here
//	      ↑
//	messaging.RabbitMQBroker / NoopBroker
type Publisher struct {
	broker messaging.Broker
}

var _ domain.EventPublisher = (*Publisher)(nil)

// New creates a Publisher backed by the given Broker.
// Production: pass messaging.NewRabbitMQBroker(url, exchange)
// Tests:      pass messaging.NewNoopBroker()
func New(broker messaging.Broker) *Publisher {
	return &Publisher{broker: broker}
}

// Publish is the domain.EventPublisher implementation.
// Serialises the event and delegates the raw bytes to the broker.
// Called indirectly by the domain-specific methods below.
func (p *Publisher) Publish(ctx context.Context, routingKey string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("publisher: marshal %q: %w", routingKey, err)
	}
	return p.broker.Publish(ctx, routingKey, payload)
}

// ── Domain-specific publish methods ──────────────────────────────────────────
// These translate domain entities into event structs.
// The service calls these — never constructs event structs itself.
// Publishing failure after a committed transaction is non-fatal:
// the order is durable, the event is best-effort. In production,
// use the transactional outbox pattern for guaranteed delivery.

func (p *Publisher) OrderCreated(ctx context.Context, order *domain.Order) {
	items := make([]messaging.OrderItem, len(order.Items))
	for i, it := range order.Items {
		items[i] = messaging.OrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
		}
	}

	event := messaging.OrderCreatedEvent{
		BaseEvent:   newBase(messaging.EventOrderCreated),
		OrderID:     order.ID,
		UserID:      order.UserID,
		Items:       items,
		TotalAmount: order.TotalAmount,
	}

	if err := p.Publish(ctx, messaging.EventOrderCreated, event); err != nil {
		slog.ErrorContext(ctx, "publisher: order.created failed",
			"err", err, "order_id", order.ID)
	}
}

func (p *Publisher) OrderFailed(ctx context.Context, orderID, reason string) {
	event := messaging.OrderFailedEvent{
		BaseEvent: newBase(messaging.EventOrderFailed),
		OrderID:   orderID,
		Reason:    reason,
	}
	if err := p.Publish(ctx, messaging.EventOrderFailed, event); err != nil {
		slog.ErrorContext(ctx, "publisher: order.failed failed",
			"err", err, "order_id", orderID)
	}
}

func (p *Publisher) OrderPaid(ctx context.Context, orderID, paymentID string) {
	event := messaging.OrderPaidEvent{
		BaseEvent: newBase(messaging.EventOrderPaid),
		OrderID:   orderID,
		PaidAt:    time.Now().UTC(),
		PaymentID: paymentID,
	}
	if err := p.Publish(ctx, messaging.EventOrderPaid, event); err != nil {
		slog.ErrorContext(ctx, "publisher: order.paid failed",
			"err", err, "order_id", orderID)
	}
}

func newBase(eventType string) messaging.BaseEvent {
	return messaging.BaseEvent{
		EventID:   uuid.NewString(),
		EventType: eventType,
		Timestamp: time.Now().UTC(),
	}
}
