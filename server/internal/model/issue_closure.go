package model

import (
	"github.com/google/uuid"
)

// IssueClosure Issue 层级闭包表模型
// 用于高效查询任意深度的子 Issue 层级关系
type IssueClosure struct {
	AncestorID   uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"ancestor_id"`
	DescendantID uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"descendant_id"`
	Depth        int       `gorm:"not null" json:"depth"`

	// 关联关系
	Ancestor   *Issue `gorm:"foreignKey:AncestorID;constraint:OnDelete:CASCADE" json:"ancestor,omitempty"`
	Descendant *Issue `gorm:"foreignKey:DescendantID;constraint:OnDelete:CASCADE" json:"descendant,omitempty"`
}

// TableName 指定表名
func (IssueClosure) TableName() string {
	return "issue_closure"
}

// IsSelfReference 检查是否为自引用（depth = 0）
func (c *IssueClosure) IsSelfReference() bool {
	return c.Depth == 0
}

// IsDirectChild 检查是否为直接子级（depth = 1）
func (c *IssueClosure) IsDirectChild() bool {
	return c.Depth == 1
}
