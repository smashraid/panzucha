package domain

import "time"

// Routing key constants used as RabbitMQ topic routing keys.
// Consumers subscribe using patterns: "order.*" receives all order events.
const (
	EventOrderCreated = "order.created"
	EventOrderPaid    = "order.paid"
	EventOrderFailed  = "order.failed"
	EventOrderShipped = "order.shipped"
)

// BaseEvent carries fields every event must have.
// EventID enables consumer-side deduplication.
// Timestamp is always UTC.
type BaseEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderCreatedEvent is published after a successful order commit.
// Items contains the full line items — not a single ProductID — because
// an order can contain multiple products.
type OrderCreatedEvent struct {
	BaseEvent
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
}

// OrderPaidEvent is published after payment confirmation.
type OrderPaidEvent struct {
	BaseEvent
	OrderID   string    `json:"order_id"`
	PaidAt    time.Time `json:"paid_at"`
	PaymentID string    `json:"payment_id,omitempty"`
}

// OrderFailedEvent is published when order processing fails.
type OrderFailedEvent struct {
	BaseEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// OrderShippedEvent is published when an order ships.
type OrderShippedEvent struct {
	BaseEvent
	OrderID   string    `json:"order_id"`
	ShippedAt time.Time `json:"shipped_at"`
	Tracking  string    `json:"tracking,omitempty"`
}
