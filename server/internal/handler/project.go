// Package handler 提供 HTTP 处理器
package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/middleware"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectService service.ProjectService
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// CreateProject 创建项目
// POST /api/v1/workspaces/:workspaceId/projects
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	if workspaceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少工作区 ID"})
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工作区 ID"})
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description *string  `json:"description"`
		LeadID      *string  `json:"lead_id"`
		StartDate   *string  `json:"start_date"`
		TargetDate  *string  `json:"target_date"`
		Teams       []string `json:"teams"`
		Labels      []string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	// 验证名称
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称不能为空"})
		return
	}

	params := &service.CreateProjectParams{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
	}

	if req.LeadID != nil && *req.LeadID != "" {
		leadID, _ := uuid.Parse(*req.LeadID)
		params.LeadID = &leadID
	}

	if len(req.Teams) > 0 {
		params.Teams = make([]uuid.UUID, len(req.Teams))
		for i, id := range req.Teams {
			params.Teams[i], _ = uuid.Parse(id)
		}
	}

	if len(req.Labels) > 0 {
		params.Labels = make([]uuid.UUID, len(req.Labels))
		for i, id := range req.Labels {
			params.Labels[i], _ = uuid.Parse(id)
		}
	}

	ctx := h.contextWithAuth(c)
	project, err := h.projectService.CreateProject(ctx, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           project.ID,
		"workspace_id": project.WorkspaceID,
		"name":         project.Name,
		"description":  project.Description,
		"status":       project.Status,
		"lead_id":      project.LeadID,
		"created_at":   project.CreatedAt,
	})
}

// ListTeamProjects 获取团队项目列表
// GET /api/v1/teams/:teamId/projects
func (h *ProjectHandler) ListTeamProjects(c *gin.Context) {
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

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// 解析过滤条件
	var filter *store.ProjectFilter
	if status := c.Query("status"); status != "" {
		projectStatus := model.ProjectStatus(status)
		filter = &store.ProjectFilter{Status: &projectStatus}
	}

	ctx := h.contextWithAuth(c)
	projects, total, err := h.projectService.ListProjectsByTeam(ctx, teamID, filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": projects,
		"total": total,
		"page":  page,
	})
}

// GetProject 获取项目详情
// GET /api/v1/projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目 ID"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目 ID"})
		return
	}

	ctx := h.contextWithAuth(c)
	project, err := h.projectService.GetProject(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject 更新项目
// PUT /api/v1/projects/:id
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目 ID"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目 ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	ctx := h.contextWithAuth(c)
	project, err := h.projectService.UpdateProject(ctx, id, req)
	if err != nil {
		if err.Error() == "项目不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject 删除项目
// DELETE /api/v1/projects/:id
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目 ID"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目 ID"})
		return
	}

	ctx := h.contextWithAuth(c)
	if err := h.projectService.DeleteProject(ctx, id); err != nil {
		if err.Error() == "权限不足" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "项目不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetProjectProgress 获取项目进度
// GET /api/v1/projects/:id/progress
func (h *ProjectHandler) GetProjectProgress(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目 ID"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目 ID"})
		return
	}

	ctx := h.contextWithAuth(c)
	progress, err := h.projectService.GetProjectProgress(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// ListProjectIssues 获取项目关联的 Issue 列表
// GET /api/v1/projects/:id/issues
func (h *ProjectHandler) ListProjectIssues(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目 ID"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目 ID"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// 解析过滤条件
	var filter *store.IssueFilter
	if statusID := c.Query("status_id"); statusID != "" {
		sid, _ := uuid.Parse(statusID)
		filter = &store.IssueFilter{StatusID: &sid}
	}

	ctx := h.contextWithAuth(c)
	issues, total, err := h.projectService.ListProjectIssues(ctx, id, filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": issues,
		"total": total,
		"page":  page,
	})
}

// contextWithAuth 从 Gin Context 创建带认证信息的 Context
func (h *ProjectHandler) contextWithAuth(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	// 从 Gin Context 获取用户信息
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
