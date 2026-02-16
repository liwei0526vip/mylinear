package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Attachment 附件模型
type Attachment struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IssueID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"issue_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Filename  string     `gorm:"type:varchar(255);not null" json:"filename"`
	URL       string     `gorm:"type:text;not null" json:"url"`
	Size      int64      `gorm:"type:bigint;not null;default:0" json:"size"`
	MimeType  string     `gorm:"type:varchar(100)" json:"mime_type,omitempty"`
	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// 关联关系
	Issue *Issue `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"issue,omitempty"`
	User  *User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (Attachment) TableName() string {
	return "attachments"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (a *Attachment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// IsImage 检查附件是否为图片
func (a *Attachment) IsImage() bool {
	if a.MimeType == "" {
		return false
	}
	return a.MimeType == "image/jpeg" ||
		a.MimeType == "image/png" ||
		a.MimeType == "image/gif" ||
		a.MimeType == "image/webp" ||
		a.MimeType == "image/svg+xml"
}

// IsVideo 检查附件是否为视频
func (a *Attachment) IsVideo() bool {
	if a.MimeType == "" {
		return false
	}
	return a.MimeType == "video/mp4" ||
		a.MimeType == "video/webm" ||
		a.MimeType == "video/quicktime"
}

// HumanSize 返回人类可读的文件大小
func (a *Attachment) HumanSize() string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case a.Size >= GB:
		return string(rune(a.Size/GB)) + " GB"
	case a.Size >= MB:
		return string(rune(a.Size/MB)) + " MB"
	case a.Size >= KB:
		return string(rune(a.Size/KB)) + " KB"
	default:
		return string(rune(a.Size)) + " B"
	}
}
