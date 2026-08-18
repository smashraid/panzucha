package consumer_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"panzucha/internal/shared/consumer"
	"panzucha/internal/shared/db"
	shareddomain "panzucha/internal/shared/domain"
	"panzucha/internal/shared/inbox"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

// --- fakes ---

type fakeSubscriber struct {
	deliveries chan amqp.Delivery
}

func (f *fakeSubscriber) Subscribe(ctx context.Context, queue string, prefetch int) (<-chan amqp.Delivery, error) {
	return f.deliveries, nil
}

// fakeAcknowledger records how each delivery was resolved.
type fakeAcknowledger struct {
	mu       sync.Mutex
	acked    []uint64
	nacked   []uint64
	requeued []bool
}

func (f *fakeAcknowledger) Ack(tag uint64, multiple bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.acked = append(f.acked, tag)
	return nil
}

func (f *fakeAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nacked = append(f.nacked, tag)
	f.requeued = append(f.requeued, requeue)
	return nil
}

func (f *fakeAcknowledger) Reject(tag uint64, requeue bool) error {
	return f.Nack(tag, false, requeue)
}

func (f *fakeAcknowledger) calls() (acked, nacked []uint64, requeued []bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]uint64{}, f.acked...), append([]uint64{}, f.nacked...), append([]bool{}, f.requeued...)
}

type fakeTx struct {
	pgx.Tx
	commitErr  error
	committed  bool
	rolledBack bool
}

func (f *fakeTx) Commit(context.Context) error {
	if f.commitErr != nil {
		return f.commitErr
	}
	f.committed = true
	return nil
}

func (f *fakeTx) Rollback(context.Context) error { f.rolledBack = true; return nil }

type fakeTransactor struct {
	beginErr error
	tx       *fakeTx
}

func (f *fakeTransactor) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return f.tx, nil
}

type fakeInboxRepo struct {
	processed map[string]bool
}

func (f *fakeInboxRepo) Create(ctx context.Context, tx pgx.Tx, eventID string) error {
	if f.processed[eventID] {
		return shareddomain.ErrConflict
	}
	f.processed[eventID] = true
	return nil
}

func newDelivery(messageID string, ack *fakeAcknowledger) amqp.Delivery {
	return amqp.Delivery{
		Acknowledger: ack,
		DeliveryTag:  1,
		MessageId:    messageID,
		Body:         []byte(`{"event_id":"` + messageID + `"}`),
	}
}

// consumerRun runs one consumer in a goroutine and stops it at the end.
type consumerRun struct {
	cancel context.CancelFunc
	done   chan error
}

func startConsumer(t *testing.T, sub *fakeSubscriber, transactor db.Transactor, inboxRepo inbox.InboxRepository, handler consumer.HandlerFunc) *consumerRun {
	t.Helper()
	c := consumer.New(sub, transactor, inboxRepo, "test.queue", handler)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Start(ctx) }()
	return &consumerRun{cancel: cancel, done: done}
}

