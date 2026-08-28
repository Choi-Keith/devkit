package order

import "context"

// Repository defines the persistence contract for the order aggregate.
// Implementations live in the infra layer and are injected at runtime.
type Repository interface {
	// Save persists a new order and fills its auto-increment ID.
	Save(ctx context.Context, o *Order) error
	// FindByID returns the order with the given id, or ErrNotFound.
	FindByID(ctx context.Context, id uint64) (*Order, error)
	// List returns a page of orders and the total count.
	List(ctx context.Context, page, pageSize int) ([]*Order, int64, error)
	// Update persists the changes of an existing order, or returns
	// ErrNotFound.
	Update(ctx context.Context, o *Order) error
}
