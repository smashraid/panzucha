package messaging

import "context"

// NoopBroker implements Broker by silently discarding all messages.
// Use in tests and local development where a real broker is not available.
// Swap it in via dependency injection — no code changes needed in services.
//
// Example in tests:
//
//	broker := messaging.NewNoopBroker()
//	relay := outbox.NewRelay(pool, repo, broker, outbox.Config{})
type NoopBroker struct{}

var _ Broker = (*NoopBroker)(nil)

func NewNoopBroker() *NoopBroker { return &NoopBroker{} }

func (n *NoopBroker) Publish(_ context.Context, _, _ string, _ []byte) error { return nil }
func (n *NoopBroker) Close()                                                 {}
