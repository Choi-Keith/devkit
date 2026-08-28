package userdb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/Choi-Keith/golang-common/database"

	"github.com/Choi-Keith/golang-ddd-template/internal/domain/user"
)

// MysqlUserRepo implements user.Repository with GORM.
type MysqlUserRepo struct {
	db *gorm.DB
}

// NewMysqlUserRepo creates a MysqlUserRepo.
func NewMysqlUserRepo(db *gorm.DB) *MysqlUserRepo {
	return &MysqlUserRepo{db: db}
}

// Save persists a new user and fills u.ID with the auto-increment value.
func (r *MysqlUserRepo) Save(ctx context.Context, u *user.User) error {
	po := toPO(u)
	err := database.WithContext(ctx, r.db).Create(po).Error
	switch {
	case err == nil:
		u.ID = po.ID
		return nil
	case isDuplicateKey(err):
		return user.ErrEmailTaken
	default:
		return fmt.Errorf("MysqlUserRepo.Save: %w", err)
	}
}

// FindByID returns the user with the given id, or user.ErrNotFound.
func (r *MysqlUserRepo) FindByID(ctx context.Context, id uint64) (*user.User, error) {
	var po UserPO
	err := database.WithContext(ctx, r.db).Where("id = ?", id).First(&po).Error
	if err != nil {
		return nil, mapReadError("FindByID", err)
	}
	return toDomain(&po), nil
}

// FindByEmail returns the user with the given email, or user.ErrNotFound.
func (r *MysqlUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var po UserPO
	err := database.WithContext(ctx, r.db).Where("email = ?", email).First(&po).Error
	if err != nil {
		return nil, mapReadError("FindByEmail", err)
	}
	return toDomain(&po), nil
}

// List returns a page of users ordered by creation time (newest first).
func (r *MysqlUserRepo) List(ctx context.Context, page, pageSize int) ([]*user.User, int64, error) {
	var (
		pos   []UserPO
		total int64
	)

	q := database.WithContext(ctx, r.db).Model(&UserPO{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("MysqlUserRepo.List count: %w", err)
	}

	err := q.Scopes(
		database.Paginate(page, pageSize),
		database.OrderBy("created_at", true),
	).Find(&pos).Error
	if err != nil {
		return nil, 0, fmt.Errorf("MysqlUserRepo.List find: %w", err)
	}

	users := make([]*user.User, 0, len(pos))
	for i := range pos {
		users = append(users, toDomain(&pos[i]))
	}
	return users, total, nil
}

// Delete removes the user with the given id, or returns user.ErrNotFound.
func (r *MysqlUserRepo) Delete(ctx context.Context, id uint64) error {
	res := database.WithContext(ctx, r.db).Where("id = ?", id).Delete(&UserPO{})
	if res.Error != nil {
		return fmt.Errorf("MysqlUserRepo.Delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return user.ErrNotFound
	}
	return nil
}

func mapReadError(op string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user.ErrNotFound
	}
	return fmt.Errorf("MysqlUserRepo.%s: %w", op, err)
}

// isDuplicateKey reports whether err is a unique-constraint violation
// (MySQL 1062).
func isDuplicateKey(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "1062") || strings.Contains(msg, "Duplicate entry")
}
