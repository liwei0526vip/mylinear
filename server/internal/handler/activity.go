// Package handler 提供 HTTP 处理器
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/middleware"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
)

// ActivityHandler 活动处理器
type ActivityHandler struct {
	activityService service.ActivityService
}

// NewActivityHandler 创建活动处理器
func NewActivityHandler(activityService service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

// ListIssueActivities 获取 Issue 的活动列表
// GET /api/v1/issues/:id/activities
func (h *ActivityHandler) ListIssueActivities(c *gin.Context) {
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

	// 解析类型过滤参数
	var types []model.ActivityType
	if typesStr := c.Query("types"); typesStr != "" {
		typeStrs := strings.Split(typesStr, ",")
		for _, t := range typeStrs {
			activityType := model.ActivityType(strings.TrimSpace(t))
			if activityType.Valid() {
				types = append(types, activityType)
			}
		}
	}

	ctx := c.Request.Context()

	activities, total, err := h.activityService.GetIssueActivities(ctx, issueID, page, pageSize, types)
	if err != nil {
		h.handleError(c, err)
		return
	}

	result := make([]gin.H, len(activities))
	for i, activity := range activities {
		result[i] = gin.H{
			"id":         activity.ID,
			"issue_id":   activity.IssueID,
			"type":       activity.Type,
			"actor_id":   activity.ActorID,
			"payload":    activity.Payload,
			"created_at": activity.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": result,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// handleError 处理错误响应
func (h *ActivityHandler) handleError(c *gin.Context, err error) {
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

// contextWithAuth 从 Gin Context 创建带认证信息的 Context（备用）
func (h *ActivityHandler) contextWithAuth(c *gin.Context) *gin.Context {
	// 这里可以添加认证逻辑
	return c
}

// getCurrentUserID 获取当前用户 ID
func (h *ActivityHandler) getCurrentUserID(c *gin.Context) uuid.UUID {
	return middleware.GetCurrentUserID(c)
}
