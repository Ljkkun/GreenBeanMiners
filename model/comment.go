package model

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	CommentID uint64         `gorm:"column:id;primary_key;NOT NULL" redis:"-"`
	User      uint64         `gorm:"column:user;NOT NULL" redis:"user_id"`
	Content   string         `gorm:"column:content;NOT NULL" redis:"content"`
	CreatedAt time.Time      `gorm:"column:created_at" redis:"-"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" redis:"-"`
}
