package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mylinear/server/internal/config"
	"github.com/mylinear/server/internal/model"
	"github.com/mylinear/server/internal/store"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestAuthService_Interface 测试 AuthService 接口定义存在
func TestAuthService_Interface(t *testing.T) {
	var _ AuthService = (*authService)(nil)
}

// 认证服务集成测试
// 需要真实数据库和 Redis 连接

var authTestDB *gorm.DB
var authTestRedis *redis.Client
var authTestWorkspaceID uuid.UUID

func setupAuthTest(t *testing.T) (AuthService, context.Context, func()) {
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
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("无法连接 Redis")
	}

	// 创建测试工作区
	var workspace model.Workspace
	result := db.Where("name = ?", "Auth Test Workspace").First(&workspace)
	if result.Error == gorm.ErrRecordNotFound {
		workspace = model.Workspace{
			Name: "Auth Test Workspace",
			Slug: "auth-test-workspace",
		}
		db.Create(&workspace)
	}

	authTestDB = db
	authTestRedis = rdb
	authTestWorkspaceID = workspace.ID

	// 创建服务
	cfg := &config.Config{
		JWTSecret:        "auth-test-secret-key",
		JWTAccessExpiry:  15 * 60 * 1000000000, // 15 分钟
		JWTRefreshExpiry: 7 * 24 * 60 * 60 * 1000000000, // 7 天
	}

	userStore := store.NewUserStore(db)
	jwtService := NewJWTService(cfg)
	authService := NewAuthService(userStore, jwtService, rdb, cfg)

	// 返回清理函数
	cleanup := func() {
		// 清理测试数据（使用事务内创建的用户会被自动回滚）
	}

	return authService, ctx, cleanup
}

// =============================================================================
// Register 测试
// =============================================================================

func TestAuthService_Register_Success(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]

	tests := []struct {
		name      string
		email     string
		username  string
		password  string
		fullName  string
	}{
		{
			name:      "成功注册普通用户",
			email:     prefix + "_register@example.com",
			username:  prefix + "_registeruser",
			password:  "SecurePass123!",
			fullName:  "Register Test User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, accessToken, refreshToken, err := authService.Register(ctx, authTestWorkspaceID, tt.email, tt.username, tt.password, tt.fullName)

			if err != nil {
				t.Errorf("Register() error = %v", err)
				return
			}

			if user == nil {
				t.Error("Register() user 为 nil")
				return
			}

			if user.Email != tt.email {
				t.Errorf("Register().Email = %v, want %v", user.Email, tt.email)
			}

			if user.Username != tt.username {
				t.Errorf("Register().Username = %v, want %v", user.Username, tt.username)
			}

			if accessToken == "" {
				t.Error("Register() accessToken 为空")
			}

			if refreshToken == "" {
				t.Error("Register() refreshToken 为空")
			}

			// 验证密码已哈希
			if user.PasswordHash == tt.password {
				t.Error("Register() 密码未哈希")
			}

			if user.PasswordHash == "" {
				t.Error("Register() 密码哈希为空")
			}
		})
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]
	email := prefix + "_dupemail@example.com"

	// 第一次注册
	_, _, _, err := authService.Register(ctx, authTestWorkspaceID, email, prefix+"_user1", "Password123!", "User 1")
	if err != nil {
		t.Fatalf("第一次注册失败: %v", err)
	}

	// 尝试用相同邮箱注册
	_, _, _, err = authService.Register(ctx, authTestWorkspaceID, email, prefix+"_user2", "Password123!", "User 2")
	if err == nil {
		t.Error("Register() 应该返回错误，邮箱重复")
	}
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]
	username := prefix + "_dupuser"

	// 第一次注册
	_, _, _, err := authService.Register(ctx, authTestWorkspaceID, prefix+"_email1@example.com", username, "Password123!", "User 1")
	if err != nil {
		t.Fatalf("第一次注册失败: %v", err)
	}

	// 尝试用相同用户名注册
	_, _, _, err = authService.Register(ctx, authTestWorkspaceID, prefix+"_email2@example.com", username, "Password123!", "User 2")
	if err == nil {
		t.Error("Register() 应该返回错误，用户名重复")
	}
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]

	tests := []struct {
		name     string
		password string
	}{
		{"密码太短", "abc"},
		{"只有小写字母", "abcdefgh"},
		{"只有数字", "12345678"},
		{"无数字", "abcdefghI"},
		{"无大写字母", "abcdefgh1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := authService.Register(ctx, authTestWorkspaceID, prefix+tt.name+"@example.com", prefix+tt.name, tt.password, "Test User")
			if err == nil {
				t.Errorf("Register() 应该拒绝弱密码: %s", tt.password)
			}
		})
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]

	tests := []struct {
		name  string
		email string
	}{
		{"空邮箱", ""},
		{"无@符号", "invalidemail"},
		{"无域名", "test@"},
		{"无用户名", "@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := authService.Register(ctx, authTestWorkspaceID, tt.email, prefix+tt.name, "Password123!", "Test User")
			if err == nil {
				t.Errorf("Register() 应该拒绝无效邮箱: %s", tt.email)
			}
		})
	}
}

