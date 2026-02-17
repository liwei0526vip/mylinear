// Package handler 提供 HTTP 处理器
package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylinear/server/internal/middleware"
	"github.com/mylinear/server/internal/model"
	"github.com/mylinear/server/internal/service"
)

// WorkspaceHandler 工作区处理器
type WorkspaceHandler struct {
	workspaceService service.WorkspaceService
}

// NewWorkspaceHandler 创建工作区处理器
func NewWorkspaceHandler(workspaceService service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
	}
}

// GetWorkspace 获取工作区
func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少工作区ID"})
		return
	}

	// 设置上下文
	ctx := contextWithUser(c)

	workspace, err := h.workspaceService.GetWorkspace(ctx, workspaceID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         workspace.ID,
		"name":       workspace.Name,
		"slug":       workspace.Slug,
		"logo_url":   workspace.LogoURL,
		"created_at": workspace.CreatedAt,
		"updated_at": workspace.UpdatedAt,
	})
}

// UpdateWorkspace 更新工作区
func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少工作区ID"})
		return
	}

	var req struct {
		Name    *string `json:"name"`
		LogoURL *string `json:"logo_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	// 构建更新
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.LogoURL != nil {
		updates["logo_url"] = *req.LogoURL
	}

	// 设置上下文
	ctx := contextWithUser(c)

	workspace, err := h.workspaceService.UpdateWorkspace(ctx, workspaceID, updates)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         workspace.ID,
		"name":       workspace.Name,
		"slug":       workspace.Slug,
		"logo_url":   workspace.LogoURL,
		"created_at": workspace.CreatedAt,
		"updated_at": workspace.UpdatedAt,
	})
}

// contextWithUser 将用户信息注入上下文
func contextWithUser(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	user := middleware.GetCurrentUser(c)
	if user != nil {
		ctx = context.WithValue(ctx, "user_id", uuid.MustParse(user.UserID))
		ctx = context.WithValue(ctx, "user_role", model.Role(user.Role))
	}
	return ctx
}

// handleError 处理错误响应
func handleError(c *gin.Context, err error) {
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
	case strings.Contains(errMsg, "已存在") || strings.Contains(errMsg, "已被使用") || strings.Contains(errMsg, "duplicate key"):
		c.JSON(http.StatusConflict, gin.H{"error": errMsg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
	}
}
