// Package model 定义数据库模型
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model 基础模型结构体，包含所有模型的通用字段
type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// BeforeCreate GORM 钩子，在创建记录前自动生成 UUID
func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ModelWithSoftDelete 带软删除的基础模型
type ModelWithSoftDelete struct {
	Model
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TimestampOnly 仅包含时间戳的模型（用于不需要 ID 的关联表）
type TimestampOnly struct {
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}
