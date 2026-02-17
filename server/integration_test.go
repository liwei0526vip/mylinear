package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/config"
	"github.com/liwei0526vip/mylinear/internal/handler"
	"github.com/liwei0526vip/mylinear/internal/middleware"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// 设置集成测试
func setupIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB, *redis.Client, uuid.UUID) {
	// 设置数据库
	databaseURL := "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable"
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skip("无法连接数据库")
	}

	// 设置 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建测试工作区
	var workspace model.Workspace
	result := db.Where("name = ?", "Integration Test Workspace").First(&workspace)
	if result.Error == gorm.ErrRecordNotFound {
		workspace = model.Workspace{
			Name: "Integration Test Workspace",
			Slug: "integration-test-workspace",
		}
		db.Create(&workspace)
	}

	// 创建配置
	cfg := &config.Config{
		JWTSecret:        "integration-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	// 初始化服务
	userStore := store.NewUserStore(db)
	jwtService := service.NewJWTService(cfg)
	authService := service.NewAuthService(userStore, jwtService, rdb, cfg)
	userService := service.NewUserService(userStore)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// 创建路由
	router := gin.New()
	authMiddleware := middleware.Auth(jwtService)

	v1 := router.Group("/api/v1")

	// 认证路由（公开）
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// 用户路由（需认证）
	usersGroup := v1.Group("/users")
	usersGroup.Use(authMiddleware)
	{
		usersGroup.GET("/me", userHandler.GetMe)
		usersGroup.PATCH("/me", userHandler.UpdateMe)
	}

	return router, db, rdb, workspace.ID
}

// TestPublicRoutesNoAuth 测试公开路由无需认证
func TestPublicRoutesNoAuth(t *testing.T) {
	router, _, _, workspaceID := setupIntegrationTest(t)

	tests := []struct {
		name   string
		method string
		path   string
		body   map[string]string
	}{
		{
			name:   "注册",
			method: "POST",
			path:   "/api/v1/auth/register",
			body: map[string]string{
				"email":        uuid.New().String()[:8] + "@example.com",
				"username":     uuid.New().String()[:8] + "_user",
				"password":     "Password123!",
				"name":         "Test User",
				"workspace_id": workspaceID.String(),
			},
		},
		{
			name:   "登录",
			method: "POST",
			path:   "/api/v1/auth/login",
			body: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "Password123!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 不应该返回 401（未授权），可能返回其他错误
			if w.Code == http.StatusUnauthorized && tt.name != "登录" {
				t.Errorf("公开路由不应返回 401: %s", w.Body.String())
			}
		})
	}
}

// TestProtectedRoutesNeedAuth 测试受保护路由需要认证
func TestProtectedRoutesNeedAuth(t *testing.T) {
	router, _, _, _ := setupIntegrationTest(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"获取当前用户", "GET", "/api/v1/users/me"},
		{"更新当前用户", "PATCH", "/api/v1/users/me"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("受保护路由应返回 401，实际返回 %d", w.Code)
			}
		})
	}
}

// TestFullAuthFlow 测试完整认证流程
func TestFullAuthFlow(t *testing.T) {
	router, _, _, workspaceID := setupIntegrationTest(t)

	prefix := uuid.New().String()[:8]
	email := prefix + "_fullflow@example.com"
	username := prefix + "_fullflowuser"
	password := "Password123!"

	// 1. 注册
	t.Run("注册", func(t *testing.T) {
		body := map[string]string{
			"email":        email,
			"username":     username,
			"password":     password,
			"name":         "Full Flow User",
			"workspace_id": workspaceID.String(),
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("注册失败: %d, %s", w.Code, w.Body.String())
		}
	})

	var accessToken, refreshToken string

	// 2. 登录
	t.Run("登录", func(t *testing.T) {
		body := map[string]string{
			"email":    email,
			"password": password,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("登录失败: %d, %s", w.Code, w.Body.String())
			return
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		accessToken = data["access_token"].(string)
		refreshToken = data["refresh_token"].(string)
	})

	// 3. 访问受保护资源
	t.Run("访问受保护资源", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("访问受保护资源失败: %d, %s", w.Code, w.Body.String())
		}
	})

	var newRefreshToken string

	// 4. 刷新令牌
	t.Run("刷新令牌", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": refreshToken,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("刷新令牌失败: %d, %s", w.Code, w.Body.String())
			return
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		newRefreshToken = data["refresh_token"].(string)
	})

	// 5. 旧刷新令牌应该失效
	t.Run("旧刷新令牌失效", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": refreshToken,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("旧刷新令牌应该失效，实际返回 %d", w.Code)
		}
	})

	// 6. 登出
	t.Run("登出", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": newRefreshToken,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("登出失败: %d, %s", w.Code, w.Body.String())
		}
	})

	// 7. 登出后令牌失效
	t.Run("登出后令牌失效", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": newRefreshToken,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("登出后令牌应该失效，实际返回 %d", w.Code)
		}
	})
}
