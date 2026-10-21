package model

import "time"

// User 用户模型,对应 users 表。
type User struct {
	ID        uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement"`
	Email     string    `gorm:"uniqueIndex;size:255;not null"`
	Name      string    `gorm:"size:128;not null"`
	Password  string    `gorm:"size:255;not null"` // bcrypt 哈希,永不返回给客户端
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName 指定表名。
func (User) TableName() string { return "users" }
