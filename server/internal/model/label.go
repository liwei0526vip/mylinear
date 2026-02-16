package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Label 标签模型
type Label struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;not null;index" json:"workspace_id"`
	TeamID      *uuid.UUID `gorm:"type:uuid;index" json:"team_id,omitempty"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Description *string    `gorm:"type:text" json:"description,omitempty"`
	Color       string     `gorm:"type:varchar(20);default:'#808080'" json:"color"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	IsArchived  bool       `gorm:"not null;default:false;index" json:"is_archived"`
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// 关联关系
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"workspace,omitempty"`
	Team      *Team      `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"team,omitempty"`
	Parent    *Label     `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL" json:"parent,omitempty"`
	Children  []Label    `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL" json:"children,omitempty"`
}

// TableName 指定表名
func (Label) TableName() string {
	return "labels"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (l *Label) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// IsGlobal 检查是否为全局标签（不属于特定团队）
func (l *Label) IsGlobal() bool {
	return l.TeamID == nil
}

// IsActive 检查标签是否活跃（未归档）
func (l *Label) IsActive() bool {
	return !l.IsArchived
}
