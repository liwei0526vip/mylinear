package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
)

type LabelHandler struct {
	labelService service.LabelService
	teamStore    store.TeamStore
}

func NewLabelHandler(labelService service.LabelService, teamStore store.TeamStore) *LabelHandler {
	return &LabelHandler{labelService: labelService, teamStore: teamStore}
}

type CreateLabelRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// CreateLabel 为团队创建标签
func (h *LabelHandler) CreateLabel(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的团队 ID"})
		return
	}

	// 获取 Team 以确定 WorkspaceID
	team, err := h.teamStore.GetByID(c.Request.Context(), teamID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "团队不存在"})
		return
	}

	var req CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	label, err := h.labelService.CreateLabel(c.Request.Context(), &service.CreateLabelParams{
		WorkspaceID: team.WorkspaceID,
		TeamID:      &teamID,
		Name:        req.Name,
		Color:       req.Color,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": label})
}

// ListLabels 获取团队及其所属工作区的所有标签
func (h *LabelHandler) ListLabels(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的团队 ID"})
		return
	}

	// 获取 Team 以确定 WorkspaceID
	team, err := h.teamStore.GetByID(c.Request.Context(), teamID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "团队不存在"})
		return
	}

	labels, err := h.labelService.ListLabels(c.Request.Context(), team.WorkspaceID, &teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": labels})
}
