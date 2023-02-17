package model

import (
	"time"
)

type User struct {
	UserID        uint64    `gorm:"column:id;primary_key;NOT NULL;autoIncrement" redis:"user_id"`
	Name          string    `gorm:"column:name;NOT NULL" redis:"name"`
	Password      string    `gorm:"column:password;NOT NULL" redis:"password"`
	FollowCount   int       `gorm:"column:follow_count;NOT NULL" redis:"follow_count"`
	FollowerCount int       `gorm:"column:follower_count;NOT NULL" redis:"follower_count"`
	IsFollower    int       `gorm:"column:is_follower;NOT NULL" redis:"favorite_count"`
	WorkCount     int       `gorm:"column:work_count;NOT NULL" redis:"favorite_count"`
	FavoriteCount int       `gorm:"column:favorite_count;NOT NULL" redis:"favorite_count"`
	CreatedAt     time.Time `gorm:"column:created_at" redis:"-"`
	ExtInfo       *string   `gorm:"column:ext_info" redis:"-"`
}
