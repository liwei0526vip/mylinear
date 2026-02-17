package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
)

type WorkflowHandler struct {
	workflowService service.WorkflowService
}

type CreateStateRequest struct {
	Name        string          `json:"name" binding:"required"`
	Type        model.StateType `json:"type" binding:"required"`
	Color       string          `json:"color"`
	Position    float64         `json:"position"`
	Description string          `json:"description"`
}

func NewWorkflowHandler(workflowService service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{workflowService: workflowService}
}

// CreateState 创建工作流状态
func (h *WorkflowHandler) CreateState(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的团队 ID"})
		return
	}

	var req CreateStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state, err := h.workflowService.CreateState(c.Request.Context(), &service.CreateStateParams{
		TeamID:      teamID,
		Name:        req.Name,
		Type:        req.Type,
		Color:       req.Color,
		Position:    req.Position,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": state})
}

// ListStates 获取团队的所有工作流状态
func (h *WorkflowHandler) ListStates(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的团队 ID"})
		return
	}

	states, err := h.workflowService.ListStates(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": states})
}

type UpdateStateRequest struct {
	Name        *string  `json:"name"`
	Color       *string  `json:"color"`
	Position    *float64 `json:"position"`
	Description *string  `json:"description"`
}

// UpdateState 更新工作流状态
func (h *WorkflowHandler) UpdateState(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态 ID"})
		return
	}

	var req UpdateStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state, err := h.workflowService.UpdateState(c.Request.Context(), id, &service.UpdateStateParams{
		Name:        req.Name,
		Color:       req.Color,
		Position:    req.Position,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": state})
}

// DeleteState 删除工作流状态
func (h *WorkflowHandler) DeleteState(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态 ID"})
		return
	}

	if err := h.workflowService.DeleteState(c.Request.Context(), id); err != nil {
		// Business logic errors (like cannot delete last state) should ideally be 400
		// But service layer currently returns generic error.
		// For simplicity, return 500 or 400 based on message.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
