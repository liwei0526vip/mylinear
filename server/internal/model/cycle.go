package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cycle 迭代模型
type Cycle struct {
	ID             uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TeamID         uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_cycle_team_number;index" json:"team_id"`
	Number         int         `gorm:"not null;uniqueIndex:idx_cycle_team_number" json:"number"`
	Name           string      `gorm:"type:varchar(255)" json:"name,omitempty"`
	Description    *string     `gorm:"type:text" json:"description,omitempty"`
	StartDate      time.Time   `gorm:"type:date;not null" json:"start_date"`
	EndDate        time.Time   `gorm:"type:date;not null" json:"end_date"`
	CooldownEndDate *time.Time `gorm:"type:date" json:"cooldown_end_date,omitempty"`
	Status         CycleStatus `gorm:"type:varchar(20);not null;default:'upcoming';index" json:"status"`
	CreatedAt      time.Time   `gorm:"not null;default:now()" json:"created_at"`

	// 关联关系
	Team   *Team   `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"team,omitempty"`
	Issues []Issue `gorm:"foreignKey:CycleID;constraint:OnDelete:SET NULL" json:"issues,omitempty"`
}

// TableName 指定表名
func (Cycle) TableName() string {
	return "cycles"
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (c *Cycle) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsActive 检查迭代是否活跃
func (c *Cycle) IsActive() bool {
	return c.Status == CycleStatusActive
}

// IsCompleted 检查迭代是否已完成
func (c *Cycle) IsCompleted() bool {
	return c.Status == CycleStatusCompleted
}

// IsUpcoming 检查迭代是否即将开始
func (c *Cycle) IsUpcoming() bool {
	return c.Status == CycleStatusUpcoming
}

// Duration 返回迭代持续天数
func (c *Cycle) Duration() int {
	return int(c.EndDate.Sub(c.StartDate).Hours() / 24)
}

// DisplayName 返回迭代显示名称
func (c *Cycle) DisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	return "Cycle " + string(rune(c.Number))
}
