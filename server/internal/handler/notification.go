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
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// ListNotifications 获取通知列表
// GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析过滤参数
	var read *bool
	if readStr := c.Query("read"); readStr != "" {
		r := readStr == "true"
		read = &r
	}

	// 解析类型过滤
	var types []model.NotificationType
	if typesStr := c.Query("types"); typesStr != "" {
		// 简化实现：支持单个类型过滤
		types = []model.NotificationType{model.NotificationType(typesStr)}
	}

	ctx := h.contextWithAuth(c)

	notifications, total, err := h.notificationService.ListNotifications(ctx, userID, page, pageSize, read, types)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
	})
}

// GetUnreadCount 获取未读通知数量
// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	ctx := h.contextWithAuth(c)

	count, err := h.notificationService.GetUnreadCount(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// MarkAsRead 标记单条通知为已读
// POST /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知 ID"})
		return
	}

	ctx := h.contextWithAuth(c)

	err = h.notificationService.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		if err.Error() == "未找到通知或无权操作" {
			c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已标记为已读"})
}

// MarkAllAsRead 标记所有通知为已读
// POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	ctx := h.contextWithAuth(c)

	count, err := h.notificationService.MarkAllAsRead(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "已标记全部已读",
		"marked":  count,
	})
}

// MarkBatchAsRead 批量标记通知为已读
// POST /api/v1/notifications/batch-read
func (h *NotificationHandler) MarkBatchAsRead(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	// 转换 ID
	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue // 忽略无效 ID
		}
		ids = append(ids, id)
	}

	ctx := h.contextWithAuth(c)

	count, err := h.notificationService.MarkBatchAsRead(ctx, ids, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "批量标记已读成功",
		"marked":  count,
	})
}

// RegisterRoutes 注册路由
func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	notifications := rg.Group("/notifications")
	{
		notifications.GET("", h.ListNotifications)
		notifications.GET("/unread-count", h.GetUnreadCount)
		notifications.POST("/:id/read", h.MarkAsRead)
		notifications.POST("/read-all", h.MarkAllAsRead)
		notifications.POST("/batch-read", h.MarkBatchAsRead)
	}
}

// contextWithAuth 创建带有认证信息的上下文
func (h *NotificationHandler) contextWithAuth(c *gin.Context) context.Context {
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

// =============================================================================
// NotificationPreferenceHandler
// =============================================================================

// NotificationPreferenceHandler 通知偏好配置处理器
type NotificationPreferenceHandler struct {
	preferenceService service.NotificationPreferenceService
}

// NewNotificationPreferenceHandler 创建通知偏好配置处理器
func NewNotificationPreferenceHandler(preferenceService service.NotificationPreferenceService) *NotificationPreferenceHandler {
	return &NotificationPreferenceHandler{preferenceService: preferenceService}
}

// GetPreferences 获取通知偏好配置
// GET /api/v1/notification-preferences
func (h *NotificationPreferenceHandler) GetPreferences(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 解析渠道过滤
	var channel *model.NotificationChannel
	if channelStr := c.Query("channel"); channelStr != "" {
		ch := model.NotificationChannel(channelStr)
		channel = &ch
	}

	ctx := context.Background()

	prefs, err := h.preferenceService.GetPreferences(ctx, userID, channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"preferences": prefs,
	})
}

// UpdatePreferences 更新通知偏好配置
// PUT /api/v1/notification-preferences
func (h *NotificationPreferenceHandler) UpdatePreferences(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Preferences []struct {
			Channel string `json:"channel"`
			Type    string `json:"type"`
			Enabled bool   `json:"enabled"`
		} `json:"preferences"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	// 转换为服务层参数
	updates := make([]service.NotificationPreferenceUpdate, 0, len(req.Preferences))
	for _, pref := range req.Preferences {
		updates = append(updates, service.NotificationPreferenceUpdate{
			Channel: model.NotificationChannel(pref.Channel),
			Type:    model.NotificationType(pref.Type),
			Enabled: pref.Enabled,
		})
	}

	ctx := context.Background()

	err := h.preferenceService.UpdatePreferences(ctx, userID, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已更新"})
}

// RegisterRoutes 注册路由
func (h *NotificationPreferenceHandler) RegisterRoutes(rg *gin.RouterGroup) {
	preferences := rg.Group("/notification-preferences")
	{
		preferences.GET("", h.GetPreferences)
		preferences.PUT("", h.UpdatePreferences)
	}
}
