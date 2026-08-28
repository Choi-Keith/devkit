// Package userdb holds the GORM-backed repository implementations.
package userdb

import (
	"time"

	"github.com/Choi-Keith/golang-ddd-template/internal/domain/user"
)

// UserPO is the persistence object for the users table. It is an infra
// concern and never leaks into the domain layer.
type UserPO struct {
	ID        uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement"`
	Email     string    `gorm:"uniqueIndex;size:255;not null"`
	Name      string    `gorm:"size:128;not null"`
	Password  string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName implements gorm's tabler interface.
func (UserPO) TableName() string { return "users" }

func toPO(u *user.User) *UserPO {
	return &UserPO{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toDomain(po *UserPO) *user.User {
	return &user.User{
		ID:        po.ID,
		Email:     po.Email,
		Name:      po.Name,
		Password:  po.Password,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}
