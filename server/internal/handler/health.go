// Package handler 提供 HTTP 请求处理器
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthHandler 健康检查处理器
type HealthHandler struct {
	dbHealthy bool
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(dbHealthy bool) *HealthHandler {
	return &HealthHandler{
		dbHealthy: dbHealthy,
	}
}

// Check 处理 GET /api/v1/health 请求
func (h *HealthHandler) Check(c *gin.Context) {
	if !h.dbHealthy {
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:  "error",
			Message: "database unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
	})
}
