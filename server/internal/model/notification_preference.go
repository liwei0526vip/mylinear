package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app" // 应用内
	NotificationChannelEmail NotificationChannel = "email"  // 邮件
	NotificationChannelSlack NotificationChannel = "slack"  // Slack
)

// Valid 验证通知渠道是否有效
func (c NotificationChannel) Valid() bool {
	switch c {
	case NotificationChannelInApp, NotificationChannelEmail, NotificationChannelSlack:
		return true
	default:
		return false
	}
}

// NotificationPreference 用户通知偏好配置
type NotificationPreference struct {
	ID        uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID          `gorm:"type:uuid;not null;uniqueIndex:idx_user_channel_type" json:"user_id"`
	Channel   NotificationChannel `gorm:"type:varchar(20);not null;default:'in_app';uniqueIndex:idx_user_channel_type" json:"channel"`
	Type      NotificationType   `gorm:"type:varchar(50);not null;uniqueIndex:idx_user_channel_type" json:"type"`
	Enabled   *bool              `gorm:"not null;default:true" json:"enabled"`
	CreatedAt time.Time          `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time          `gorm:"not null;default:now()" json:"updated_at"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (p *NotificationPreference) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate GORM 钩子，在更新记录前自动更新 UpdatedAt
func (p *NotificationPreference) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// DefaultPreferences 返回默认的通知偏好配置（全部启用）
func DefaultPreferences(userID uuid.UUID) []NotificationPreference {
	types := []NotificationType{
		NotificationTypeIssueAssigned,
		NotificationTypeIssueMentioned,
		NotificationTypeIssueCommented,
		NotificationTypeIssueStatusChanged,
		NotificationTypeIssuePriorityChanged,
	}

	trueVal := true
	preferences := make([]NotificationPreference, len(types))
	for i, t := range types {
		preferences[i] = NotificationPreference{
			UserID:  userID,
			Channel: NotificationChannelInApp,
			Type:    t,
			Enabled: &trueVal,
		}
	}
	return preferences
}