func (r *consumerRun) finish(t *testing.T) {
	t.Helper()
	r.cancel()
	select {
	case err := <-r.done:
		if err != nil {
			t.Fatalf("consumer start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("consumer did not stop after cancel")
	}
}

// --- tests ---

func TestConsumer(t *testing.T) {
	cases := []struct {
		name        string
		handlerErr  error
		beginErr    error
		commitErr   error
		messageID   string
		preProcess  bool // eventID already in inbox → duplicate
		wantAcked   bool
		wantNacked  bool
		wantRequeue bool
		wantHandler bool
	}{
		{
			name:        "success",
			messageID:   "evt-1",
			wantAcked:   true,
			wantHandler: true,
		},
		{
			name:        "duplicate event skipped and acked",
			messageID:   "evt-dup",
			preProcess:  true,
			wantAcked:   true,
			wantHandler: false,
		},
		{
			name:        "handler error routes to DLQ",
			messageID:   "evt-bad",
			handlerErr:  errors.New("business rule violated"),
			wantNacked:  true,
			wantRequeue: false,
			wantHandler: true,
		},
		{
			name:        "begin tx error retries",
			messageID:   "evt-begin",
			beginErr:    errors.New("connection down"),
			wantNacked:  true,
			wantRequeue: true,
			wantHandler: false,
		},
		{
			name:        "commit error retries",
			messageID:   "evt-commit",
			commitErr:   errors.New("db unavailable"),
			wantNacked:  true,
			wantRequeue: true,
			wantHandler: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ack := &fakeAcknowledger{}
			sub := &fakeSubscriber{deliveries: make(chan amqp.Delivery, 1)}
			tx := &fakeTx{commitErr: tc.commitErr}
			transactor := &fakeTransactor{tx: tx, beginErr: tc.beginErr}
			inboxRepo := &fakeInboxRepo{processed: map[string]bool{}}
			if tc.preProcess {
				inboxRepo.processed[tc.messageID] = true
			}

			handlerCalls := 0
			handler := func(ctx context.Context, tx pgx.Tx, payload []byte) error {
				handlerCalls++
				return tc.handlerErr
			}

			run := startConsumer(t, sub, transactor, inboxRepo, handler)
			sub.deliveries <- newDelivery(tc.messageID, ack)
			// give the loop time to process before teardown
			time.Sleep(50 * time.Millisecond)
			run.finish(t)

			acked, nacked, requeued := ack.calls()

			if tc.wantAcked && len(acked) != 1 {
				t.Errorf("expected 1 ack, got %d (nacked=%v)", len(acked), nacked)
			}
			if !tc.wantAcked && len(acked) != 0 {
				t.Errorf("expected 0 acks, got %d", len(acked))
			}
			if tc.wantNacked {
				if len(nacked) != 1 {
					t.Fatalf("expected 1 nack, got %d", len(nacked))
				}
				if len(requeued) != 1 || requeued[0] != tc.wantRequeue {
					t.Errorf("expected requeue=%v, got %v", tc.wantRequeue, requeued)
				}
			} else if len(nacked) != 0 {
				t.Errorf("expected 0 nacks, got %d", len(nacked))
			}

			if tc.wantHandler && handlerCalls != 1 {
				t.Errorf("expected handler called once, got %d", handlerCalls)
			}
			if !tc.wantHandler && handlerCalls != 0 {
				t.Errorf("expected handler not called, got %d calls", handlerCalls)
			}
		})
	}
}

func TestConsumerHandlerReceivesPayloadAndCommits(t *testing.T) {
	ack := &fakeAcknowledger{}
	sub := &fakeSubscriber{deliveries: make(chan amqp.Delivery, 1)}
	tx := &fakeTx{}
	transactor := &fakeTransactor{tx: tx}
	inboxRepo := &fakeInboxRepo{processed: map[string]bool{}}

	var gotPayload []byte
	handler := func(ctx context.Context, tx pgx.Tx, payload []byte) error {
		gotPayload = append([]byte{}, payload...)
		return nil
	}

	run := startConsumer(t, sub, transactor, inboxRepo, handler)
	sub.deliveries <- amqp.Delivery{
		Acknowledger: ack,
		DeliveryTag:  1,
		MessageId:    "evt-payload",
		Body:         []byte(`{"hello":"world"}`),
	}
	time.Sleep(50 * time.Millisecond)
	run.finish(t)

	if string(gotPayload) != `{"hello":"world"}` {
		t.Errorf("handler got payload %q, want %q", gotPayload, `{"hello":"world"}`)
	}
	if !tx.committed {
		t.Error("expected transaction to be committed on success")
	}
	acked, _, _ := ack.calls()
	if len(acked) != 1 {
		t.Errorf("expected message to be acked, got %d acks", len(acked))
	}
}
