// Package model 订单模块的 GORM 模型。本模板有意合并 model/PO/domain
// (贫血模型直通),状态迁移等业务规则由 service 判定。
package model

import "time"

// 订单状态。
const (
	StatusPending   = "pending"
	StatusCancelled = "cancelled"
)

// Order 订单模型,对应 orders 表。金额单位为分。
type Order struct {
	ID        uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"index;not null"`
	OrderNo   string    `gorm:"uniqueIndex;size:64;not null"`
	Amount    int64     `gorm:"not null"`
	Status    string    `gorm:"size:16;not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName 指定表名。
func (Order) TableName() string { return "orders" }
