package model

import "time"

type Video struct {
	VideoID       uint64    `gorm:"column:id;primary_key;NOT NULL" redis:"-"`
	User          int       `gorm:"column:user;NOT NULL" redis:"title"`
	PlayUrl       string    `gorm:"column:play_url;NOT NULL" redis:"play_name"`
	CoverUrl      string    `gorm:"column:url;NOT NULL" redis:"cover_name"`
	FavoriteCount int64     `gorm:"-" redis:"favorite_count"`
	CommentCount  int64     `gorm:"-" redis:"comment_count"`
	IsFavorite    uint64    `gorm:"column:is_favorite;NOT NULL" redis:"author_id"`
	CreatedAt     time.Time `gorm:"column:created_at" redis:"-"`
	ExtInfo       *string   `gorm:"column:ext_info" redis:"-"`
}

type VideoCount struct {
	VideoID      uint64
	CommentCount int64
}
