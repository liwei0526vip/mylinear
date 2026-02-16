package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IssueStatusHistory Issue 状态变更历史模型
type IssueStatusHistory struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IssueID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"issue_id"`
	FromStatusID *uuid.UUID `gorm:"type:uuid" json:"from_status_id,omitempty"`
	ToStatusID   uuid.UUID  `gorm:"type:uuid;not null" json:"to_status_id"`
	ChangedByID  uuid.UUID  `gorm:"type:uuid;not null" json:"changed_by_id"`
	ChangedAt    time.Time  `gorm:"not null;default:now();index" json:"changed_at"`

	// 关联关系
	Issue     *Issue         `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"issue,omitempty"`
	FromState *WorkflowState `gorm:"foreignKey:FromStatusID;constraint:OnDelete:SET NULL" json:"from_state,omitempty"`
	ToState   *WorkflowState `gorm:"foreignKey:ToStatusID;constraint:OnDelete:RESTRICT" json:"to_state,omitempty"`
	ChangedBy *User          `gorm:"foreignKey:ChangedByID;constraint:OnDelete:RESTRICT" json:"changed_by,omitempty"`
}

// TableName 指定表名
func (IssueStatusHistory) TableName() string {
	return "issue_status_history"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (h *IssueStatusHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

// IsInitialTransition 检查是否为初始状态转换（没有 from_state）
func (h *IssueStatusHistory) IsInitialTransition() bool {
	return h.FromStatusID == nil
}
