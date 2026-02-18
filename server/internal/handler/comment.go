// Package handler 提供 HTTP 处理器
package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/middleware"
	"github.com/liwei0526vip/mylinear/internal/service"
)

// CommentHandler 评论处理器
type CommentHandler struct {
	commentService service.CommentService
}

// NewCommentHandler 创建评论处理器
func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// CreateComment 创建评论
// POST /api/v1/issues/:id/comments
func (h *CommentHandler) CreateComment(c *gin.Context) {
	// 获取用户 ID
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 解析 Issue ID
	issueIDStr := c.Param("id")
	if issueIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	issueID, err := uuid.Parse(issueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 Issue ID"})
		return
	}

	// 解析请求体
	var req struct {
		Body     string `json:"body" binding:"required"`
		ParentID string `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	if req.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
		return
	}

	// 解析父评论 ID
	var parentID *uuid.UUID
	if req.ParentID != "" {
		id, err := uuid.Parse(req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的父评论 ID"})
			return
		}
		parentID = &id
	}

	ctx := c.Request.Context()

	comment, err := h.commentService.CreateComment(ctx, issueID, userID, req.Body, parentID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         comment.ID,
		"issue_id":   comment.IssueID,
		"parent_id":  comment.ParentID,
		"user_id":    comment.UserID,
		"body":       comment.Body,
		"created_at": comment.CreatedAt,
	})
}

// ListIssueComments 获取 Issue 的评论列表
// GET /api/v1/issues/:id/comments
func (h *CommentHandler) ListIssueComments(c *gin.Context) {
	issueIDStr := c.Param("id")
	if issueIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	issueID, err := uuid.Parse(issueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 Issue ID"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	ctx := c.Request.Context()

	comments, total, err := h.commentService.GetCommentsByIssueID(ctx, issueID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	result := make([]gin.H, len(comments))
	for i, comment := range comments {
		result[i] = gin.H{
			"id":         comment.ID,
			"issue_id":   comment.IssueID,
			"parent_id":  comment.ParentID,
			"user_id":    comment.UserID,
			"body":       comment.Body,
			"created_at": comment.CreatedAt,
			"updated_at": comment.UpdatedAt,
			"edited_at":  comment.EditedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"comments":  result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateComment 更新评论
// PUT /api/v1/comments/:commentId
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	// 获取用户 ID
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	commentIDStr := c.Param("commentId")
	if commentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少评论 ID"})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评论 ID"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	if req.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
		return
	}

	ctx := c.Request.Context()

	comment, err := h.commentService.UpdateComment(ctx, commentID, userID, req.Body)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         comment.ID,
		"issue_id":   comment.IssueID,
		"parent_id":  comment.ParentID,
		"user_id":    comment.UserID,
		"body":       comment.Body,
		"created_at": comment.CreatedAt,
		"updated_at": comment.UpdatedAt,
		"edited_at":  comment.EditedAt,
	})
}

// DeleteComment 删除评论
// DELETE /api/v1/comments/:commentId
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	// 获取用户 ID
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	commentIDStr := c.Param("commentId")
	if commentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少评论 ID"})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评论 ID"})
		return
	}

	ctx := c.Request.Context()

	err = h.commentService.DeleteComment(ctx, commentID, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "评论已删除"})
}

// contextWithAuth 从 Gin Context 创建带认证信息的 Context
func (h *CommentHandler) contextWithAuth(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	userID := middleware.GetCurrentUserID(c)
	if userID != uuid.Nil {
		ctx = context.WithValue(ctx, "user_id", userID)
	}

	userRole := middleware.GetCurrentUserRole(c)
	if userRole != "" {
		ctx = context.WithValue(ctx, "user_role", userRole)
	}

	return ctx
}

// handleError 处理错误响应
func (h *CommentHandler) handleError(c *gin.Context, err error) {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "未授权"):
		c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
	case strings.Contains(errMsg, "无权限"):
		c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
	case strings.Contains(errMsg, "不存在") || strings.Contains(errMsg, "未找到"):
		c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
	case strings.Contains(errMsg, "无效"):
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
	}
}
