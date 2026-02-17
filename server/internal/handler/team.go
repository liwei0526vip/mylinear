// Package handler 提供 HTTP 处理器
package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylinear/server/internal/service"
)

// TeamHandler 团队处理器
type TeamHandler struct {
	teamService service.TeamService
}

// NewTeamHandler 创建团队处理器
func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// ListTeams 获取团队列表
func (h *TeamHandler) ListTeams(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少工作区ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx := contextWithUser(c)

	teams, total, err := h.teamService.ListTeams(ctx, workspaceID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]gin.H, len(teams))
	for i, team := range teams {
		result[i] = gin.H{
			"id":           team.ID,
			"workspace_id": team.WorkspaceID,
			"name":         team.Name,
			"key":          team.Key,
			"description":  team.Description,
			"created_at":   team.CreatedAt,
			"updated_at":   team.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": result,
		"total": total,
		"page":  page,
	})
}

// CreateTeam 创建团队
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Key         string `json:"key" binding:"required"`
		Description string `json:"description"`
		WorkspaceID string `json:"workspace_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	ctx := contextWithUser(c)
	ctx = context.WithValue(ctx, "workspace_id", uuid.MustParse(req.WorkspaceID))

	team, err := h.teamService.CreateTeam(ctx, req.Name, req.Key, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           team.ID,
		"workspace_id": team.WorkspaceID,
		"name":         team.Name,
		"key":          team.Key,
		"created_at":   team.CreatedAt,
	})
}

// GetTeam 获取团队
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamID := c.Param("teamId")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少团队ID"})
		return
	}

	ctx := contextWithUser(c)

	team, err := h.teamService.GetTeam(ctx, teamID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           team.ID,
		"workspace_id": team.WorkspaceID,
		"name":         team.Name,
		"key":          team.Key,
		"created_at":   team.CreatedAt,
		"updated_at":   team.UpdatedAt,
	})
}

// UpdateTeam 更新团队
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	teamID := c.Param("teamId")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少团队ID"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Key         *string `json:"key"`
		Description *string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Key != nil {
		updates["key"] = *req.Key
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	ctx := contextWithUser(c)

	team, err := h.teamService.UpdateTeam(ctx, teamID, updates)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           team.ID,
		"workspace_id": team.WorkspaceID,
		"name":         team.Name,
		"key":          team.Key,
		"updated_at":   team.UpdatedAt,
	})
}

// DeleteTeam 删除团队
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	teamID := c.Param("teamId")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少团队ID"})
		return
	}

	ctx := contextWithUser(c)

	err := h.teamService.DeleteTeam(ctx, teamID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "团队已删除"})
}
