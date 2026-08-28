// Package model 定义 GORM 数据模型,与数据库表结构对应。
package model

import "time"

// User 用户模型
type User struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Email     string    `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	Name      string    `gorm:"type:varchar(64);not null" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
