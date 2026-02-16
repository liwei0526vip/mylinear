package model

import (
	"github.com/google/uuid"
)

// Document 文档模型
type Document struct {
	Model
	WorkspaceID uuid.UUID  `gorm:"type:uuid;not null;index" json:"workspace_id"`
	ProjectID  *uuid.UUID `gorm:"type:uuid;index" json:"project_id,omitempty"`
	IssueID    *uuid.UUID `gorm:"type:uuid;index" json:"issue_id,omitempty"`
	Title      string     `gorm:"type:varchar(500);not null" json:"title"`
	Content    *string    `gorm:"type:text" json:"content,omitempty"`
	Icon       string     `gorm:"type:varchar(50)" json:"icon,omitempty"`
	CreatedByID uuid.UUID `gorm:"type:uuid;not null;index" json:"created_by_id"`

	// 关联关系
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"workspace,omitempty"`
	Project   *Project  `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL" json:"project,omitempty"`
	Issue     *Issue    `gorm:"foreignKey:IssueID;constraint:OnDelete:SET NULL" json:"issue,omitempty"`
	CreatedBy *User     `gorm:"foreignKey:CreatedByID;constraint:OnDelete:RESTRICT" json:"created_by,omitempty"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// IsProjectDocument 检查是否为项目文档
func (d *Document) IsProjectDocument() bool {
	return d.ProjectID != nil
}

// IsIssueDocument 检查是否为 Issue 文档
func (d *Document) IsIssueDocument() bool {
	return d.IssueID != nil
}

// IsWorkspaceDocument 检查是否为工作区文档
func (d *Document) IsWorkspaceDocument() bool {
	return d.ProjectID == nil && d.IssueID == nil
}
