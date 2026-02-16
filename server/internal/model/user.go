package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// User 用户模型
type User struct {
	Model
	WorkspaceID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Username     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	AvatarURL    *string        `gorm:"type:text" json:"avatar_url,omitempty"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Role         Role           `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	Settings     datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings,omitempty"`

	// 关联关系
	Workspace    *Workspace   `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"workspace,omitempty"`
	TeamMembers  []TeamMember `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"team_members,omitempty"`
	CreatedIssues []Issue     `gorm:"foreignKey:CreatedByID;constraint:OnDelete:RESTRICT" json:"created_issues,omitempty"`
	AssignedIssues []Issue    `gorm:"foreignKey:AssigneeID;constraint:OnDelete:SET NULL" json:"assigned_issues,omitempty"`
	Comments     []Comment    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	Notifications []Notification `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"notifications,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsGlobalAdmin 检查是否为全局管理员
func (u *User) IsGlobalAdmin() bool {
	return u.Role == RoleGlobalAdmin
}

// IsAdmin 检查是否为管理员（全局或团队）
func (u *User) IsAdmin() bool {
	return u.Role == RoleGlobalAdmin || u.Role == RoleAdmin
}
