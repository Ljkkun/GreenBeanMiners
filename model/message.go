package model

import (
	"time"
)

type Message struct {
	FollowID   uint64    `gorm:"column:id;primary_key;NOT NULL" redis:"-"`
	ToUserID   uint64    `gorm:"column:to_user_id;NOT NULL" redis:"to_user_id"`
	FromUserID uint64    `gorm:"column:from_user_id;NOT NULL" redis:"from_user_id"`
	Content    string    `gorm:"column:content;NOT NULL" redis:"content"`
	CreatedAt  time.Time `gorm:"column:id;create_at;NOT NULL" redis:"-"`
	UpdatedAt  time.Time `gorm:"column:id;update_at;NOT NULL" redis:"-"`
}
