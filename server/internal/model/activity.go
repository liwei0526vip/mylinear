package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Activity 活动记录模型
type Activity struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IssueID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"issue_id"`
	Type      ActivityType   `gorm:"type:varchar(50);not null;index" json:"type"`
	ActorID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"actor_id"`
	Payload   datatypes.JSON `gorm:"type:jsonb" json:"payload"`
	CreatedAt time.Time      `gorm:"not null;default:now();index" json:"created_at"`

	// 关联关系
	Issue *Issue `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"issue,omitempty"`
	Actor *User  `gorm:"foreignKey:ActorID;constraint:OnDelete:RESTRICT" json:"actor,omitempty"`
}

// TableName 指定表名
func (Activity) TableName() string {
	return "activities"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// ActivityPayloadTitle 标题变更 Payload
type ActivityPayloadTitle struct {
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// ActivityPayloadDescription 描述变更 Payload
type ActivityPayloadDescription struct {
	OldValue string `json:"old_value,omitempty"`
	NewValue string `json:"new_value,omitempty"`
}

// ActivityPayloadStatus 状态变更 Payload
type ActivityPayloadStatus struct {
	OldStatus *ActivityStatusRef `json:"old_status,omitempty"`
	NewStatus *ActivityStatusRef `json:"new_status"`
}

// ActivityStatusRef 状态引用（用于 Payload）
type ActivityStatusRef struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

// ActivityPayloadUser 用户引用（用于 Payload）
type ActivityPayloadUser struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
}

// ActivityPayloadAssignee 负责人变更 Payload
type ActivityPayloadAssignee struct {
	OldAssignee *ActivityPayloadUser `json:"old_assignee,omitempty"`
	NewAssignee *ActivityPayloadUser `json:"new_assignee,omitempty"`
}

// ActivityPayloadPriority 优先级变更 Payload
type ActivityPayloadPriority struct {
	OldValue int `json:"old_value"`
	NewValue int `json:"new_value"`
}

// ActivityPayloadDueDate 截止日期变更 Payload
type ActivityPayloadDueDate struct {
	OldValue *time.Time `json:"old_value,omitempty"`
	NewValue *time.Time `json:"new_value,omitempty"`
}

// ActivityPayloadProject 项目变更 Payload
type ActivityPayloadProject struct {
	OldProject *ActivityProjectRef `json:"old_project,omitempty"`
	NewProject *ActivityProjectRef `json:"new_project,omitempty"`
}

// ActivityProjectRef 项目引用（用于 Payload）
type ActivityProjectRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ActivityPayloadLabels 标签变更 Payload
type ActivityPayloadLabels struct {
	Added   []ActivityLabelRef `json:"added"`
	Removed []ActivityLabelRef `json:"removed"`
}

// ActivityLabelRef 标签引用（用于 Payload）
type ActivityLabelRef struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

// ActivityPayloadComment 评论添加 Payload
type ActivityPayloadComment struct {
	CommentID      uuid.UUID `json:"comment_id"`
	CommentPreview string    `json:"comment_preview"`
}
