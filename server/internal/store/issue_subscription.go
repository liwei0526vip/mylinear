// Package store 提供数据访问层
package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// IssueSubscriptionStore 定义 Issue 订阅数据访问接口
type IssueSubscriptionStore interface {
	// Subscribe 订阅 Issue
	Subscribe(ctx context.Context, issueID, userID uuid.UUID) error
	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, issueID, userID uuid.UUID) error
	// ListSubscribers 获取 Issue 的订阅者列表
	ListSubscribers(ctx context.Context, issueID uuid.UUID) ([]model.User, error)
	// IsSubscribed 检查用户是否订阅了 Issue
	IsSubscribed(ctx context.Context, issueID, userID uuid.UUID) (bool, error)
}

// issueSubscriptionStore 实现 IssueSubscriptionStore 接口
type issueSubscriptionStore struct {
	db *gorm.DB
}

// NewIssueSubscriptionStore 创建 Issue 订阅存储实例
func NewIssueSubscriptionStore(db *gorm.DB) IssueSubscriptionStore {
	return &issueSubscriptionStore{db: db}
}

// Subscribe 订阅 Issue（幂等操作）
func (s *issueSubscriptionStore) Subscribe(ctx context.Context, issueID, userID uuid.UUID) error {
	subscription := &model.IssueSubscription{
		IssueID: issueID,
		UserID:  userID,
	}

	// 使用 GORM 的 FirstOrCreate 实现幂等
	result := s.db.WithContext(ctx).
		Where("issue_id = ? AND user_id = ?", issueID, userID).
		FirstOrCreate(subscription)
	if result.Error != nil {
		return fmt.Errorf("订阅 Issue 失败: %w", result.Error)
	}

	return nil
}

// Unsubscribe 取消订阅
func (s *issueSubscriptionStore) Unsubscribe(ctx context.Context, issueID, userID uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("issue_id = ? AND user_id = ?", issueID, userID).
		Delete(&model.IssueSubscription{})
	if result.Error != nil {
		return fmt.Errorf("取消订阅失败: %w", result.Error)
	}

	return nil
}

// ListSubscribers 获取 Issue 的订阅者列表
func (s *issueSubscriptionStore) ListSubscribers(ctx context.Context, issueID uuid.UUID) ([]model.User, error) {
	var users []model.User

	err := s.db.WithContext(ctx).
		Joins("JOIN issue_subscriptions ON users.id = issue_subscriptions.user_id").
		Where("issue_subscriptions.issue_id = ?", issueID).
		Order("issue_subscriptions.created_at ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("查询订阅者列表失败: %w", err)
	}

	return users, nil
}

// IsSubscribed 检查用户是否订阅了 Issue
func (s *issueSubscriptionStore) IsSubscribed(ctx context.Context, issueID, userID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&model.IssueSubscription{}).
		Where("issue_id = ? AND user_id = ?", issueID, userID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查订阅状态失败: %w", err)
	}

	return count > 0, nil
}
