package orderapp

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Choi-Keith/golang-ddd-template/internal/domain/order"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// OrderService orchestrates the order use cases.
type OrderService struct {
	repo order.Repository
}

// NewOrderService creates an OrderService backed by the given repository.
func NewOrderService(repo order.Repository) *OrderService {
	return &OrderService{repo: repo}
}

// Create creates a new pending order. The order number is generated here:
// timestamp for ordering plus random digits against collisions.
func (s *OrderService) Create(ctx context.Context, in CreateInput) (*OrderDTO, error) {
	orderNo := fmt.Sprintf("O%d%04d", time.Now().UnixNano(), rand.IntN(10000))

	o, err := order.NewOrder(in.UserID, orderNo, in.Amount)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, fmt.Errorf("orderapp.Create: save order: %w", err)
	}
	return toDTO(o), nil
}

// GetByID returns a single order.
func (s *OrderService) GetByID(ctx context.Context, id uint64) (*OrderDTO, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("orderapp.GetByID: %w", err)
	}
	return toDTO(o), nil
}

// List returns a page of orders.
func (s *OrderService) List(ctx context.Context, page, pageSize int) (*OrderListDTO, error) {
	page, pageSize = normalizePage(page, pageSize)

	orders, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("orderapp.List: %w", err)
	}

	list := make([]OrderDTO, 0, len(orders))
	for _, o := range orders {
		list = append(list, *toDTO(o))
	}
	return &OrderListDTO{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// Cancel cancels a pending order.
func (s *OrderService) Cancel(ctx context.Context, id uint64) (*OrderDTO, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("orderapp.Cancel: %w", err)
	}
	if err := o.Cancel(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, o); err != nil {
		return nil, fmt.Errorf("orderapp.Cancel: update order: %w", err)
	}
	return toDTO(o), nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func toDTO(o *order.Order) *OrderDTO {
	return &OrderDTO{
		ID:        o.ID,
		UserID:    o.UserID,
		OrderNo:   o.OrderNo,
		Amount:    o.Amount,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
	}
}
