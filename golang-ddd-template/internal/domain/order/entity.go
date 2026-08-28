// Package order contains the order aggregate: the entity, its domain
// invariants, and the persistence contract. The domain layer has zero
// third-party dependencies.
package order

import (
	"strings"
	"time"
)

// Status is the lifecycle state of an order.
type Status string

const (
	StatusPending   Status = "pending"
	StatusPaid      Status = "paid"
	StatusCancelled Status = "cancelled"
)

// Order is the aggregate root of the order domain.
type Order struct {
	ID        uint64 // auto-increment, assigned by the database on Save
	UserID    uint64
	OrderNo   string // unique business number
	Amount    int64  // amount in cents, always positive
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewOrder creates an Order in the pending state after enforcing the
// domain invariants. The ID is assigned by the database (auto-increment)
// when the order is persisted.
func NewOrder(userID uint64, orderNo string, amount int64) (*Order, error) {
	if userID == 0 {
		return nil, ErrUserIDRequired
	}
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return nil, ErrOrderNoRequired
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	now := time.Now()
	return &Order{
		UserID:    userID,
		OrderNo:   orderNo,
		Amount:    amount,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Cancel transitions a pending order to cancelled. Only pending orders
// can be cancelled.
func (o *Order) Cancel() error {
	if o.Status != StatusPending {
		return ErrNotCancellable
	}
	o.Status = StatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}
