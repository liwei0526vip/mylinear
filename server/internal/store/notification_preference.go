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

// NotificationPreferenceStore 定义通知偏好配置数据访问接口
type NotificationPreferenceStore interface {
	// GetByUser 获取用户的通知偏好配置
	GetByUser(ctx context.Context, userID uuid.UUID, channel *model.NotificationChannel) ([]model.NotificationPreference, error)
	// Upsert 创建或更新通知偏好配置
	Upsert(ctx context.Context, preference *model.NotificationPreference) error
	// BatchUpsert 批量创建或更新通知偏好配置
	BatchUpsert(ctx context.Context, preferences []model.NotificationPreference) error
	// IsEnabled 检查用户是否启用了某类型的通知
	IsEnabled(ctx context.Context, userID uuid.UUID, channel model.NotificationChannel, notifyType model.NotificationType) (bool, error)
}

// notificationPreferenceStore 实现 NotificationPreferenceStore 接口
type notificationPreferenceStore struct {
	db *gorm.DB
}

// NewNotificationPreferenceStore 创建通知偏好配置存储实例
func NewNotificationPreferenceStore(db *gorm.DB) NotificationPreferenceStore {
	return &notificationPreferenceStore{db: db}
}

// GetByUser 获取用户的通知偏好配置
func (s *notificationPreferenceStore) GetByUser(ctx context.Context, userID uuid.UUID, channel *model.NotificationChannel) ([]model.NotificationPreference, error) {
	query := s.db.WithContext(ctx).
		Where("user_id = ?", userID)

	// 按渠道过滤
	if channel != nil {
		query = query.Where("channel = ?", *channel)
	}

	var preferences []model.NotificationPreference
	if err := query.Find(&preferences).Error; err != nil {
		return nil, fmt.Errorf("查询通知偏好配置失败: %w", err)
	}

	return preferences, nil
}

// Upsert 创建或更新通知偏好配置
func (s *notificationPreferenceStore) Upsert(ctx context.Context, preference *model.NotificationPreference) error {
	if preference == nil {
		return fmt.Errorf("preference 不能为 nil")
	}

	// 先尝试查找现有记录
	var existing model.NotificationPreference
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND channel = ? AND type = ?", preference.UserID, preference.Channel, preference.Type).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		if err := s.db.WithContext(ctx).Create(preference).Error; err != nil {
			return fmt.Errorf("创建通知偏好配置失败: %w", err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("查询通知偏好配置失败: %w", err)
	}

	// 存在，更新记录
	enabled := true
	if preference.Enabled != nil {
		enabled = *preference.Enabled
	}
	if err := s.db.WithContext(ctx).
		Model(&existing).
		Updates(map[string]interface{}{
			"enabled":    enabled,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("更新通知偏好配置失败: %w", err)
	}

	// 更新传入的 preference 的 ID
	preference.ID = existing.ID
	return nil
}

// BatchUpsert 批量创建或更新通知偏好配置
func (s *notificationPreferenceStore) BatchUpsert(ctx context.Context, preferences []model.NotificationPreference) error {
	if len(preferences) == 0 {
		return nil
	}

	// 逐个处理（简化实现，可后续优化为批量 SQL）
	for i := range preferences {
		if err := s.Upsert(ctx, &preferences[i]); err != nil {
			return fmt.Errorf("批量创建或更新通知偏好配置失败: %w", err)
		}
	}

	return nil
}

// IsEnabled 检查用户是否启用了某类型的通知
func (s *notificationPreferenceStore) IsEnabled(ctx context.Context, userID uuid.UUID, channel model.NotificationChannel, notifyType model.NotificationType) (bool, error) {
	var preference model.NotificationPreference

	err := s.db.WithContext(ctx).
		Where("user_id = ? AND channel = ? AND type = ?", userID, channel, notifyType).
		First(&preference).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 无配置时默认启用
			return true, nil
		}
		return false, fmt.Errorf("查询通知偏好配置失败: %w", err)
	}

	if preference.Enabled == nil {
		return true, nil // nil 表示默认启用
	}
	return *preference.Enabled, nil
}
