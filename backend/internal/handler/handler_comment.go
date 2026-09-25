package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gbtreehole/backend/internal/constants"
	"github.com/gbtreehole/backend/internal/dto"
	"github.com/gbtreehole/backend/internal/model"
	"github.com/gbtreehole/backend/internal/service"
)

type CommentHandler struct {
	comments service.CommentService
	likes    service.LikeService
	aliases  service.AliasService
	logger   *slog.Logger
}

func NewCommentHandler(comments service.CommentService, likes service.LikeService, aliases service.AliasService, logger *slog.Logger) *CommentHandler {
	return &CommentHandler{comments: comments, likes: likes, aliases: aliases, logger: logger}
}

// CreateComment 发表评论
// @Summary 发表评论
// @Tags comment
// @Accept json
// @Produce json
// @Param request body dto.CreateCommentRequest true "评论内容"
// @Success 200 {object} Response
// @Router /api/v1/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	identityID := c.GetUint("identityId")
	var req dto.CreateCommentRequest
	if !BindAndValidate(c, &req) {
		return
	}
	comment, hits, blocked, err := h.comments.Create(identityID, req.PostID, req.Content)
	if err != nil {
		h.logger.Error("create comment", "error", err)
		Fail(c, http.StatusInternalServerError, constants.CodeInternal, "create comment failed")
		return
	}
	OK(c, gin.H{"comment": h.toCommentResponse(comment, false), "blocked": blocked, "hitWords": hits})
}

// ListComments 帖子评论列表
// @Summary 评论列表
// @Tags comment
// @Produce json
// @Param id path int true "帖子ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} Response
// @Router /api/v1/posts/{id}/comments [get]
func (h *CommentHandler) ListComments(c *gin.Context) {
	postID := parseID(c)
	if postID == 0 {
		return
	}
	var req dto.ListCommentRequest
	if !BindQuery(c, &req) {
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	comments, total, err := h.comments.ListByPostID(postID, req.Page, req.PageSize)
	if err != nil {
		h.logger.Error("list comments", "error", err)
		Fail(c, http.StatusInternalServerError, constants.CodeInternal, "list comments failed")
		return
	}
	ids := make([]uint, 0, len(comments))
	for _, cm := range comments {
		ids = append(ids, cm.ID)
	}
	likedMap := map[uint]bool{}
	if identityID := c.GetUint("identityId"); identityID > 0 {
		if m, err := h.likes.IsLiked(identityID, "comment", ids); err == nil {
			likedMap = m
		}
	}
	items := make([]dto.CommentResponse, 0, len(comments))
	for _, cm := range comments {
		items = append(items, h.toCommentResponse(&cm, likedMap[cm.ID]))
	}
	OK(c, dto.PageResult{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize})
}

// toCommentResponse 组装评论对外响应：昵称与头像使用该身份在所属帖子内的树洞化名，
// 与帖子楼主及同帖其他评论保持同一套派生规则，且不暴露真实身份编号。
func (h *CommentHandler) toCommentResponse(comment *model.Comment, liked bool) dto.CommentResponse {
	alias := h.aliases.For(comment.PostID, comment.IdentityID)
	return dto.CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		Nickname:  alias.Nickname,
		Avatar:    alias.Avatar,
		Content:   comment.Content,
		LikeCount: comment.LikeCount,
		Liked:     liked,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}
}
