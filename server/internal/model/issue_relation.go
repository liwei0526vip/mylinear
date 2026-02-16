package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IssueRelation Issue 关系模型
type IssueRelation struct {
	ID              uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IssueID         uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_issue_relation_unique;index" json:"issue_id"`
	RelatedIssueID  uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_issue_relation_unique;index" json:"related_issue_id"`
	Type            IssueRelationType `gorm:"type:varchar(20);not null;default:'related';uniqueIndex:idx_issue_relation_unique" json:"type"`
	CreatedAt       time.Time        `gorm:"not null;default:now()" json:"created_at"`

	// 关联关系
	Issue        *Issue `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"issue,omitempty"`
	RelatedIssue *Issue `gorm:"foreignKey:RelatedIssueID;constraint:OnDelete:CASCADE" json:"related_issue,omitempty"`
}

// TableName 指定表名
func (IssueRelation) TableName() string {
	return "issue_relations"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (r *IssueRelation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
