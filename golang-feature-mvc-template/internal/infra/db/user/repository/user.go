// Package repository 用户模块的 GORM 数据访问实现。
// 接口定义在消费方(internal/user/service),本包只提供具体实现,
// 由 main 手工装配;错误语义遵循接口约定:不存在返回
// gorm.ErrRecordNotFound,唯一键冲突返回 gorm.ErrDuplicatedKey。
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/Choi-Keith/golang-common/database"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/infra/db/user/model"
)

// Repository 基于 GORM 的用户仓储。
type Repository struct {
	db *gorm.DB
}

// New 创建用户仓储。
func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 插入新用户并把自增 ID 回填到 u.ID;唯一键冲突返回 gorm.ErrDuplicatedKey。
func (r *Repository) Create(ctx context.Context, u *model.User) error {
	err := database.WithContext(ctx, r.db).Create(u).Error
	switch {
	case err == nil:
		return nil
	case isDuplicateKey(err):
		return gorm.ErrDuplicatedKey
	default:
		return fmt.Errorf("user.repository.Create: %w", err)
	}
}

// FindByID 按主键查询,不存在时返回 gorm.ErrRecordNotFound。
func (r *Repository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var u model.User
	err := database.WithContext(ctx, r.db).Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, mapReadError("FindByID", err)
	}
	return &u, nil
}

// FindByEmail 按邮箱查询,不存在时返回 gorm.ErrRecordNotFound。
func (r *Repository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := database.WithContext(ctx, r.db).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, mapReadError("FindByEmail", err)
	}
	return &u, nil
}

// List 分页查询,按创建时间倒序。
func (r *Repository) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	var (
		users []*model.User
		total int64
	)

	q := database.WithContext(ctx, r.db).Model(&model.User{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("user.repository.List count: %w", err)
	}

	err := q.Scopes(
		database.Paginate(page, pageSize),
		database.OrderBy("created_at", true),
	).Find(&users).Error
	if err != nil {
		return nil, 0, fmt.Errorf("user.repository.List find: %w", err)
	}
	return users, total, nil
}

// Update 全量保存用户,不存在时返回 gorm.ErrRecordNotFound。
func (r *Repository) Update(ctx context.Context, u *model.User) error {
	res := database.WithContext(ctx, r.db).Save(u)
	if res.Error != nil {
		return fmt.Errorf("user.repository.Update: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 按主键删除,不存在时返回 gorm.ErrRecordNotFound。
func (r *Repository) Delete(ctx context.Context, id uint64) error {
	res := database.WithContext(ctx, r.db).Where("id = ?", id).Delete(&model.User{})
	if res.Error != nil {
		return fmt.Errorf("user.repository.Delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func mapReadError(op string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return gorm.ErrRecordNotFound
	}
	return fmt.Errorf("user.repository.%s: %w", op, err)
}

// isDuplicateKey 判断是否为唯一键冲突(MySQL 1062 / PostgreSQL 23505)。
func isDuplicateKey(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "1062") || strings.Contains(msg, "Duplicate entry") ||
		strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key")
}
