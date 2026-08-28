package orderdb

import (
	"time"

	"github.com/Choi-Keith/golang-ddd-template/internal/domain/order"
)

// OrderPO is the persistence object for the orders table. It is an infra
// concern and never leaks into the domain layer.
type OrderPO struct {
	ID        uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"column:user_id;type:bigint unsigned;not null;index"`
	OrderNo   string    `gorm:"column:order_no;size:64;uniqueIndex;not null"`
	Amount    int64     `gorm:"column:amount;not null"`
	Status    string    `gorm:"column:status;size:16;not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName implements gorm's tabler interface.
func (OrderPO) TableName() string { return "orders" }

func toOrderPO(o *order.Order) *OrderPO {
	return &OrderPO{
		ID:        o.ID,
		UserID:    o.UserID,
		OrderNo:   o.OrderNo,
		Amount:    o.Amount,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func orderToDomain(po *OrderPO) *order.Order {
	return &order.Order{
		ID:        po.ID,
		UserID:    po.UserID,
		OrderNo:   po.OrderNo,
		Amount:    po.Amount,
		Status:    order.Status(po.Status),
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}
