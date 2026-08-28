package order

import (
	"errors"
	"testing"
)

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name    string
		userID  uint64
		orderNo string
		amount  int64
		wantErr error
	}{
		{name: "ok", userID: 1, orderNo: "O123", amount: 199900},
		{name: "user id required", userID: 0, orderNo: "O123", amount: 199900, wantErr: ErrUserIDRequired},
		{name: "order no required", userID: 1, orderNo: "  ", amount: 199900, wantErr: ErrOrderNoRequired},
		{name: "amount zero", userID: 1, orderNo: "O123", amount: 0, wantErr: ErrInvalidAmount},
		{name: "amount negative", userID: 1, orderNo: "O123", amount: -1, wantErr: ErrInvalidAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := NewOrder(tt.userID, tt.orderNo, tt.amount)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if o.ID != 0 {
				t.Fatalf("id must be zero before persistence, got %d", o.ID)
			}
			if o.Status != StatusPending {
				t.Fatalf("new order must be pending, got %s", o.Status)
			}
			if o.UserID != tt.userID || o.Amount != tt.amount {
				t.Fatalf("unexpected order: %+v", o)
			}
			if o.CreatedAt.IsZero() || o.UpdatedAt.IsZero() {
				t.Fatal("timestamps must be set")
			}
		})
	}
}

func TestOrder_Cancel(t *testing.T) {
	o, err := NewOrder(1, "O123", 199900)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := o.Cancel(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Status != StatusCancelled {
		t.Fatalf("want status %s, got %s", StatusCancelled, o.Status)
	}

	// A cancelled order cannot be cancelled again.
	if err := o.Cancel(); !errors.Is(err, ErrNotCancellable) {
		t.Fatalf("want ErrNotCancellable, got %v", err)
	}
}
