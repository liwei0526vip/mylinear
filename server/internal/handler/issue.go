// Package handler 提供 HTTP 处理器
package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/service"
)

// IssueHandler Issue 处理器
type IssueHandler struct {
	issueService service.IssueService
}

// NewIssueHandler 创建 Issue 处理器
func NewIssueHandler(issueService service.IssueService) *IssueHandler {
	return &IssueHandler{issueService: issueService}
}

// CreateIssue 创建 Issue
// POST /api/v1/teams/:teamId/issues
func (h *IssueHandler) CreateIssue(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	if teamIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少团队 ID"})
		return
	}

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的团队 ID"})
		return
	}

	var req struct {
		Title       string  `json:"title" binding:"required"`
		Description *string `json:"description"`
		StatusID    string  `json:"status_id"`
		Priority    int     `json:"priority"`
		AssigneeID  *string `json:"assignee_id"`
		ProjectID   *string `json:"project_id"`
		Labels      []string `json:"labels"`
		DueDate     *string `json:"due_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	ctx := h.contextWithAuth(c)

	var statusID uuid.UUID
	if req.StatusID != "" {
		statusID, _ = uuid.Parse(req.StatusID)
	}

	var assigneeID *uuid.UUID
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		id, _ := uuid.Parse(*req.AssigneeID)
		assigneeID = &id
	}

	var projectID *uuid.UUID
	if req.ProjectID != nil && *req.ProjectID != "" {
		id, _ := uuid.Parse(*req.ProjectID)
		projectID = &id
	}

	params := &service.CreateIssueParams{
		TeamID:      teamID,
		Title:       req.Title,
		Description: req.Description,
		StatusID:    statusID,
		Priority:    req.Priority,
		AssigneeID:  assigneeID,
		ProjectID:   projectID,
	}

	issue, err := h.issueService.CreateIssue(ctx, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          issue.ID,
		"team_id":     issue.TeamID,
		"number":      issue.Number,
		"title":       issue.Title,
		"description": issue.Description,
		"status_id":   issue.StatusID,
		"priority":    issue.Priority,
		"assignee_id": issue.AssigneeID,
		"position":    issue.Position,
		"created_at":  issue.CreatedAt,
	})
}

// GetIssue 获取 Issue
// GET /api/v1/issues/:id
func (h *IssueHandler) GetIssue(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	ctx := h.contextWithAuth(c)

	issue, err := h.issueService.GetIssue(ctx, issueID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          issue.ID,
		"team_id":     issue.TeamID,
		"number":      issue.Number,
		"title":       issue.Title,
		"description": issue.Description,
		"status_id":   issue.StatusID,
		"priority":    issue.Priority,
		"assignee_id": issue.AssigneeID,
		"project_id":  issue.ProjectID,
		"position":    issue.Position,
		"created_at":  issue.CreatedAt,
		"updated_at":  issue.UpdatedAt,
		"created_by":  issue.CreatedByID,
	})
}

// ListIssues 获取 Issue 列表
// GET /api/v1/teams/:teamId/issues
func (h *IssueHandler) ListIssues(c *gin.Context) {
	teamID := c.Param("teamId")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少团队 ID"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析过滤参数
	filter := &service.IssueFilter{}
	if statusID := c.Query("status_id"); statusID != "" {
		filter.StatusID = &statusID
	}
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority, _ := strconv.Atoi(priorityStr)
		filter.Priority = &priority
	}
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		filter.AssigneeID = &assigneeID
	}
	if projectID := c.Query("project_id"); projectID != "" {
		filter.ProjectID = &projectID
	}

	ctx := h.contextWithAuth(c)

	issues, total, err := h.issueService.ListIssues(ctx, teamID, filter, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	result := make([]gin.H, len(issues))
	for i, issue := range issues {
		result[i] = gin.H{
			"id":          issue.ID,
			"team_id":     issue.TeamID,
			"number":      issue.Number,
			"title":       issue.Title,
			"description": issue.Description,
			"status_id":   issue.StatusID,
			"priority":    issue.Priority,
			"assignee_id": issue.AssigneeID,
			"position":    issue.Position,
			"created_at":  issue.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"issues": result,
		"total":  total,
		"page":   page,
	})
}

// UpdateIssue 更新 Issue
// PUT /api/v1/issues/:id
func (h *IssueHandler) UpdateIssue(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	ctx := h.contextWithAuth(c)

	issue, err := h.issueService.UpdateIssue(ctx, issueID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          issue.ID,
		"team_id":     issue.TeamID,
		"number":      issue.Number,
		"title":       issue.Title,
		"description": issue.Description,
		"status_id":   issue.StatusID,
		"priority":    issue.Priority,
		"assignee_id": issue.AssigneeID,
		"position":    issue.Position,
		"updated_at":  issue.UpdatedAt,
	})
}

// DeleteIssue 删除 Issue
// DELETE /api/v1/issues/:id
func (h *IssueHandler) DeleteIssue(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	ctx := h.contextWithAuth(c)

	err := h.issueService.DeleteIssue(ctx, issueID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Issue 已删除"})
}

// RestoreIssue 恢复已删除的 Issue
// POST /api/v1/issues/:id/restore
func (h *IssueHandler) RestoreIssue(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	ctx := h.contextWithAuth(c)

	err := h.issueService.RestoreIssue(ctx, issueID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Issue 已恢复"})
}

// Subscribe 订阅 Issue
// POST /api/v1/issues/:id/subscribe
func (h *IssueHandler) Subscribe(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	ctx := h.contextWithAuth(c)

	err := h.issueService.Subscribe(ctx, issueID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已订阅"})
}

// Unsubscribe 取消订阅
// DELETE /api/v1/issues/:id/subscribe
func (h *IssueHandler) Unsubscribe(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	ctx := h.contextWithAuth(c)

	err := h.issueService.Unsubscribe(ctx, issueID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已取消订阅"})
}

// ListSubscribers 获取订阅者列表
// GET /api/v1/issues/:id/subscribers
func (h *IssueHandler) ListSubscribers(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	ctx := h.contextWithAuth(c)

	subscribers, err := h.issueService.ListSubscribers(ctx, issueID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	result := make([]gin.H, len(subscribers))
	for i, user := range subscribers {
		result[i] = gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"name":       user.Name,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
		}
	}

	c.JSON(http.StatusOK, gin.H{"subscribers": result})
}

// UpdatePosition 更新 Issue 位置
// PUT /api/v1/issues/:id/position
func (h *IssueHandler) UpdatePosition(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Issue ID"})
		return
	}

	var req struct {
		Position float64 `json:"position" binding:"required"`
		StatusID *string `json:"status_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	ctx := h.contextWithAuth(c)

	err := h.issueService.UpdatePosition(ctx, issueID, req.Position, req.StatusID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "位置已更新"})
}

// contextWithAuth 从 Gin Context 创建带认证信息的 Context
func (h *IssueHandler) contextWithAuth(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	// 从 Gin Context 获取用户信息
	if userID, exists := c.Get("user_id"); exists {
		if uid, err := uuid.Parse(userID.(string)); err == nil {
			ctx = context.WithValue(ctx, "user_id", uid)
		}
	}
	if userRole, exists := c.Get("user_role"); exists {
		ctx = context.WithValue(ctx, "user_role", userRole)
	}

	return ctx
}

// handleError 处理错误响应
func (h *IssueHandler) handleError(c *gin.Context, err error) {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "未认证"):
		c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
	case strings.Contains(errMsg, "无权限"):
		c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
	case strings.Contains(errMsg, "不存在"):
		c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
	case strings.Contains(errMsg, "无效"):
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
	case strings.Contains(errMsg, "标题"):
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
	}
}
