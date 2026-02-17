// Package service 提供业务逻辑层
package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/config"
	"github.com/liwei0526vip/mylinear/internal/model"
)

// TokenClaims 令牌声明结构
type TokenClaims struct {
	UserID string
	Email  string
	Role   string
	Type   string // "access" 或 "refresh"
	JTI    string // 令牌唯一标识符
}

// JWTService 定义 JWT 服务接口
type JWTService interface {
	GenerateAccessToken(userID uuid.UUID, email string, role model.Role) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	GetTokenClaims(tokenString string) (*TokenClaims, error)
}

// jwtService 实现 JWTService 接口
type jwtService struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTService 创建 JWT 服务实例
func NewJWTService(cfg *config.Config) JWTService {
	return &jwtService{
		secret:        []byte(cfg.JWTSecret),
		accessExpiry:  cfg.JWTAccessExpiry,
		refreshExpiry: cfg.JWTRefreshExpiry,
	}
}

// GenerateAccessToken 生成访问令牌
func (s *jwtService) GenerateAccessToken(userID uuid.UUID, email string, role model.Role) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"role":  string(role),
		"exp":   now.Add(s.accessExpiry).Unix(),
		"iat":   now.Unix(),
		"jti":   uuid.New().String(),
		"type":  "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken 生成刷新令牌
func (s *jwtService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"type": "refresh",
		"exp":  now.Add(s.refreshExpiry).Unix(),
		"iat":  now.Unix(),
		"jti":  uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken 验证令牌并返回 claims
func (s *jwtService) ValidateToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("令牌为空")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("令牌验证失败: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("令牌无效")
	}

	return s.GetTokenClaims(tokenString)
}

// GetTokenClaims 获取令牌 claims（不验证签名）
func (s *jwtService) GetTokenClaims(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("令牌为空")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("无法获取令牌声明")
	}

	result := &TokenClaims{
		JTI: getStringClaim(claims, "jti"),
	}

	if sub, ok := claims["sub"]; ok {
		result.UserID = fmt.Sprintf("%v", sub)
	}
	if email, ok := claims["email"]; ok {
		result.Email = fmt.Sprintf("%v", email)
	}
	if role, ok := claims["role"]; ok {
		result.Role = fmt.Sprintf("%v", role)
	}
	if tokenType, ok := claims["type"]; ok {
		result.Type = fmt.Sprintf("%v", tokenType)
	}

	return result, nil
}

// getStringClaim 安全获取字符串类型的 claim
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}
