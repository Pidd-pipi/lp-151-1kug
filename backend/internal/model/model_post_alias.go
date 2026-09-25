package model

import "time"

// PostAlias 记录某个匿名身份在单篇帖子下的树洞化名。
// 同一身份在同一帖子（发帖 + 评论）内复用同一组化名与头像，
// 换一篇帖子则生成另一组，避免读者跨帖拼接同一身份的发言。
type PostAlias struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PostID     uint      `gorm:"uniqueIndex:uk_post_alias_post_identity;not null" json:"postId"`
	IdentityID uint      `gorm:"uniqueIndex:uk_post_alias_post_identity;not null" json:"identityId"`
	Nickname   string    `gorm:"type:varchar(64);not null" json:"nickname"`
	Avatar     string    `gorm:"type:varchar(255)" json:"avatar"`
	CreatedAt  time.Time `json:"createdAt"`
}
