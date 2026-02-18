// Package service 提供业务逻辑层
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

// NotificationService 定义通知服务接口
type NotificationService interface {
	// CreateNotification 创建通知
	CreateNotification(ctx context.Context, notification *model.Notification) error
	// NotifyIssueAssigned 通知用户被分配了 Issue
	NotifyIssueAssigned(ctx context.Context, actorID uuid.UUID, assigneeID *uuid.UUID, issueID uuid.UUID, issueTitle string) error
	// NotifyIssueMentioned 通知用户在 Issue 中被 @mention
	NotifyIssueMentioned(ctx context.Context, actorID uuid.UUID, mentionedUsernames []string, issueID uuid.UUID, issueTitle string) error
	// NotifySubscribers 通知订阅者
	NotifySubscribers(ctx context.Context, actorID uuid.UUID, subscriberIDs []uuid.UUID, notifyType model.NotificationType, issueID uuid.UUID, title string, body string) error
	// ListNotifications 获取用户的通知列表
	ListNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int, read *bool, types []model.NotificationType) ([]model.Notification, int64, error)
	// GetUnreadCount 获取未读通知数量
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkAsRead 标记单条通知为已读
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	// MarkAllAsRead 标记所有通知为已读
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkBatchAsRead 批量标记通知为已读
	MarkBatchAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error)
}

// notificationService 实现 NotificationService 接口
type notificationService struct {
	notificationStore store.NotificationStore
	preferenceStore   store.NotificationPreferenceStore
	userStore         store.UserStore
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(notificationStore store.NotificationStore, preferenceStore store.NotificationPreferenceStore, userStore store.UserStore) NotificationService {
	return &notificationService{
		notificationStore: notificationStore,
		preferenceStore:   preferenceStore,
		userStore:         userStore,
	}
}

// CreateNotification 创建通知
func (s *notificationService) CreateNotification(ctx context.Context, notification *model.Notification) error {
	if notification == nil {
		return fmt.Errorf("通知不能为 nil")
	}

	// 验证
	if notification.UserID == uuid.Nil {
		return fmt.Errorf("用户 ID 不能为空")
	}

	if notification.Title == "" {
		return fmt.Errorf("标题不能为空")
	}

	if !notification.Type.Valid() {
		return fmt.Errorf("无效的通知类型: %s", notification.Type)
	}

	return s.notificationStore.CreateNotification(ctx, notification)
}

// NotifyIssueAssigned 通知用户被分配了 Issue
func (s *notificationService) NotifyIssueAssigned(ctx context.Context, actorID uuid.UUID, assigneeID *uuid.UUID, issueID uuid.UUID, issueTitle string) error {
	// 取消指派不通知
	if assigneeID == nil {
		return nil
	}

	// 指派给自己不通知
	if *assigneeID == actorID {
		return nil
	}

	// 检查用户是否启用该类型通知
	enabled, err := s.preferenceStore.IsEnabled(ctx, *assigneeID, model.NotificationChannelInApp, model.NotificationTypeIssueAssigned)
	if err != nil {
		return fmt.Errorf("检查通知偏好失败: %w", err)
	}
	if !enabled {
		return nil
	}

	notification := &model.Notification{
		UserID:       *assigneeID,
		Type:         model.NotificationTypeIssueAssigned,
		Title:        fmt.Sprintf("您被分配了 Issue: %s", issueTitle),
		ResourceType: "issue",
		ResourceID:   &issueID,
	}

	return s.CreateNotification(ctx, notification)
}

// NotifyIssueMentioned 通知用户在 Issue 中被 @mention
func (s *notificationService) NotifyIssueMentioned(ctx context.Context, actorID uuid.UUID, mentionedUsernames []string, issueID uuid.UUID, issueTitle string) error {
	if len(mentionedUsernames) == 0 {
		return nil
	}

	// 检查用户是否启用该类型通知
	enabled, err := s.preferenceStore.IsEnabled(ctx, uuid.Nil, model.NotificationChannelInApp, model.NotificationTypeIssueMentioned)
	if err != nil {
		return fmt.Errorf("检查通知偏好失败: %w", err)
	}
	_ = enabled // 暂时不使用，因为每个用户可能有不同的偏好

	for _, username := range mentionedUsernames {
		// 根据 username 查找用户
		user, err := s.userStore.GetUserByUsername(ctx, username)
		if err != nil {
			// 用户不存在，忽略
			continue
		}

		// 自己 mention 自己不通知
		if user.ID == actorID {
			continue
		}

		// 检查用户是否启用该类型通知
		userEnabled, err := s.preferenceStore.IsEnabled(ctx, user.ID, model.NotificationChannelInApp, model.NotificationTypeIssueMentioned)
		if err != nil {
			continue
		}
		if !userEnabled {
			continue
		}

		notification := &model.Notification{
			UserID:       user.ID,
			Type:         model.NotificationTypeIssueMentioned,
			Title:        fmt.Sprintf("您在 Issue 中被提及: %s", issueTitle),
			ResourceType: "issue",
			ResourceID:   &issueID,
		}

		if err := s.CreateNotification(ctx, notification); err != nil {
			// 单个通知创建失败不影响其他通知
			continue
		}
	}

	return nil
}

// NotifySubscribers 通知订阅者
func (s *notificationService) NotifySubscribers(ctx context.Context, actorID uuid.UUID, subscriberIDs []uuid.UUID, notifyType model.NotificationType, issueID uuid.UUID, title string, body string) error {
	if len(subscriberIDs) == 0 {
		return nil
	}

	for _, subscriberID := range subscriberIDs {
		// 排除操作者本人
		if subscriberID == actorID {
			continue
		}

		// 检查用户是否启用该类型通知
		enabled, err := s.preferenceStore.IsEnabled(ctx, subscriberID, model.NotificationChannelInApp, notifyType)
		if err != nil {
			continue
		}
		if !enabled {
			continue
		}

		notification := &model.Notification{
			UserID:       subscriberID,
			Type:         notifyType,
			Title:        title,
			Body:         &body,
			ResourceType: "issue",
			ResourceID:   &issueID,
		}

		if err := s.CreateNotification(ctx, notification); err != nil {
			// 单个通知创建失败不影响其他通知
			continue
		}
	}

	return nil
}

// ListNotifications 获取用户的通知列表
func (s *notificationService) ListNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int, read *bool, types []model.NotificationType) ([]model.Notification, int64, error) {
	opts := &store.ListNotificationsOptions{
		Page:     page,
		PageSize: pageSize,
		Read:     read,
		Types:    types,
	}

	return s.notificationStore.ListNotifications(ctx, userID, opts)
}

