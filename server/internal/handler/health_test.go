package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		dbHealthy    bool
		wantStatus   int
		wantResponse HealthResponse
	}{
		{
			name:       "全部健康",
			dbHealthy:  true,
			wantStatus: http.StatusOK,
			wantResponse: HealthResponse{
				Status: "ok",
			},
		},
		{
			name:       "数据库不可用",
			dbHealthy:  false,
			wantStatus: http.StatusServiceUnavailable,
			wantResponse: HealthResponse{
				Status:  "error",
				Message: "database unavailable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置 Gin 为测试模式
			gin.SetMode(gin.TestMode)

			// 创建测试路由
			router := gin.New()
			h := NewHealthHandler(tt.dbHealthy)
			router.GET("/api/v1/health", h.Check)

			// 创建测试请求
			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证状态码
			if w.Code != tt.wantStatus {
				t.Errorf("HealthCheck() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// 验证响应体
			var response HealthResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to parse response: %v", err)
				return
			}

			if response.Status != tt.wantResponse.Status {
				t.Errorf("HealthCheck() status = %v, want %v", response.Status, tt.wantResponse.Status)
			}

			if response.Message != tt.wantResponse.Message {
				t.Errorf("HealthCheck() message = %v, want %v", response.Message, tt.wantResponse.Message)
			}
		})
	}
}

func TestHealthCheck_ResponseContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	h := NewHealthHandler(true)
	router.GET("/api/v1/health", h.Check)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %v, want application/json; charset=utf-8", contentType)
	}
}
