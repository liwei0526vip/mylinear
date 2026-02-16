package model

import (
	"gorm.io/datatypes"
)

// Workspace 工作区模型
type Workspace struct {
	Model
	Name     string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"`
	LogoURL  *string        `gorm:"type:text" json:"logo_url,omitempty"`
	Settings datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings,omitempty"`

	// 关联关系
	Teams  []Team  `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"teams,omitempty"`
	Users  []User  `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"users,omitempty"`
	Labels []Label `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"labels,omitempty"`
}

// TableName 指定表名
func (Workspace) TableName() string {
	return "workspaces"
}
