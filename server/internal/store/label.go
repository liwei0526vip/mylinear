package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// LabelStore 定义标签数据访问接口
type LabelStore interface {
	// Create 创建标签
	Create(ctx context.Context, label *model.Label) error

	// GetByID 根据ID获取标签
	GetByID(ctx context.Context, id uuid.UUID) (*model.Label, error)

	// Update 更新标签
	Update(ctx context.Context, label *model.Label) error

	// Delete 删除标签
	Delete(ctx context.Context, id uuid.UUID) error

	// ListForWorkspace 获取工作区下的全局标签（不属于特定团队）
	ListForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*model.Label, error)

	// ListForTeam 获取团队可见的所有标签（工作区全局标签 + 团队私有标签）
	ListForTeam(ctx context.Context, workspaceID, teamID uuid.UUID) ([]*model.Label, error)
}

type labelStore struct {
	db *gorm.DB
}

// NewLabelStore 创建 LabelStore 实例
func NewLabelStore(db *gorm.DB) LabelStore {
	return &labelStore{db: db}
}

func (s *labelStore) Create(ctx context.Context, label *model.Label) error {
	return s.db.WithContext(ctx).Create(label).Error
}

func (s *labelStore) GetByID(ctx context.Context, id uuid.UUID) (*model.Label, error) {
	var label model.Label
	if err := s.db.WithContext(ctx).First(&label, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &label, nil
}

func (s *labelStore) Update(ctx context.Context, label *model.Label) error {
	return s.db.WithContext(ctx).Save(label).Error
}

func (s *labelStore) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&model.Label{}, "id = ?", id).Error
}

func (s *labelStore) ListForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*model.Label, error) {
	var labels []*model.Label
	// 获取 workspace_id = ? AND team_id IS NULL
	if err := s.db.WithContext(ctx).
		Where("workspace_id = ? AND team_id IS NULL", workspaceID).
		Order("name ASC").
		Find(&labels).Error; err != nil {
		return nil, err
	}
	return labels, nil
}

func (s *labelStore) ListForTeam(ctx context.Context, workspaceID, teamID uuid.UUID) ([]*model.Label, error) {
	var labels []*model.Label
	// 获取 (workspace_id = ? AND team_id IS NULL) OR (workspace_id = ? AND team_id = ?)
	// 简化为 workspace_id = ? AND (team_id IS NULL OR team_id = ?)
	if err := s.db.WithContext(ctx).
		Where("workspace_id = ? AND (team_id IS NULL OR team_id = ?)", workspaceID, teamID).
		Order("name ASC"). // 按名称排序，或者按 (team_id IS NOT NULL, name) 排序将团队特定标签分开显示？Spec没说，按名字混排即可。
		Find(&labels).Error; err != nil {
		return nil, err
	}
	return labels, nil
}
