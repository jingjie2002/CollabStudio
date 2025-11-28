package models

import (
	"gorm.io/gorm"
)

type Document struct {
	gorm.Model
	// 确保 RoomID 是唯一的，这样我们才能准确找到它
	RoomID  string `gorm:"uniqueIndex;size:100;not null" json:"room_id"`
	Content string `gorm:"type:text" json:"content"`
}
