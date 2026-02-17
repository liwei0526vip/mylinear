// Package handler 提供 HTTP 处理器
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylinear/server/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Password    string `json:"password" binding:"required,min=8"`
	Name        string `json:"name" binding:"required"`
	WorkspaceID string `json:"workspace_id"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         *UserDTO `json:"user"`
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Role        string `json:"role"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 验证密码强度
	if !isValidPassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "密码强度不足，需要至少8个字符，包含大小写字母和数字",
		})
		return
	}

	// 解析工作区 ID
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "工作区 ID 无效",
		})
		return
	}

	user, accessToken, refreshToken, err := h.authService.Register(c.Request.Context(), workspaceID, req.Email, req.Username, req.Password, req.Name)
	if err != nil {
		switch err.Error() {
		case "邮箱已被注册", "用户名已被使用":
			c.JSON(http.StatusConflict, gin.H{
				"error":   "conflict",
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "注册失败",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User: &UserDTO{
				ID:          user.ID.String(),
				WorkspaceID: user.WorkspaceID.String(),
				Email:       user.Email,
				Username:    user.Username,
				Name:        user.Name,
				Role:        string(user.Role),
			},
		},
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "请求参数无效",
		})
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "邮箱或密码错误" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "登录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User: &UserDTO{
				ID:          user.ID.String(),
				WorkspaceID: user.WorkspaceID.String(),
				Email:       user.Email,
				Username:    user.Username,
				Name:        user.Name,
				Role:        string(user.Role),
			},
		},
	})
}

// Refresh 刷新令牌
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "请求参数无效",
		})
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		errMsg := err.Error()
		// 认证相关错误返回 401
		if isAuthError(errMsg) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": errMsg,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "刷新令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

// isAuthError 检查是否为认证错误
func isAuthError(errMsg string) bool {
	authErrors := []string{
		"令牌验证失败",
		"令牌已失效",
		"无效的令牌类型",
		"用户不存在",
		"无效的用户ID",
	}
	for _, e := range authErrors {
		if strings.Contains(errMsg, e) {
			return true
		}
	}
	return false
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "请求参数无效",
		})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "登出失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "登出成功",
		},
	})
}

// isValidPassword 验证密码强度
func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}
