package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylinear/server/internal/config"
	"github.com/mylinear/server/internal/middleware"
	"github.com/mylinear/server/internal/model"
	"github.com/mylinear/server/internal/service"
	"github.com/mylinear/server/internal/store"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var userTestDB *gorm.DB
var userTestRedis *redis.Client
var userTestWorkspaceID uuid.UUID

func setupUserHandlerTest(t *testing.T) (*gin.Engine, service.AuthService, service.UserService, context.Context) {
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
	result := db.Where("name = ?", "User Handler Test Workspace").First(&workspace)
	if result.Error == gorm.ErrRecordNotFound {
		workspace = model.Workspace{
			Name: "User Handler Test Workspace",
			Slug: "user-handler-test-workspace",
		}
		db.Create(&workspace)
	}

	userTestDB = db
	userTestRedis = rdb
	userTestWorkspaceID = workspace.ID

	// 创建服务
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	userStore := store.NewUserStore(db)
	jwtService := service.NewJWTService(cfg)
	authService := service.NewAuthService(userStore, jwtService, rdb, cfg)
	userService := service.NewUserService(userStore)

	// 创建路由
	router := gin.New()

	return router, authService, userService, context.Background()
}

// =============================================================================
// UserHandler 测试
// =============================================================================

func TestUserHandler_GetMe_Success(t *testing.T) {
	router, authService, userService, ctx := setupUserHandlerTest(t)
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userHandler := NewUserHandler(userService)
	authMiddleware := middleware.Auth(jwtService)

	// 注册用户
	prefix := uuid.New().String()[:8]
	email := prefix + "_getme@example.com"
	password := "Password123!"
	user, _, _, _ := authService.Register(ctx, userTestWorkspaceID, email, prefix+"_getmeuser", password, "GetMe User")

	// 生成访问令牌
	accessToken, _ := jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)

	// 设置路由
	router.GET("/api/v1/users/me", authMiddleware, userHandler.GetMe)

	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

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
	if data["email"] != email {
		t.Errorf("email = %v, want %v", data["email"], email)
	}
}

