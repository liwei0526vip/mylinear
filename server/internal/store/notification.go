// Package store 提供数据访问层
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// ListNotificationsOptions 通知列表查询选项
type ListNotificationsOptions struct {
	Page     int                   // 页码，从 1 开始
	PageSize int                   // 每页数量，默认 20
	Read     *bool                 // 按已读状态过滤（nil 表示全部）
	Types    []model.NotificationType // 按类型过滤
}

// NotificationStore 定义通知数据访问接口
type NotificationStore interface {
	// CreateNotification 创建通知
	CreateNotification(ctx context.Context, notification *model.Notification) error
	// GetNotificationByID 根据 ID 获取通知
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*model.Notification, error)
	// ListNotifications 获取用户的通知列表
	ListNotifications(ctx context.Context, userID uuid.UUID, opts *ListNotificationsOptions) ([]model.Notification, int64, error)
	// CountUnread 统计用户未读通知数量
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkAsRead 标记单条通知为已读
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	// MarkAllAsRead 标记用户所有通知为已读
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkBatchAsRead 批量标记通知为已读
	MarkBatchAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error)
}

// notificationStore 实现 NotificationStore 接口
type notificationStore struct {
	db *gorm.DB
}

// NewNotificationStore 创建通知存储实例
func NewNotificationStore(db *gorm.DB) NotificationStore {
	return &notificationStore{db: db}
}

// CreateNotification 创建通知
func (s *notificationStore) CreateNotification(ctx context.Context, notification *model.Notification) error {
	if notification == nil {
		return fmt.Errorf("notification 不能为 nil")
	}

	if err := s.db.WithContext(ctx).Create(notification).Error; err != nil {
		return fmt.Errorf("创建通知失败: %w", err)
	}

	return nil
}

// GetNotificationByID 根据 ID 获取通知
func (s *notificationStore) GetNotificationByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	var notification model.Notification

	err := s.db.WithContext(ctx).
		Preload("User").
		First(&notification, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到通知: %w", err)
		}
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}

	return &notification, nil
}

// ListNotifications 获取用户的通知列表
func (s *notificationStore) ListNotifications(ctx context.Context, userID uuid.UUID, opts *ListNotificationsOptions) ([]model.Notification, int64, error) {
	query := s.db.WithContext(ctx).
		Where("user_id = ?", userID)

	// 应用已读状态过滤
	if opts != nil && opts.Read != nil {
		if *opts.Read {
			query = query.Where("read_at IS NOT NULL")
		} else {
			query = query.Where("read_at IS NULL")
		}
	}

	// 应用类型过滤
	if opts != nil && len(opts.Types) > 0 {
		query = query.Where("type IN ?", opts.Types)
	}

	// 统计总数
	var total int64
	if err := query.Model(&model.Notification{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计通知数量失败: %w", err)
	}

	// 未读优先排序：未读在前，已读在后；同组内按创建时间倒序
	query = query.Order("CASE WHEN read_at IS NULL THEN 0 ELSE 1 END, created_at DESC")

	// 应用分页
	if opts != nil && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		if offset < 0 {
			offset = 0
		}
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	var notifications []model.Notification
	if err := query.Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("查询通知列表失败: %w", err)
	}

	return notifications, total, nil
}

// CountUnread 统计用户未读通知数量
func (s *notificationStore) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计未读通知数量失败: %w", err)
	}
	return count, nil
}

// MarkAsRead 标记单条通知为已读
func (s *notificationStore) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read_at", now)

	if result.Error != nil {
		return fmt.Errorf("标记通知已读失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到通知或无权操作")
	}

	return nil
}

// MarkAllAsRead 标记用户所有通知为已读
func (s *notificationStore) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now)

	if result.Error != nil {
		return 0, fmt.Errorf("标记所有通知已读失败: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// MarkBatchAsRead 批量标记通知为已读
func (s *notificationStore) MarkBatchAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id IN ? AND user_id = ?", ids, userID).
		Update("read_at", now)

	if result.Error != nil {
		return 0, fmt.Errorf("批量标记通知已读失败: %w", result.Error)
	}

	return result.RowsAffected, nil
}
