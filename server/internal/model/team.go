package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Team 团队模型
type Team struct {
	Model
	WorkspaceID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspace_id"`
	ParentID        *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	Key             string         `gorm:"type:varchar(10);uniqueIndex;not null" json:"key"`
	IconURL         *string        `gorm:"type:text" json:"icon_url,omitempty"`
	Timezone        string         `gorm:"type:varchar(64);not null;default:'UTC'" json:"timezone"`
	IsPrivate       bool           `gorm:"not null;default:false" json:"is_private"`
	CycleSettings   datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"cycle_settings,omitempty"`
	WorkflowSettings datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"workflow_settings,omitempty"`

	// 关联关系
	Workspace    *Workspace      `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"workspace,omitempty"`
	Parent       *Team           `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL" json:"parent,omitempty"`
	Children     []Team          `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL" json:"children,omitempty"`
	Members      []TeamMember    `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"members,omitempty"`
	Issues       []Issue         `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"issues,omitempty"`
	WorkflowStates []WorkflowState `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"workflow_states,omitempty"`
	Cycles       []Cycle         `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"cycles,omitempty"`
}

// TableName 指定表名
func (Team) TableName() string {
	return "teams"
}

// TeamMember 团队成员模型
type TeamMember struct {
	TeamID   uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"team_id"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"user_id"`
	Role     Role      `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	JoinedAt time.Time `gorm:"not null;default:now()" json:"joined_at"`

	// 关联关系
	Team *Team `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"team,omitempty"`
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (TeamMember) TableName() string {
	return "team_members"
}