func TestUserHandler_GetMe_Unauthorized(t *testing.T) {
	router, _, userService, _ := setupUserHandlerTest(t)
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userHandler := NewUserHandler(userService)
	authMiddleware := middleware.Auth(jwtService)

	router.GET("/api/v1/users/me", authMiddleware, userHandler.GetMe)

	// 不带 Authorization 头
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUserHandler_UpdateMe_Success(t *testing.T) {
	router, authService, userService, ctx := setupUserHandlerTest(t)
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userHandler := NewUserHandler(userService)
	authMiddleware := middleware.Auth(jwtService)

	// 注册用户
	prefix := uuid.New().String()[:8]
	email := prefix + "_updateme@example.com"
	password := "Password123!"
	user, _, _, _ := authService.Register(ctx, userTestWorkspaceID, email, prefix+"_updatemeuser", password, "UpdateMe User")

	// 生成访问令牌
	accessToken, _ := jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)

	// 设置路由
	router.PATCH("/api/v1/users/me", authMiddleware, userHandler.UpdateMe)

	// 更新用户信息
	body := map[string]string{
		"name": "Updated Name",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PATCH", "/api/v1/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
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
	if data["name"] != "Updated Name" {
		t.Errorf("name = %v, want %v", data["name"], "Updated Name")
	}
}

func TestUserHandler_UpdateMe_EmailConflict(t *testing.T) {
	router, authService, userService, ctx := setupUserHandlerTest(t)
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userHandler := NewUserHandler(userService)
	authMiddleware := middleware.Auth(jwtService)

	// 注册两个用户
	prefix := uuid.New().String()[:8]
	email1 := prefix + "_conflict1@example.com"
	email2 := prefix + "_conflict2@example.com"
	user1, _, _, _ := authService.Register(ctx, userTestWorkspaceID, email1, prefix+"_conflict1user", "Password123!", "Conflict User 1")
	authService.Register(ctx, userTestWorkspaceID, email2, prefix+"_conflict2user", "Password123!", "Conflict User 2")

	// 生成用户1的访问令牌
	accessToken, _ := jwtService.GenerateAccessToken(user1.ID, user1.Email, user1.Role)

	// 设置路由
	router.PATCH("/api/v1/users/me", authMiddleware, userHandler.UpdateMe)

	// 尝试更新为已存在的邮箱
	body := map[string]string{
		"email": email2,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PATCH", "/api/v1/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestUserHandler_UpdateMe_UsernameConflict(t *testing.T) {
	router, authService, userService, ctx := setupUserHandlerTest(t)
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userHandler := NewUserHandler(userService)
	authMiddleware := middleware.Auth(jwtService)

	// 注册两个用户
	prefix := uuid.New().String()[:8]
	email1 := prefix + "_uconflict1@example.com"
	email2 := prefix + "_uconflict2@example.com"
	user1, _, _, _ := authService.Register(ctx, userTestWorkspaceID, email1, prefix+"_uconflict1user", "Password123!", "Username Conflict 1")
	authService.Register(ctx, userTestWorkspaceID, email2, prefix+"_uconflict2user", "Password123!", "Username Conflict 2")

	// 生成用户1的访问令牌
	accessToken, _ := jwtService.GenerateAccessToken(user1.ID, user1.Email, user1.Role)

	// 设置路由
	router.PATCH("/api/v1/users/me", authMiddleware, userHandler.UpdateMe)

	// 尝试更新为已存在的用户名
	body := map[string]string{
		"username": prefix + "_uconflict2user",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PATCH", "/api/v1/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusConflict)
	}
}

// =============================================================================
// UploadAvatar 测试
// =============================================================================

// createMultipartFile 创建模拟的文件上传请求
func createMultipartFile(filename, contentType string, content []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("avatar", filename)
	part.Write(content)
	writer.Close()

	return body, writer.FormDataContentType()
}

func TestUserHandler_UploadAvatar_FileTooLarge(t *testing.T) {
	router, authService, userService, ctx := setupUserHandlerTest(t)
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userHandler := NewUserHandlerWithAvatar(userService, nil)
	authMiddleware := middleware.Auth(jwtService)

	// 注册用户
	prefix := uuid.New().String()[:8]
	email := prefix + "_avatar_big@example.com"
	user, _, _, _ := authService.Register(ctx, userTestWorkspaceID, email, prefix+"_avatarbiguser", "Password123!", "Avatar Big User")

	// 生成访问令牌
	accessToken, _ := jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)

	// 设置路由
	router.POST("/api/v1/users/me/avatar", authMiddleware, userHandler.UploadAvatar)

	// 创建超过 2MB 的文件内容
	largeContent := make([]byte, 3*1024*1024) // 3MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	body, contentType := createMultipartFile("large.png", "image/png", largeContent)

	req := httptest.NewRequest("POST", "/api/v1/users/me/avatar", body)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestUserHandler_UploadAvatar_InvalidFileType(t *testing.T) {
	router, authService, userService, ctx := setupUserHandlerTest(t)
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userHandler := NewUserHandlerWithAvatar(userService, nil)
	authMiddleware := middleware.Auth(jwtService)

	// 注册用户
	prefix := uuid.New().String()[:8]
	email := prefix + "_avatar_type@example.com"
	user, _, _, _ := authService.Register(ctx, userTestWorkspaceID, email, prefix+"_avatartypeuser", "Password123!", "Avatar Type User")

	// 生成访问令牌
	accessToken, _ := jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)

	// 设置路由
	router.POST("/api/v1/users/me/avatar", authMiddleware, userHandler.UploadAvatar)

	// 创建无效文件类型的请求（exe 文件）
	body, contentType := createMultipartFile("virus.exe", "application/octet-stream", []byte("fake exe content"))

	req := httptest.NewRequest("POST", "/api/v1/users/me/avatar", body)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestUserHandler_UploadAvatar_MissingFile(t *testing.T) {
	router, authService, userService, ctx := setupUserHandlerTest(t)
	cfg := &config.Config{
		JWTSecret:        "user-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userHandler := NewUserHandlerWithAvatar(userService, nil)
	authMiddleware := middleware.Auth(jwtService)

	// 注册用户
	prefix := uuid.New().String()[:8]
	email := prefix + "_avatar_missing@example.com"
	user, _, _, _ := authService.Register(ctx, userTestWorkspaceID, email, prefix+"_avatarmisssuer", "Password123!", "Avatar Missing User")

	// 生成访问令牌
	accessToken, _ := jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)

	// 设置路由
	router.POST("/api/v1/users/me/avatar", authMiddleware, userHandler.UploadAvatar)

	// 创建没有文件的请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/avatar", body)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestUserHandler_UploadAvatar_Unauthorized(t *testing.T) {
	router, _, userService, _ := setupUserHandlerTest(t)

	userHandler := NewUserHandlerWithAvatar(userService, nil)

	// 设置路由（不带认证中间件进行测试会更准确，但这里测试的是未带 token 的情况）
	router.POST("/api/v1/users/me/avatar", userHandler.UploadAvatar)

	// 创建一个有效的图片文件（1x1 PNG）
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG magic number
	body, contentType := createMultipartFile("avatar.png", "image/png", pngHeader)

	req := httptest.NewRequest("POST", "/api/v1/users/me/avatar", body)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// UserHandler 内部会检查用户上下文，如果没有应该返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestValidateFile 测试文件验证逻辑
func TestValidateFile(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		size        int64
		wantError   bool
	}{
		{"有效的 PNG", "image/png", 1024, false},
		{"有效的 JPEG", "image/jpeg", 1024, false},
		{"有效的 GIF", "image/gif", 1024, false},
		{"有效的 WebP", "image/webp", 1024, false},
		{"无效的类型 - PDF", "application/pdf", 1024, true},
		{"无效的类型 - EXE", "application/octet-stream", 1024, true},
		{"文件过大", "image/png", 3 * 1024 * 1024, true},
		{"空文件", "image/png", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAvatarFile(tt.contentType, tt.size)
			if (err != nil) != tt.wantError {
				t.Errorf("validateAvatarFile(%s, %d) error = %v, wantError %v", tt.contentType, tt.size, err, tt.wantError)
			}
		})
	}
}
