// Package middleware 提供 HTTP 中间件
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole 角色检查中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetCurrentUserRole(c)
		if userRole == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		// 检查角色是否在允许列表中
		allowed := false
		for _, role := range roles {
			if userRole == role {
				allowed = true
				break
			}
			// 全局管理员拥有所有权限
			if userRole == "global_admin" {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 管理员权限中间件（admin 或 global_admin）
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin", "global_admin")
}

// RequireGlobalAdmin 全局管理员权限中间件（仅 global_admin）
func RequireGlobalAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsGlobalAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "需要全局管理员权限",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
