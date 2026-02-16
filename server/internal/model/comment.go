package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment 评论模型
type Comment struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IssueID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"issue_id"`
	ParentID  *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Body      string     `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time  `gorm:"not null;default:now();index" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	EditedAt  *time.Time `gorm:"type:timestamptz" json:"edited_at,omitempty"`

	// 关联关系
	Issue   *Issue    `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"issue,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"replies,omitempty"`
	User    *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsReply 检查是否为回复评论
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}

// IsEdited 检查评论是否被编辑过
func (c *Comment) IsEdited() bool {
	return c.EditedAt != nil
}
