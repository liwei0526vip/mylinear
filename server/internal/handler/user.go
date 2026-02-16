// Package handler 提供 HTTP 处理器
package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mylinear/server/internal/middleware"
	"github.com/mylinear/server/internal/service"
)

// 文件上传配置
const (
	MaxAvatarSize = 2 * 1024 * 1024 // 2MB
)

// 允许的图片类型
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// UserHandler 用户处理器
type UserHandler struct {
	userService   service.UserService
	avatarService service.AvatarService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService:   userService,
		avatarService: nil,
	}
}

// NewUserHandlerWithAvatar 创建带头像服务的用户处理器
func NewUserHandlerWithAvatar(userService service.UserService, avatarService service.AvatarService) *UserHandler {
	return &UserHandler{
		userService:   userService,
		avatarService: avatarService,
	}
}

// UpdateMeRequest 更新当前用户请求
type UpdateMeRequest struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Name     *string `json:"name"`
}

// GetMe 获取当前用户信息
func (h *UserHandler) GetMe(c *gin.Context) {
	userCtx := middleware.GetCurrentUser(c)
	if userCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未认证",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userCtx.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "获取用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": UserDTO{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Role:     string(user.Role),
		},
	})
}

// UpdateMe 更新当前用户信息
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userCtx := middleware.GetCurrentUser(c)
	if userCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未认证",
		})
		return
	}

	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "请求参数无效",
		})
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "没有需要更新的字段",
		})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userCtx.UserID, updates)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "邮箱已被使用" || errMsg == "用户名已被使用" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "conflict",
				"message": errMsg,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "更新用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": UserDTO{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Role:     string(user.Role),
		},
	})
}

// UploadAvatar 上传头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userCtx := middleware.GetCurrentUser(c)
	if userCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未认证",
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "请选择要上传的头像文件",
		})
		return
	}
	defer file.Close()

	// 验证文件
	contentType := header.Header.Get("Content-Type")
	if err := validateAvatarFile(contentType, header.Size); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": err.Error(),
		})
		return
	}

	// 检查是否有 AvatarService
	if h.avatarService == nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "not_implemented",
			"message": "头像上传服务未配置",
		})
		return
	}

	// 上传到存储服务
	avatarURL, err := h.avatarService.UploadAvatar(c.Request.Context(), userCtx.UserID, file, header.Filename, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "头像上传失败",
		})
		return
	}

	// 更新用户头像 URL
	_, err = h.userService.UpdateUser(c.Request.Context(), userCtx.UserID, map[string]interface{}{
		"avatar_url": avatarURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "更新头像失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"avatar_url": avatarURL,
		},
	})
}

// validateAvatarFile 验证头像文件
func validateAvatarFile(contentType string, size int64) error {
	// 检查文件大小
	if size == 0 {
		return errors.New("文件为空")
	}
	if size > MaxAvatarSize {
		return fmt.Errorf("文件大小超过限制（最大 %dMB）", MaxAvatarSize/1024/1024)
	}

	// 检查文件类型
	if !allowedImageTypes[contentType] {
		return errors.New("不支持的文件类型，仅支持 JPG、PNG、GIF、WebP")
	}

	return nil
}
