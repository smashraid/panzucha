package consumers_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"panzucha/internal/order/consumers"
)

// fakeTx satisfies pgx.Tx — the no-op handler never touches it.
type fakeTx struct {
	pgx.Tx
}

func TestHandleOrderCreated(t *testing.T) {
	valid := `{"event_id":"evt-1","event_type":"order.created","timestamp":"2026-01-01T00:00:00Z","order_id":"ord-1","user_id":"usr-1","items":[],"total_amount":9.99}`

	cases := []struct {
		name    string
		payload string
		wantErr bool
	}{
		{name: "valid event returns nil", payload: valid},
		{name: "malformed json routes to DLQ", payload: `{"order_id":`, wantErr: true},
		{name: "empty payload routes to DLQ", payload: ``, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := consumers.HandleOrderCreated(context.Background(), &fakeTx{}, []byte(tc.payload))
			if tc.wantErr && err == nil {
				t.Errorf("expected error for payload %q, got nil", tc.payload)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected nil error for payload %q, got %v", tc.payload, err)
			}
		})
	}
}