func TestAuthService_Register_InvalidUsername(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]

	tests := []struct {
		name     string
		username string
	}{
		{"用户名太短", "ab"},
		{"包含空格", "test user"},
		{"包含特殊字符", "test@user"},
		{"只有数字", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := authService.Register(ctx, authTestWorkspaceID, prefix+tt.name+"@example.com", tt.username, "Password123!", "Test User")
			if err == nil {
				t.Errorf("Register() 应该拒绝无效用户名: %s", tt.username)
			}
		})
	}
}

// =============================================================================
// Login 测试
// =============================================================================

func TestAuthService_Login_Success(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]
	email := prefix + "_login@example.com"
	password := "Password123!"

	// 先注册用户
	_, _, _, err := authService.Register(ctx, authTestWorkspaceID, email, prefix+"_loginuser", password, "Login Test User")
	if err != nil {
		t.Fatalf("注册用户失败: %v", err)
	}

	// 登录
	user, accessToken, refreshToken, err := authService.Login(ctx, email, password)
	if err != nil {
		t.Errorf("Login() error = %v", err)
		return
	}

	if user == nil {
		t.Error("Login() user 为 nil")
		return
	}

	if user.Email != email {
		t.Errorf("Login().Email = %v, want %v", user.Email, email)
	}

	if accessToken == "" {
		t.Error("Login() accessToken 为空")
	}

	if refreshToken == "" {
		t.Error("Login() refreshToken 为空")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]
	email := prefix + "_wrongpass@example.com"
	password := "Password123!"

	// 先注册用户
	_, _, _, err := authService.Register(ctx, authTestWorkspaceID, email, prefix+"_wrongpassuser", password, "Wrong Pass User")
	if err != nil {
		t.Fatalf("注册用户失败: %v", err)
	}

	// 用错误密码登录
	_, _, _, err = authService.Login(ctx, email, "WrongPassword123!")
	if err == nil {
		t.Error("Login() 应该返回错误，密码错误")
	}
}

func TestAuthService_Login_NonExistentEmail(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	// 用不存在的邮箱登录
	_, _, _, err := authService.Login(ctx, "nonexistent@example.com", "Password123!")
	if err == nil {
		t.Error("Login() 应该返回错误，邮箱不存在")
	}
}

// =============================================================================
// RefreshToken 测试
// =============================================================================

func TestAuthService_RefreshToken_Success(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]
	email := prefix + "_refresh@example.com"
	password := "Password123!"

	// 注册用户
	_, _, refreshToken, err := authService.Register(ctx, authTestWorkspaceID, email, prefix+"_refreshuser", password, "Refresh Test User")
	if err != nil {
		t.Fatalf("注册用户失败: %v", err)
	}

	// 刷新令牌
	newAccessToken, newRefreshToken, err := authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		t.Errorf("RefreshToken() error = %v", err)
		return
	}

	if newAccessToken == "" {
		t.Error("RefreshToken() newAccessToken 为空")
	}

	if newRefreshToken == "" {
		t.Error("RefreshToken() newRefreshToken 为空")
	}

	// 新的刷新令牌应该与旧的不同（令牌轮换）
	if newRefreshToken == refreshToken {
		t.Error("RefreshToken() 应该返回新的刷新令牌（令牌轮换）")
	}
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	_, _, err := authService.RefreshToken(ctx, "invalid-refresh-token")
	if err == nil {
		t.Error("RefreshToken() 应该返回错误，令牌无效")
	}
}

func TestAuthService_RefreshToken_AccessTokenUsed(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]
	email := prefix + "_accesstoken@example.com"
	password := "Password123!"

	// 注册用户
	_, accessToken, _, err := authService.Register(ctx, authTestWorkspaceID, email, prefix+"_accesstokenuser", password, "Access Token User")
	if err != nil {
		t.Fatalf("注册用户失败: %v", err)
	}

	// 尝试用访问令牌刷新
	_, _, err = authService.RefreshToken(ctx, accessToken)
	if err == nil {
		t.Error("RefreshToken() 应该返回错误，不能用访问令牌刷新")
	}
}

// =============================================================================
// Logout 测试
// =============================================================================

func TestAuthService_Logout_Success(t *testing.T) {
	authService, ctx, cleanup := setupAuthTest(t)
	defer cleanup()

	prefix := uuid.New().String()[:8]
	email := prefix + "_logout@example.com"
	password := "Password123!"

	// 注册用户
	_, _, refreshToken, err := authService.Register(ctx, authTestWorkspaceID, email, prefix+"_logoutuser", password, "Logout Test User")
	if err != nil {
		t.Fatalf("注册用户失败: %v", err)
	}

	// 登出
	err = authService.Logout(ctx, refreshToken)
	if err != nil {
		t.Errorf("Logout() error = %v", err)
		return
	}

	// 尝试用已登出的刷新令牌刷新（应该失败）
	_, _, err = authService.RefreshToken(ctx, refreshToken)
	if err == nil {
		t.Error("RefreshToken() 应该返回错误，令牌已登出")
	}
}
