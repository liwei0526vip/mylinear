package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Issue 工单模型
type Issue struct {
	Model
	TeamID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_issue_team_number" json:"team_id"`
	Number       int            `gorm:"not null;uniqueIndex:idx_issue_team_number" json:"number"`
	Title        string         `gorm:"type:varchar(500);not null" json:"title"`
	Description  *string        `gorm:"type:text" json:"description,omitempty"`
	StatusID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"status_id"`
	Priority     int            `gorm:"not null;default:0" json:"priority"`
	AssigneeID   *uuid.UUID     `gorm:"type:uuid;index" json:"assignee_id,omitempty"`
	ProjectID    *uuid.UUID     `gorm:"type:uuid;index" json:"project_id,omitempty"`
	MilestoneID  *uuid.UUID     `gorm:"type:uuid;index" json:"milestone_id,omitempty"`
	CycleID      *uuid.UUID     `gorm:"type:uuid;index" json:"cycle_id,omitempty"`
	ParentID     *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Estimate     *int           `gorm:"type:integer" json:"estimate,omitempty"`
	DueDate      *time.Time     `gorm:"type:date" json:"due_date,omitempty"`
	SLADueAt     *time.Time     `gorm:"type:timestamptz" json:"sla_due_at,omitempty"`
	Labels       pq.StringArray `gorm:"type:uuid[];default:'{}'" json:"labels"`
	CreatedByID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"created_by_id"`
	CompletedAt  *time.Time     `gorm:"type:timestamptz" json:"completed_at,omitempty"`
	CancelledAt  *time.Time     `gorm:"type:timestamptz" json:"cancelled_at,omitempty"`

	// 关联关系
	Team          *Team           `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"team,omitempty"`
	Status        *WorkflowState  `gorm:"foreignKey:StatusID;constraint:OnDelete:RESTRICT" json:"status,omitempty"`
	Assignee      *User           `gorm:"foreignKey:AssigneeID;constraint:OnDelete:SET NULL" json:"assignee,omitempty"`
	Project       *Project        `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL" json:"project,omitempty"`
	Milestone     *Milestone      `gorm:"foreignKey:MilestoneID;constraint:OnDelete:SET NULL" json:"milestone,omitempty"`
	Cycle         *Cycle          `gorm:"foreignKey:CycleID;constraint:OnDelete:SET NULL" json:"cycle,omitempty"`
	Parent        *Issue          `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL" json:"parent,omitempty"`
	Children      []Issue         `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL" json:"children,omitempty"`
	CreatedBy     *User           `gorm:"foreignKey:CreatedByID;constraint:OnDelete:RESTRICT" json:"created_by,omitempty"`
	Relations     []IssueRelation `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"relations,omitempty"`
	Comments      []Comment       `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	Attachments   []Attachment    `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"attachments,omitempty"`
	StatusHistory []IssueStatusHistory `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"status_history,omitempty"`
}

// TableName 指定表名
func (Issue) TableName() string {
	return "issues"
}

// IsCompleted 检查是否已完成
func (i *Issue) IsCompleted() bool {
	return i.CompletedAt != nil
}

// IsCancelled 检查是否已取消
func (i *Issue) IsCancelled() bool {
	return i.CancelledAt != nil
}

// IsActive 检查是否活跃（未完成且未取消）
func (i *Issue) IsActive() bool {
	return !i.IsCompleted() && !i.IsCancelled()
}

// Identifier 返回 Issue 标识符（如 ENG-123）
func (i *Issue) Identifier(teamKey string) string {
	if teamKey == "" {
		return string(rune(i.Number))
	}
	return teamKey + "-" + string(rune(i.Number))
}
