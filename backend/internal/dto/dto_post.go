package dto

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"omitempty,max=255"`
	Content string   `json:"content" validate:"required,min=1,max=5000"`
	Images  []string `json:"images" validate:"omitempty,max=9,dive,max=500"`
	Tags    []string `json:"tags" validate:"omitempty,max=10,dive,max=32"`
}

type ListPostRequest struct {
	Page     int  `json:"page" form:"page" validate:"omitempty,min=1"`
	PageSize int  `json:"pageSize" form:"page_size" validate:"omitempty,min=1,max=100"`
	TagID    uint `json:"tagId" form:"tag_id"`
	Featured bool `json:"featured" form:"featured"`
}

// PostResponse 帖子对外视图。昵称与头像为该帖专属的树洞化名，
// 不携带身份编号，避免读者跨帖拼接同一身份的发言。
type PostResponse struct {
	ID           uint          `json:"id"`
	Nickname     string        `json:"nickname"`
	Avatar       string        `json:"avatar"`
	Title        string        `json:"title"`
	Content      string        `json:"content"`
	Images       []string      `json:"images"`
	Status       int           `json:"status"`
	LikeCount    int           `json:"likeCount"`
	CommentCount int           `json:"commentCount"`
	ViewCount    int           `json:"viewCount"`
	IsFeatured   bool          `json:"isFeatured"`
	Liked        bool          `json:"liked"`
	Tags         []TagResponse `json:"tags"`
	CreatedAt    string        `json:"createdAt"`
}

type PageResult struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}
