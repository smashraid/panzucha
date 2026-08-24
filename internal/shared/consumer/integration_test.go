package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"

	"panzucha/internal/shared/consumer"
	"panzucha/internal/shared/db"
	"panzucha/internal/shared/inbox"
	"panzucha/internal/shared/messaging"
)

// TestIntegrationDuplicateDelivery proves the transactional inbox absorbs a
// forced redelivery against REAL RabbitMQ and Postgres: two messages sharing
// the same MessageId (= outbox EventID, the dedup key) must result in exactly
// one handler invocation, both deliveries acked, and a
// "duplicate event skipped" log for the second.
//
// Skipped when the local stack is not running, so CI without services stays green.
func TestIntegrationDuplicateDelivery(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	const (
		rabbitURL = "amqp://guest:guest@localhost:5672/"
		dbURL     = "postgres://admin:admin@localhost:5432/panzucha?sslmode=disable"
	)

	broker := messaging.NewRabbitMQBroker(rabbitURL, "it.events."+uuid.NewString()[:8])
	if err := broker.Connect(); err != nil {
		t.Skipf("rabbitmq not available: %v", err)
	}
	defer broker.Close()

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("postgres not available: %v", err)
	}

	runID := uuid.NewString()
	queueSpec := messaging.QueueSpec{
		Name:       "it.queue." + runID,
		RoutingKey: "it.created." + runID,
		DLX:        "it.dlx." + runID,
		DLQ:        "it.dlq." + runID,
	}
	if err := broker.DeclareQueue(context.Background(), queueSpec); err != nil {
		t.Fatalf("declare queue topology: %v", err)
	}
	t.Cleanup(func() { cleanupTopology(t, rabbitURL, queueSpec) })

	eventID := uuid.NewString()
	payload, err := json.Marshal(map[string]any{"event_id": eventID, "order_id": "it-ord-1"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	// Two messages, same MessageId — equivalent to a forced management-UI
	// redelivery, since inbox dedup keys solely on the MessageId.
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if err := broker.Publish(ctx, queueSpec.RoutingKey, eventID, payload); err != nil {
			t.Fatalf("publish %d: %v", i+1, err)
		}
	}

	var mu sync.Mutex
	handlerCalls := 0
	handler := func(_ context.Context, _ pgxTx, _ []byte) error {
		mu.Lock()
		defer mu.Unlock()
		handlerCalls++
		return nil
	}

	transactor := db.NewPgxTransactor(pool)
	inboxRepo := inbox.NewPostgresInboxRepository(pool)
	orderConsumer := consumer.New(broker, transactor, inboxRepo, queueSpec, handler)

	consumerCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- orderConsumer.Start(consumerCtx) }()

	// Wait until both deliveries are resolved: handler ran once AND the
	// duplicate was logged as skipped.
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		calls := handlerCalls
		mu.Unlock()
		if calls >= 1 && strings.Contains(logOutput(), "duplicate event skipped") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("consumer returned error: %v", err)
	}

	mu.Lock()
	calls := handlerCalls
	mu.Unlock()
	if calls != 1 {
		t.Errorf("handler called %d times, want exactly 1 (duplicate must be absorbed)", calls)
	}

	logs := logOutput()
	if !strings.Contains(logs, "duplicate event skipped") {
		t.Error("log does not contain \"duplicate event skipped\"")
	}
	if !strings.Contains(logs, eventID) {
		t.Errorf("log does not reference deduped event_id %q", eventID)
	}
}

// pgxTx aliases pgx.Tx to keep the handler signature readable.
type pgxTx = pgx.Tx

// TestIntegrationHandlerFailureRoutesToDLQ proves a handler error nacks with
// requeue=false and RabbitMQ dead-letters the message to the configured DLQ:
// the DLQ receives it with MessageId/body intact, and no redelivery loop
// occurs. Skipped when the local stack is not running.
func TestIntegrationHandlerFailureRoutesToDLQ(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	const (
		rabbitURL = "amqp://guest:guest@localhost:5672/"
		dbURL     = "postgres://admin:admin@localhost:5432/panzucha?sslmode=disable"
	)

	broker := messaging.NewRabbitMQBroker(rabbitURL, "it.events."+uuid.NewString()[:8])
	if err := broker.Connect(); err != nil {
		t.Skipf("rabbitmq not available: %v", err)
	}
	defer broker.Close()

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("postgres not available: %v", err)
	}

	runID := uuid.NewString()
	queueSpec := messaging.QueueSpec{
		Name:       "it.queue." + runID,
		RoutingKey: "it.created." + runID,
		DLX:        "it.dlx." + runID,
		DLQ:        "it.dlq." + runID,
	}
	if err := broker.DeclareQueue(context.Background(), queueSpec); err != nil {
		t.Fatalf("declare queue topology: %v", err)
	}
	t.Cleanup(func() { cleanupTopology(t, rabbitURL, queueSpec) })

	eventID := uuid.NewString()
	payload, err := json.Marshal(map[string]any{"event_id": eventID, "order_id": "it-ord-2"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	ctx := context.Background()
	if err := broker.Publish(ctx, queueSpec.RoutingKey, eventID, payload); err != nil {
		t.Fatalf("publish: %v", err)
	}

	var mu sync.Mutex
	handlerCalls := 0
	handler := func(_ context.Context, _ pgxTx, _ []byte) error {
		mu.Lock()
		defer mu.Unlock()
		handlerCalls++
		return errors.New("forced failure")
	}

	transactor := db.NewPgxTransactor(pool)
	inboxRepo := inbox.NewPostgresInboxRepository(pool)
	orderConsumer := consumer.New(broker, transactor, inboxRepo, queueSpec, handler)

	consumerCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- orderConsumer.Start(consumerCtx) }()

	// Poll the DLQ until the dead-lettered message arrives.
	var dld amqp.Delivery
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := amqp.Dial(rabbitURL)
		if err != nil {
			t.Fatalf("dlq dial: %v", err)
		}
		ch, err := conn.Channel()
		if err != nil {
			_ = conn.Close()
			t.Fatalf("dlq channel: %v", err)
		}
		d, ok, err := ch.Get(queueSpec.DLQ, false)
		_ = ch.Close()
		_ = conn.Close()
		if err == nil && ok {
			dld = d
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("consumer returned error: %v", err)
	}

	if dld.MessageId == "" {
		t.Fatal("no message arrived in the DLQ within deadline")
	}
	if dld.MessageId != eventID {
		t.Errorf("DLQ MessageId = %q, want %q (dead-letter must preserve identity)", dld.MessageId, eventID)
	}
	if string(dld.Body) != string(payload) {
		t.Errorf("DLQ body = %s, want %s", dld.Body, payload)
	}

	mu.Lock()
	calls := handlerCalls
	mu.Unlock()
	if calls != 1 {
		t.Errorf("handler called %d times, want exactly 1 (requeue=false must not loop)", calls)
	}
	if !strings.Contains(logOutput(), "handler error, routing to DLQ") {
		t.Error("log does not contain \"handler error, routing to DLQ\"")
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

var testLogBuf = &syncBuffer{}

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(testLogBuf, nil)))
}

func logOutput() string {
	return testLogBuf.String()
}

func cleanupTopology(t *testing.T, url string, spec messaging.QueueSpec) {
	t.Helper()
	conn, err := amqp.Dial(url)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()
	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer func() { _ = ch.Close() }()
	for _, name := range []string{spec.Name, spec.DLQ} {
		_, _ = ch.QueueDelete(name, false, false, false)
	}
	_ = ch.ExchangeDelete(spec.DLX, false, false)
}
