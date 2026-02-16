package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification 通知模型
type Notification struct {
	ID           uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	Type         NotificationType `gorm:"type:varchar(50);not null" json:"type"`
	Title        string           `gorm:"type:varchar(255);not null" json:"title"`
	Body         *string          `gorm:"type:text" json:"body,omitempty"`
	ResourceType string           `gorm:"type:varchar(50)" json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID       `gorm:"type:uuid" json:"resource_id,omitempty"`
	ReadAt       *time.Time       `gorm:"type:timestamptz;index" json:"read_at,omitempty"`
	CreatedAt    time.Time        `gorm:"not null;default:now();index" json:"created_at"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// IsRead 检查通知是否已读
func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

// IsUnread 检查通知是否未读
func (n *Notification) IsUnread() bool {
	return n.ReadAt == nil
}

// MarkAsRead 标记通知为已读
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.ReadAt = &now
}

// HasResource 检查通知是否关联了资源
func (n *Notification) HasResource() bool {
	return n.ResourceID != nil && n.ResourceType != ""
}
