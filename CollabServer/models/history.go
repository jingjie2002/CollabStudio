package models

import "time"

type History struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"index" json:"username"` // 索引，加速查询
	RoomID    string    `json:"room_id"`
	UpdatedAt time.Time `json:"last_visited"`
}
