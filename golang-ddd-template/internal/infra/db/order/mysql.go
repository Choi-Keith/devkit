package orderdb

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/Choi-Keith/golang-common/database"

	"github.com/Choi-Keith/golang-ddd-template/internal/domain/order"
)

// MysqlOrderRepo implements order.Repository with GORM.
type MysqlOrderRepo struct {
	db *gorm.DB
}

// NewMysqlOrderRepo creates a MysqlOrderRepo.
func NewMysqlOrderRepo(db *gorm.DB) *MysqlOrderRepo {
	return &MysqlOrderRepo{db: db}
}

// Save persists a new order and fills o.ID with the auto-increment value.
func (r *MysqlOrderRepo) Save(ctx context.Context, o *order.Order) error {
	po := toOrderPO(o)
	err := database.WithContext(ctx, r.db).Create(po).Error
	if err != nil {
		return fmt.Errorf("MysqlOrderRepo.Save: %w", err)
	}
	o.ID = po.ID
	return nil
}

// FindByID returns the order with the given id, or order.ErrNotFound.
func (r *MysqlOrderRepo) FindByID(ctx context.Context, id uint64) (*order.Order, error) {
	var po OrderPO
	err := database.WithContext(ctx, r.db).Where("id = ?", id).First(&po).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, order.ErrNotFound
		}
		return nil, fmt.Errorf("MysqlOrderRepo.FindByID: %w", err)
	}
	return orderToDomain(&po), nil
}

// List returns a page of orders ordered by creation time (newest first).
func (r *MysqlOrderRepo) List(ctx context.Context, page, pageSize int) ([]*order.Order, int64, error) {
	var (
		pos   []OrderPO
		total int64
	)

	q := database.WithContext(ctx, r.db).Model(&OrderPO{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("MysqlOrderRepo.List count: %w", err)
	}

	err := q.Scopes(
		database.Paginate(page, pageSize),
		database.OrderBy("created_at", true),
	).Find(&pos).Error
	if err != nil {
		return nil, 0, fmt.Errorf("MysqlOrderRepo.List find: %w", err)
	}

	orders := make([]*order.Order, 0, len(pos))
	for i := range pos {
		orders = append(orders, orderToDomain(&pos[i]))
	}
	return orders, total, nil
}

// Update persists the changes of an existing order, or returns
// order.ErrNotFound.
func (r *MysqlOrderRepo) Update(ctx context.Context, o *order.Order) error {
	res := database.WithContext(ctx, r.db).Model(&OrderPO{}).
		Where("id = ?", o.ID).
		Updates(map[string]any{
			"status":     string(o.Status),
			"updated_at": o.UpdatedAt,
		})
	if res.Error != nil {
		return fmt.Errorf("MysqlOrderRepo.Update: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return order.ErrNotFound
	}
	return nil
}
