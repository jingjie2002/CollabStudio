package models

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	RoomID  string `gorm:"index;size:100;not null" json:"room_id"` // ✅ 必须有 json:"room_id"
	Sender  string `gorm:"size:100" json:"sender"`                 // ✅ 必须有 json:"sender"
	Content string `gorm:"type:text" json:"content"`               // ✅ 必须有 json:"content"
}
