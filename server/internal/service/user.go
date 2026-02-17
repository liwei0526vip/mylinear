// Package service 提供业务逻辑服务
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

// UserService 用户服务接口
type UserService interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
}

// userService 用户服务实现
type userService struct {
	userStore store.UserStore
}

// NewUserService 创建用户服务
func NewUserService(userStore store.UserStore) UserService {
	return &userService{
		userStore: userStore,
	}
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.userStore.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return user, nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
	// 先获取现有用户
	user, err := s.userStore.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 检查邮箱唯一性
	if email, ok := updates["email"].(string); ok && email != user.Email {
		existingUser, err := s.userStore.GetUserByEmail(ctx, email)
		if err == nil && existingUser.ID.String() != id {
			return nil, errors.New("邮箱已被使用")
		}
	}

	// 检查用户名唯一性
	if username, ok := updates["username"].(string); ok && username != user.Username {
		existingUser, err := s.userStore.GetUserByUsername(ctx, username)
		if err == nil && existingUser.ID.String() != id {
			return nil, errors.New("用户名已被使用")
		}
	}

	// 应用更新
	if name, ok := updates["name"].(string); ok {
		user.Name = name
	}
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if username, ok := updates["username"].(string); ok {
		user.Username = username
	}
	if avatarURL, ok := updates["avatar_url"].(string); ok {
		user.AvatarURL = &avatarURL
	}

	// 保存更新
	if err := s.userStore.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return user, nil
}