// GetUnreadCount 获取未读通知数量
func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationStore.CountUnread(ctx, userID)
}

// MarkAsRead 标记单条通知为已读
func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.notificationStore.MarkAsRead(ctx, id, userID)
}

// MarkAllAsRead 标记所有通知为已读
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationStore.MarkAllAsRead(ctx, userID)
}

// MarkBatchAsRead 批量标记通知为已读
func (s *notificationService) MarkBatchAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error) {
	return s.notificationStore.MarkBatchAsRead(ctx, ids, userID)
}

// =============================================================================
// NotificationPreferenceService
// =============================================================================

// NotificationPreferenceUpdate 通知偏好更新参数
type NotificationPreferenceUpdate struct {
	Channel model.NotificationChannel
	Type    model.NotificationType
	Enabled bool
}

// NotificationPreferenceService 定义通知偏好配置服务接口
type NotificationPreferenceService interface {
	// GetPreferences 获取用户的通知偏好配置
	GetPreferences(ctx context.Context, userID uuid.UUID, channel *model.NotificationChannel) ([]model.NotificationPreference, error)
	// UpdatePreferences 更新用户的通知偏好配置
	UpdatePreferences(ctx context.Context, userID uuid.UUID, updates []NotificationPreferenceUpdate) error
}

// notificationPreferenceService 实现 NotificationPreferenceService 接口
type notificationPreferenceService struct {
	preferenceStore store.NotificationPreferenceStore
}

// NewNotificationPreferenceService 创建通知偏好配置服务实例
func NewNotificationPreferenceService(preferenceStore store.NotificationPreferenceStore) NotificationPreferenceService {
	return &notificationPreferenceService{
		preferenceStore: preferenceStore,
	}
}

// GetPreferences 获取用户的通知偏好配置
func (s *notificationPreferenceService) GetPreferences(ctx context.Context, userID uuid.UUID, channel *model.NotificationChannel) ([]model.NotificationPreference, error) {
	prefs, err := s.preferenceStore.GetByUser(ctx, userID, channel)
	if err != nil {
		return nil, fmt.Errorf("获取通知偏好配置失败: %w", err)
	}

	// 如果没有配置，返回默认值
	if len(prefs) == 0 && (channel == nil || *channel == model.NotificationChannelInApp) {
		return model.DefaultPreferences(userID), nil
	}

	return prefs, nil
}

// UpdatePreferences 更新用户的通知偏好配置
func (s *notificationPreferenceService) UpdatePreferences(ctx context.Context, userID uuid.UUID, updates []NotificationPreferenceUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	for _, update := range updates {
		// MVP 阶段仅支持 in_app 渠道
		if update.Channel != model.NotificationChannelInApp {
			return fmt.Errorf("MVP 阶段仅支持 in_app 渠道")
		}

		pref := &model.NotificationPreference{
			UserID:  userID,
			Channel: update.Channel,
			Type:    update.Type,
			Enabled: &update.Enabled,
		}

		if err := s.preferenceStore.Upsert(ctx, pref); err != nil {
			return fmt.Errorf("更新通知偏好配置失败: %w", err)
		}
	}

	return nil
}
