// Package repository 订单模块的 GORM 数据访问实现。
// 接口定义在消费方(internal/order/service),本包只提供具体实现,
// 由 main 手工装配;查询类方法在记录不存在时返回 gorm.ErrRecordNotFound。
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/Choi-Keith/golang-common/database"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/infra/db/order/model"
)

// Repository 基于 GORM 的订单仓储。
type Repository struct {
	db *gorm.DB
}

// New 创建订单仓储。
func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 插入新订单并把自增 ID 回填到 o.ID。
func (r *Repository) Create(ctx context.Context, o *model.Order) error {
	if err := database.WithContext(ctx, r.db).Create(o).Error; err != nil {
		return fmt.Errorf("order.repository.Create: %w", err)
	}
	return nil
}

// FindByID 按主键查询,不存在时返回 gorm.ErrRecordNotFound。
func (r *Repository) FindByID(ctx context.Context, id uint64) (*model.Order, error) {
	var o model.Order
	err := database.WithContext(ctx, r.db).Where("id = ?", id).First(&o).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("order.repository.FindByID: %w", err)
	}
	return &o, nil
}

// List 分页查询,按创建时间倒序。
func (r *Repository) List(ctx context.Context, page, pageSize int) ([]*model.Order, int64, error) {
	var (
		orders []*model.Order
		total  int64
	)

	q := database.WithContext(ctx, r.db).Model(&model.Order{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("order.repository.List count: %w", err)
	}

	err := q.Scopes(
		database.Paginate(page, pageSize),
		database.OrderBy("created_at", true),
	).Find(&orders).Error
	if err != nil {
		return nil, 0, fmt.Errorf("order.repository.List find: %w", err)
	}
	return orders, total, nil
}

// Update 全量保存订单,不存在时返回 gorm.ErrRecordNotFound。
func (r *Repository) Update(ctx context.Context, o *model.Order) error {
	res := database.WithContext(ctx, r.db).Save(o)
	if res.Error != nil {
		return fmt.Errorf("order.repository.Update: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
