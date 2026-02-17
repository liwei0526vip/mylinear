// Package service 提供业务逻辑层
package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/config"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// 密码强度正则表达式
var (
	hasUpper   = regexp.MustCompile(`[A-Z]`)
	hasLower   = regexp.MustCompile(`[a-z]`)
	hasDigit   = regexp.MustCompile(`[0-9]`)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)
)

// AuthService 定义认证服务接口
type AuthService interface {
	Register(ctx context.Context, workspaceID uuid.UUID, email, username, password, name string) (*model.User, string, string, error)
	Login(ctx context.Context, email, password string) (*model.User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
}

// authService 实现 AuthService 接口
type authService struct {
	userStore  store.UserStore
	jwtService JWTService
	redis      *redis.Client
	cfg        *config.Config
}

// NewAuthService 创建认证服务实例
func NewAuthService(userStore store.UserStore, jwtService JWTService, redis *redis.Client, cfg *config.Config) AuthService {
	return &authService{
		userStore:  userStore,
		jwtService: jwtService,
		redis:      redis,
		cfg:        cfg,
	}
}

// Register 注册新用户
func (s *authService) Register(ctx context.Context, workspaceID uuid.UUID, email, username, password, name string) (*model.User, string, string, error) {
	// 验证邮箱格式
	if !emailRegex.MatchString(email) {
		return nil, "", "", fmt.Errorf("邮箱格式无效")
	}

	// 验证用户名格式
	if !usernameRegex.MatchString(username) {
		return nil, "", "", fmt.Errorf("用户名格式无效，只能包含字母、数字、下划线和连字符，长度3-50")
	}

	// 验证密码强度
	if err := s.validatePassword(password); err != nil {
		return nil, "", "", err
	}

	// 检查邮箱是否已存在
	existingUser, err := s.userStore.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, "", "", fmt.Errorf("邮箱已被注册")
	}

	// 检查用户名是否已存在
	existingUser, err = s.userStore.GetUserByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, "", "", fmt.Errorf("用户名已被使用")
	}

	// 哈希密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建用户
	user := &model.User{
		WorkspaceID:  workspaceID,
		Email:        email,
		Username:     username,
		Name:         name,
		PasswordHash: string(passwordHash),
		Role:         model.RoleMember,
	}

	if err := s.userStore.CreateUser(ctx, user); err != nil {
		return nil, "", "", fmt.Errorf("创建用户失败: %w", err)
	}

	// 生成令牌
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", "", fmt.Errorf("生成访问令牌失败: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, email, password string) (*model.User, string, string, error) {
	// 查找用户
	user, err := s.userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("邮箱或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", fmt.Errorf("邮箱或密码错误")
	}

	// 生成令牌
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", "", fmt.Errorf("生成访问令牌失败: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// RefreshToken 刷新令牌
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// 验证令牌
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("令牌验证失败: %w", err)
	}

	// 检查是否为刷新令牌
	if claims.Type != "refresh" {
		return "", "", fmt.Errorf("无效的令牌类型，请使用刷新令牌")
	}

	// 检查令牌是否在黑名单中
	isBlacklisted, err := s.isTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return "", "", fmt.Errorf("检查令牌状态失败: %w", err)
	}
	if isBlacklisted {
		return "", "", fmt.Errorf("令牌已失效")
	}

	// 将旧令牌加入黑名单
	if err := s.blacklistToken(ctx, claims.JTI, s.cfg.JWTRefreshExpiry); err != nil {
		return "", "", fmt.Errorf("令牌失效处理失败: %w", err)
	}

	// 获取用户信息
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("无效的用户ID")
	}

	user, err := s.userStore.GetUserByID(ctx, userID.String())
	if err != nil {
		return "", "", fmt.Errorf("用户不存在")
	}

	// 生成新令牌
	newAccessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("生成访问令牌失败: %w", err)
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	// 验证令牌
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return fmt.Errorf("令牌验证失败: %w", err)
	}

	// 将令牌加入黑名单
	if err := s.blacklistToken(ctx, claims.JTI, s.cfg.JWTRefreshExpiry); err != nil {
		return fmt.Errorf("令牌失效处理失败: %w", err)
	}

	return nil
}

// validatePassword 验证密码强度
func (s *authService) validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度至少8个字符")
	}
	if !hasUpper.MatchString(password) {
		return fmt.Errorf("密码必须包含大写字母")
	}
	if !hasLower.MatchString(password) {
		return fmt.Errorf("密码必须包含小写字母")
	}
	if !hasDigit.MatchString(password) {
		return fmt.Errorf("密码必须包含数字")
	}
	return nil
}

// isTokenBlacklisted 检查令牌是否在黑名单中
func (s *authService) isTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("token_blacklist:%s", jti)
	result, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// blacklistToken 将令牌加入黑名单
func (s *authService) blacklistToken(ctx context.Context, jti string, expiry time.Duration) error {
	key := fmt.Sprintf("token_blacklist:%s", jti)
	return s.redis.Set(ctx, key, "1", expiry).Err()
}

// NormalizeEmail 规范化邮箱地址
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
