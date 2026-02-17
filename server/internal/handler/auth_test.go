package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/config"
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

var handlerTestDB *gorm.DB
var handlerTestRedis *redis.Client
var handlerTestWorkspaceID uuid.UUID

func setupHandlerTest(t *testing.T) (*gin.Engine, service.AuthService, context.Context) {
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
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("无法连接 Redis")
	}

	// 创建测试工作区
	var workspace model.Workspace
	result := db.Where("name = ?", "Handler Test Workspace").First(&workspace)
	if result.Error == gorm.ErrRecordNotFound {
		workspace = model.Workspace{
			Name: "Handler Test Workspace",
			Slug: "handler-test-workspace",
		}
		db.Create(&workspace)
	}

	handlerTestDB = db
	handlerTestRedis = rdb
	handlerTestWorkspaceID = workspace.ID

	// 创建服务
	cfg := &config.Config{
		JWTSecret:        "handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	userStore := store.NewUserStore(db)
	jwtService := service.NewJWTService(cfg)
	authService := service.NewAuthService(userStore, jwtService, rdb, cfg)

	// 创建路由
	router := gin.New()

	return router, authService, context.Background()
}

// =============================================================================
// AuthHandler 测试
// =============================================================================

func TestAuthHandler_Register_Success(t *testing.T) {
	router, authService, _ := setupHandlerTest(t)
	authHandler := NewAuthHandler(authService)

	router.POST("/api/v1/auth/register", authHandler.Register)

	prefix := uuid.New().String()[:8]
	body := map[string]string{
		"email":        prefix + "_handler@example.com",
		"username":     prefix + "_handleruser",
		"password":     "Password123!",
		"name":         "Handler Test User",
		"workspace_id": handlerTestWorkspaceID.String(),
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	// 验证响应包含 token
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Error("响应缺少 data 字段")
		return
	}
	if data["access_token"] == "" {
		t.Error("响应缺少 access_token")
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	router, authService, _ := setupHandlerTest(t)
	authHandler := NewAuthHandler(authService)

	router.POST("/api/v1/auth/register", authHandler.Register)

	prefix := uuid.New().String()[:8]
	body := map[string]string{
		"email":        prefix + "_dup@example.com",
		"username":     prefix + "_dupuser1",
		"password":     "Password123!",
		"name":         "Dup User 1",
		"workspace_id": handlerTestWorkspaceID.String(),
	}
	jsonBody, _ := json.Marshal(body)

	// 第一次注册
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 第二次用相同邮箱注册
	body["username"] = prefix + "_dupuser2"
	jsonBody, _ = json.Marshal(body)
	req = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestAuthHandler_Register_InvalidInput(t *testing.T) {
	router, authService, _ := setupHandlerTest(t)
	authHandler := NewAuthHandler(authService)

	router.POST("/api/v1/auth/register", authHandler.Register)

	prefix := uuid.New().String()[:8]

	tests := []struct {
		name string
		body map[string]string
	}{
		{
			name: "缺少邮箱",
			body: map[string]string{
				"username": prefix + "_user",
				"password": "Password123!",
				"name":     "Test User",
			},
		},
		{
			name: "缺少密码",
			body: map[string]string{
				"email":    prefix + "@example.com",
				"username": prefix + "_user",
				"name":     "Test User",
			},
		},
		{
			name: "密码太弱",
			body: map[string]string{
				"email":    prefix + "_weak@example.com",
				"username": prefix + "_weakuser",
				"password": "weak",
				"name":     "Test User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusCreated {
				t.Error("应该返回错误状态码")
			}
		})
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	router, authService, ctx := setupHandlerTest(t)
	authHandler := NewAuthHandler(authService)

	router.POST("/api/v1/auth/login", authHandler.Login)

	// 先注册用户
	prefix := uuid.New().String()[:8]
	email := prefix + "_login@example.com"
	password := "Password123!"
	authService.Register(ctx, handlerTestWorkspaceID, email, prefix+"_loginuser", password, "Login User")

	// 登录
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
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Error("响应缺少 data 字段")
		return
	}
	if data["access_token"] == "" {
		t.Error("响应缺少 access_token")
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	router, authService, ctx := setupHandlerTest(t)
	authHandler := NewAuthHandler(authService)

	router.POST("/api/v1/auth/login", authHandler.Login)

	// 先注册用户
	prefix := uuid.New().String()[:8]
	email := prefix + "_wrongpass@example.com"
	authService.Register(ctx, handlerTestWorkspaceID, email, prefix+"_wrongpassuser", "Password123!", "Wrong Pass User")

	// 用错误密码登录
	body := map[string]string{
		"email":    email,
		"password": "WrongPassword123!",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	router, authService, ctx := setupHandlerTest(t)
	authHandler := NewAuthHandler(authService)

	router.POST("/api/v1/auth/refresh", authHandler.Refresh)

	// 注册用户并获取刷新令牌
	prefix := uuid.New().String()[:8]
	_, _, refreshToken, _ := authService.Register(ctx, handlerTestWorkspaceID, prefix+"_refresh@example.com", prefix+"_refreshuser", "Password123!", "Refresh User")

	// 刷新令牌
	body := map[string]string{
		"refresh_token": refreshToken,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	router, authService, ctx := setupHandlerTest(t)
	authHandler := NewAuthHandler(authService)

	router.POST("/api/v1/auth/logout", authHandler.Logout)

	// 注册用户并获取刷新令牌
	prefix := uuid.New().String()[:8]
	_, _, refreshToken, _ := authService.Register(ctx, handlerTestWorkspaceID, prefix+"_logout@example.com", prefix+"_logoutuser", "Password123!", "Logout User")

	// 登出
	body := map[string]string{
		"refresh_token": refreshToken,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}
