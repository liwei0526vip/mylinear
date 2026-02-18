package model

import (
	"time"

	"github.com/google/uuid"
)

// IssueSubscription Issue 订阅关系模型
// 使用复合主键 (issue_id, user_id)
type IssueSubscription struct {
	IssueID   uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"issue_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:NOW()" json:"created_at"`

	// 关联关系
	Issue *Issue `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"issue,omitempty"`
	User  *User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (IssueSubscription) TableName() string {
	return "issue_subscriptions"
}
