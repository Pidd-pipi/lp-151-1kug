package model

import "time"

type Comment struct {
	ID         uint          `gorm:"primaryKey" json:"id"`
	PostID     uint          `gorm:"index;not null" json:"postId"`
	Post       *Post         `gorm:"foreignKey:PostID" json:"-"`
	IdentityID uint          `gorm:"index;not null" json:"-"`
	Identity   *UserIdentity `gorm:"foreignKey:IdentityID" json:"-"`
	Content    string        `gorm:"type:text;not null" json:"content"`
	Status     int           `gorm:"default:1;index" json:"status"`
	LikeCount  int           `gorm:"default:0" json:"likeCount"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
}
