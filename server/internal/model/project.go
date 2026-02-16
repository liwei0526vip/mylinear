package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Project 项目模型
type Project struct {
	ModelWithSoftDelete
	WorkspaceID uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	Status      ProjectStatus  `gorm:"type:varchar(20);not null;default:'planned';index" json:"status"`
	Priority    int            `gorm:"not null;default:0" json:"priority"`
	LeadID      *uuid.UUID     `gorm:"type:uuid;index" json:"lead_id,omitempty"`
	StartDate   *time.Time     `gorm:"type:date" json:"start_date,omitempty"`
	TargetDate  *time.Time     `gorm:"type:date" json:"target_date,omitempty"`
	Teams       pq.StringArray `gorm:"type:uuid[];default:'{}'" json:"teams"`
	Labels      pq.StringArray `gorm:"type:uuid[];default:'{}'" json:"labels"`
	CompletedAt *time.Time     `gorm:"type:timestamptz" json:"completed_at,omitempty"`

	// 关联关系
	Workspace  *Workspace  `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"workspace,omitempty"`
	Lead       *User       `gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL" json:"lead,omitempty"`
	Issues     []Issue     `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL" json:"issues,omitempty"`
	Milestones []Milestone `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"milestones,omitempty"`
	Documents  []Document  `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL" json:"documents,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// IsCompleted 检查项目是否已完成
func (p *Project) IsCompleted() bool {
	return p.Status == ProjectStatusCompleted || p.CompletedAt != nil
}

// IsActive 检查项目是否活跃
func (p *Project) IsActive() bool {
	return p.Status == ProjectStatusInProgress
}

// IsCancelled 检查项目是否已取消
func (p *Project) IsCancelled() bool {
	return p.Status == ProjectStatusCancelled
}
