// Package middleware 提供 HTTP 中间件
package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylinear/server/internal/model"
	"github.com/mylinear/server/internal/store"
	"gorm.io/gorm"
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

// ContextKeyDB 数据库连接上下文键
const ContextKeyDB = "db"

// GetDB 从上下文获取数据库连接
func GetDB(c *gin.Context) *gorm.DB {
	val, exists := c.Get(ContextKeyDB)
	if !exists {
		return nil
	}
	db, ok := val.(*gorm.DB)
	if !ok {
		return nil
	}
	return db
}

// RequireTeamOwner 团队 Owner 权限中间件
// 只有团队 Owner 或 workspace Admin 可以通过
func RequireTeamOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Workspace Admin 绕过团队级别权限检查
		if IsAdmin(c) {
			c.Next()
			return
		}

		userID := GetCurrentUserID(c)
		if userID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "未认证",
			})
			c.Abort()
			return
		}

		teamIDStr := c.Param("teamId")
		if teamIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "缺少团队ID",
			})
			c.Abort()
			return
		}

		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "无效的团队ID",
			})
			c.Abort()
			return
		}

		db := GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "数据库连接不可用",
			})
			c.Abort()
			return
		}

		isOwner, err := store.IsTeamOwner(context.Background(), db, userID, teamID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "检查权限失败",
			})
			c.Abort()
			return
		}

		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "需要团队 Owner 权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTeamMember 团队成员权限中间件
// 只有团队成员可以通过
func RequireTeamMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetCurrentUserID(c)
		if userID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "未认证",
			})
			c.Abort()
			return
		}

		teamIDStr := c.Param("teamId")
		if teamIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "缺少团队ID",
			})
			c.Abort()
			return
		}

		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "无效的团队ID",
			})
			c.Abort()
			return
		}

		db := GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "数据库连接不可用",
			})
			c.Abort()
			return
		}

		// Workspace Admin 绕过团队级别权限检查
		if IsAdmin(c) {
			c.Next()
			return
		}

		isMember, err := store.IsTeamMember(context.Background(), db, userID, teamID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "检查权限失败",
			})
			c.Abort()
			return
		}

		if !isMember {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "需要团队成员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SetTeamRoleContext 将团队角色信息存入上下文（可选，用于后续处理）
func SetTeamRoleContext(c *gin.Context, role model.Role) {
	c.Set("team_role", role)
}

// GetTeamRoleContext 从上下文获取团队角色
func GetTeamRoleContext(c *gin.Context) model.Role {
	val, exists := c.Get("team_role")
	if !exists {
		return ""
	}
	role, ok := val.(model.Role)
	if !ok {
		return ""
	}
	return role
}
