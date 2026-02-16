package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Milestone 里程碑模型
type Milestone struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	Description *string    `gorm:"type:text" json:"description,omitempty"`
	TargetDate  *time.Time `gorm:"type:date" json:"target_date,omitempty"`
	Position    float64    `gorm:"not null;default:0" json:"position"`
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// 关联关系
	Project *Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Issues  []Issue  `gorm:"foreignKey:MilestoneID;constraint:OnDelete:SET NULL" json:"issues,omitempty"`
}

// TableName 指定表名
func (Milestone) TableName() string {
	return "milestones"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (m *Milestone) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// IsOverdue 检查里程碑是否已逾期
func (m *Milestone) IsOverdue() bool {
	if m.TargetDate == nil {
		return false
	}
	return m.TargetDate.Before(time.Now())
}
