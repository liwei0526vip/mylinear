// Package middleware 提供 HTTP 中间件
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/service"
)

// 上下文键
const (
	ContextKeyUser = "user"
)

// UserContext 用户上下文信息
type UserContext struct {
	UserID string
	Email  string
	Role   string
}

// Auth JWT 认证中间件
func Auth(jwtService service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 解析 Bearer 令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "认证令牌格式无效",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证令牌
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "认证令牌无效",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		userCtx := &UserContext{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		}
		c.Set(ContextKeyUser, userCtx)

		c.Next()
	}
}

// GetCurrentUser 获取当前用户上下文
func GetCurrentUser(c *gin.Context) *UserContext {
	val, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil
	}
	user, ok := val.(*UserContext)
	if !ok {
		return nil
	}
	return user
}

// GetCurrentUserID 获取当前用户 ID
func GetCurrentUserID(c *gin.Context) uuid.UUID {
	user := GetCurrentUser(c)
	if user == nil {
		return uuid.Nil
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return uuid.Nil
	}
	return userID
}

// GetCurrentUserRole 获取当前用户角色
func GetCurrentUserRole(c *gin.Context) string {
	user := GetCurrentUser(c)
	if user == nil {
		return ""
	}
	return user.Role
}

// IsAdmin 检查当前用户是否为管理员（包括全局管理员）
func IsAdmin(c *gin.Context) bool {
	role := GetCurrentUserRole(c)
	return role == "admin" || role == "global_admin"
}

// IsGlobalAdmin 检查当前用户是否为全局管理员
func IsGlobalAdmin(c *gin.Context) bool {
	role := GetCurrentUserRole(c)
	return role == "global_admin"
}
