package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model        // 自动包含 ID, CreatedAt, UpdatedAt 等字段
	Username   string `gorm:"uniqueIndex;not null" json:"username"`
	Password   string `gorm:"not null" json:"-"` // json:"-" 表示返回给前端时不带密码
	Avatar     string `json:"avatar"`
	Role       string `gorm:"default:'user'" json:"role"` // user 或 admin
}
