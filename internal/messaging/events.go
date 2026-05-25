package messaging

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType constants for routing keys
const (
	EventOrderCreated = "order.created"
	EventOrderPaid    = "order.paid"
	EventOrderFailed  = "order.failed"
	EventOrderShipped = "order.shipped"
)

// BaseEvent contains common fields for all events
type BaseEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

// NewBaseEvent creates a new base event with generated ID and current timestamp
func NewBaseEvent(eventType string) BaseEvent {
	return BaseEvent{
		EventID:   uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now().UTC(),
	}
}

// OrderCreatedEvent is published when an order is created (status = pending)
type OrderCreatedEvent struct {
	BaseEvent
	OrderID    string  `json:"order_id"`
	UserID     string  `json:"user_id"`
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
}

// NewOrderCreatedEvent creates an OrderCreatedEvent with populated base fields
func NewOrderCreatedEvent(orderID, userID, productID string, quantity int, totalPrice float64) OrderCreatedEvent {
	return OrderCreatedEvent{
		BaseEvent:  NewBaseEvent(EventOrderCreated),
		OrderID:    orderID,
		UserID:     userID,
		ProductID:  productID,
		Quantity:   quantity,
		TotalPrice: totalPrice,
	}
}

// OrderPaidEvent is published after successful payment
type OrderPaidEvent struct {
	BaseEvent
	OrderID   string    `json:"order_id"`
	PaidAt    time.Time `json:"paid_at"`
	PaymentID string    `json:"payment_id,omitempty"` // optional, for future integration
}

// NewOrderPaidEvent creates an OrderPaidEvent
func NewOrderPaidEvent(orderID string, paymentID string) OrderPaidEvent {
	return OrderPaidEvent{
		BaseEvent: NewBaseEvent(EventOrderPaid),
		OrderID:   orderID,
		PaidAt:    time.Now().UTC(),
		PaymentID: paymentID,
	}
}

// OrderFailedEvent is published when order processing fails (e.g., payment failure)
type OrderFailedEvent struct {
	BaseEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// NewOrderFailedEvent creates an OrderFailedEvent
func NewOrderFailedEvent(orderID, reason string) OrderFailedEvent {
	return OrderFailedEvent{
		BaseEvent: NewBaseEvent(EventOrderFailed),
		OrderID:   orderID,
		Reason:    reason,
	}
}

// OrderShippedEvent is published when order is shipped (future use)
type OrderShippedEvent struct {
	BaseEvent
	OrderID   string    `json:"order_id"`
	ShippedAt time.Time `json:"shipped_at"`
	Tracking  string    `json:"tracking,omitempty"`
}

// Helper to JSON‑serialise any event
func MarshalEvent(event interface{}) ([]byte, error) {
	return json.Marshal(event)
}

func UnmarshalEvent(data []byte, event interface{}) error {
	return json.Unmarshal(data, event)
}
