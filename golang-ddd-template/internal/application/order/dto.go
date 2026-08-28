// Package orderapp orchestrates the order use cases: creation, queries
// and cancellation. It depends only on domain contracts and stays free of
// transport or storage details.
package orderapp

// CreateInput carries the data needed to create an order.
type CreateInput struct {
	UserID uint64
	Amount int64 // amount in cents
}

// OrderDTO is the application-layer view of an order.
type OrderDTO struct {
	ID        uint64
	UserID    uint64
	OrderNo   string
	Amount    int64
	Status    string
	CreatedAt string
}

// OrderListDTO is a page of orders.
type OrderListDTO struct {
	List     []OrderDTO
	Total    int64
	Page     int
	PageSize int
}
