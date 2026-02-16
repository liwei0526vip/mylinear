package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowTransition 工作流转换规则模型
type WorkflowTransition struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TeamID       uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_workflow_transition_unique" json:"team_id"`
	FromStateID  *uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_workflow_transition_unique" json:"from_state_id,omitempty"`
	ToStateID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_workflow_transition_unique" json:"to_state_id"`
	IsAllowed    bool      `gorm:"not null;default:true" json:"is_allowed"`
	CreatedAt    time.Time `gorm:"not null;default:now()" json:"created_at"`

	// 关联关系
	Team      *Team          `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"team,omitempty"`
	FromState *WorkflowState `gorm:"foreignKey:FromStateID;constraint:OnDelete:CASCADE" json:"from_state,omitempty"`
	ToState   *WorkflowState `gorm:"foreignKey:ToStateID;constraint:OnDelete:CASCADE" json:"to_state,omitempty"`
}

// TableName 指定表名
func (WorkflowTransition) TableName() string {
	return "workflow_transitions"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (t *WorkflowTransition) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// IsInitialTransition 检查是否为初始转换（from_state 为空）
func (t *WorkflowTransition) IsInitialTransition() bool {
	return t.FromStateID == nil
}
