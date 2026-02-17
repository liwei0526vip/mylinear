// Package store 提供数据访问层
package store

import (
	"context"

	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// UserStore 定义用户数据访问接口
type UserStore interface {
	// CreateUser 创建新用户
	CreateUser(ctx context.Context, user *model.User) error
	// GetUserByEmail 通过邮箱获取用户
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	// GetUserByID 通过 ID 获取用户
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	// GetUserByUsername 通过用户名获取用户
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, user *model.User) error
}

// userStore 实现 UserStore 接口
type userStore struct {
	db *gorm.DB
}

// NewUserStore 创建用户存储实例
func NewUserStore(db *gorm.DB) UserStore {
	return &userStore{db: db}
}

// CreateUser 创建新用户（存根实现，后续 TDD 完善）
func (s *userStore) CreateUser(ctx context.Context, user *model.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

// GetUserByEmail 通过邮箱获取用户（存根实现，后续 TDD 完善）
func (s *userStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 通过 ID 获取用户（存根实现，后续 TDD 完善）
func (s *userStore) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 通过用户名获取用户（存根实现，后续 TDD 完善）
func (s *userStore) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息（存根实现，后续 TDD 完善）
func (s *userStore) UpdateUser(ctx context.Context, user *model.User) error {
	return s.db.WithContext(ctx).Save(user).Error
}
