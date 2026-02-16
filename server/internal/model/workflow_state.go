package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowState 工作流状态模型
type WorkflowState struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TeamID    uuid.UUID `gorm:"type:uuid;not null;index" json:"team_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Type      StateType `gorm:"type:varchar(20);not null;default:'backlog'" json:"type"`
	Color     string    `gorm:"type:varchar(20);default:'#808080'" json:"color"`
	Position  float64   `gorm:"not null;default:0" json:"position"`
	IsDefault bool      `gorm:"not null;default:false" json:"is_default"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`

	// 关联关系
	Team        *Team                `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"team,omitempty"`
	Issues      []Issue              `gorm:"foreignKey:StatusID;constraint:OnDelete:RESTRICT" json:"issues,omitempty"`
	Transitions []WorkflowTransition `gorm:"foreignKey:FromStateID;constraint:OnDelete:CASCADE" json:"transitions,omitempty"`
}

// TableName 指定表名
func (WorkflowState) TableName() string {
	return "workflow_states"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (s *WorkflowState) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsCompletedType 检查是否为完成类型状态
func (s *WorkflowState) IsCompletedType() bool {
	return s.Type == StateTypeCompleted
}

// IsCancelledType 检查是否为取消类型状态
func (s *WorkflowState) IsCancelledType() bool {
	return s.Type == StateTypeCancelled
}

// IsTerminalState 检查是否为终态（完成或取消）
func (s *WorkflowState) IsTerminalState() bool {
	return s.IsCompletedType() || s.IsCancelledType()
}
