package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mylinear/server/internal/config"
	"github.com/mylinear/server/internal/model"
)

// TestJWTService_Interface 测试 JWTService 接口定义存在
func TestJWTService_Interface(t *testing.T) {
	var _ JWTService = (*jwtService)(nil)
}

// TestJWTService_NewJWTService 测试构造函数
func TestJWTService_NewJWTService(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "test-secret-key",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	service := NewJWTService(cfg)
	if service == nil {
		t.Error("NewJWTService() 返回 nil")
	}
}

// TestJWTService_GenerateAccessToken 测试生成访问令牌
func TestJWTService_GenerateAccessToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "test-secret-key-for-testing",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(cfg)

	userID := uuid.New()
	email := "test@example.com"
	role := model.RoleMember

	tests := []struct {
		name  string
		userID uuid.UUID
		email  string
		role   model.Role
	}{
		{
			name:   "成员用户",
			userID: userID,
			email:  email,
			role:   role,
		},
		{
			name:   "管理员用户",
			userID: uuid.New(),
			email:  "admin@example.com",
			role:   model.RoleAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateAccessToken(tt.userID, tt.email, tt.role)
			if err != nil {
				t.Errorf("GenerateAccessToken() error = %v", err)
				return
			}
			if token == "" {
				t.Error("GenerateAccessToken() 返回空令牌")
			}
		})
	}
}

// TestJWTService_GenerateAccessToken_Claims 测试访问令牌包含正确的 claims
func TestJWTService_GenerateAccessToken_Claims(t *testing.T) {
	secret := "test-secret-key-for-claims-testing"
	cfg := &config.Config{
		JWTSecret:        secret,
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(cfg)

	userID := uuid.New()
	email := "claims@example.com"
	role := model.RoleAdmin

	tokenString, err := service.GenerateAccessToken(userID, email, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// 解析令牌验证 claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("解析令牌失败: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("无法获取 claims")
	}

	// 验证 sub (用户ID)
	if claims["sub"] != userID.String() {
		t.Errorf("sub = %v, want %v", claims["sub"], userID.String())
	}

	// 验证 email
	if claims["email"] != email {
		t.Errorf("email = %v, want %v", claims["email"], email)
	}

	// 验证 role
	if claims["role"] != string(role) {
		t.Errorf("role = %v, want %v", claims["role"], role)
	}

	// 验证 exp 存在
	if claims["exp"] == nil {
		t.Error("缺少 exp claim")
	}

	// 验证 iat 存在
	if claims["iat"] == nil {
		t.Error("缺少 iat claim")
	}

	// 验证 jti 存在
	if claims["jti"] == nil {
		t.Error("缺少 jti claim")
	}
}

// TestJWTService_GenerateRefreshToken 测试生成刷新令牌
func TestJWTService_GenerateRefreshToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "test-secret-key-for-refresh",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(cfg)

	userID := uuid.New()

	token, err := service.GenerateRefreshToken(userID)
	if err != nil {
		t.Errorf("GenerateRefreshToken() error = %v", err)
		return
	}
	if token == "" {
		t.Error("GenerateRefreshToken() 返回空令牌")
	}
}

// TestJWTService_GenerateRefreshToken_Claims 测试刷新令牌包含正确的 claims
func TestJWTService_GenerateRefreshToken_Claims(t *testing.T) {
	secret := "test-secret-for-refresh-claims"
	cfg := &config.Config{
		JWTSecret:        secret,
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(cfg)

	userID := uuid.New()

	tokenString, err := service.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	// 解析令牌验证 claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("解析令牌失败: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("无法获取 claims")
	}

	// 验证 sub (用户ID)
	if claims["sub"] != userID.String() {
		t.Errorf("sub = %v, want %v", claims["sub"], userID.String())
	}

	// 验证 type = refresh
	if claims["type"] != "refresh" {
		t.Errorf("type = %v, want refresh", claims["type"])
	}

	// 验证 exp 存在
	if claims["exp"] == nil {
		t.Error("缺少 exp claim")
	}

	// 验证 jti 存在
	if claims["jti"] == nil {
		t.Error("缺少 jti claim")
	}
}

// TestJWTService_ValidateToken 测试令牌验证
func TestJWTService_ValidateToken(t *testing.T) {
	secret := "test-secret-for-validation"
	cfg := &config.Config{
		JWTSecret:        secret,
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(cfg)

	userID := uuid.New()
	email := "validate@example.com"
	role := model.RoleMember

	// 生成有效令牌
	validToken, _ := service.GenerateAccessToken(userID, email, role)

	tests := []struct {
		name      string
		token     string
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "有效令牌",
			token:     validToken,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "无效格式",
			token:     "invalid-token-format",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "空令牌",
			token:     "",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "错误签名",
			token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantValid && claims == nil {
				t.Error("ValidateToken() 应该返回有效的 claims")
			}
			if !tt.wantValid && claims != nil {
				t.Error("ValidateToken() 不应该返回 claims")
			}
		})
	}
}

// TestJWTService_GetTokenClaims 测试获取令牌 claims
func TestJWTService_GetTokenClaims(t *testing.T) {
	secret := "test-secret-for-getclaims"
	cfg := &config.Config{
		JWTSecret:        secret,
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(cfg)

	userID := uuid.New()
	email := "getclaims@example.com"
	role := model.RoleAdmin

	tokenString, _ := service.GenerateAccessToken(userID, email, role)

	claims, err := service.GetTokenClaims(tokenString)
	if err != nil {
		t.Fatalf("GetTokenClaims() error = %v", err)
	}

	if claims.UserID != userID.String() {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID.String())
	}
	if claims.Email != email {
		t.Errorf("Email = %v, want %v", claims.Email, email)
	}
	if claims.Role != string(role) {
		t.Errorf("Role = %v, want %v", claims.Role, role)
	}
}
