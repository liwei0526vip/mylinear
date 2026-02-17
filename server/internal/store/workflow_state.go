package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// WorkflowStateStore 定义了工作流状态的存储接口
type WorkflowStateStore interface {
	// Create 创建一个新的工作流状态
	Create(ctx context.Context, state *model.WorkflowState) error

	// GetFullByID 根据 ID 获取工作流状态
	GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowState, error)

	// ListByTeamID 获取指定团队的所有工作流状态，按 Position 升序排序
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]*model.WorkflowState, error)

	// Update 更新工作流状态
	Update(ctx context.Context, state *model.WorkflowState) error

	// Delete 删除状态
	Delete(ctx context.Context, id uuid.UUID) error
	// CountIssues 统计状态下的 Issue 数量
	CountIssues(ctx context.Context, stateID uuid.UUID) (int64, error)
	// GetMaxPosition 获取最大位置值
	GetMaxPosition(ctx context.Context, teamID uuid.UUID) (float64, error)

	// CountByTeamID 统计团队下的状态数量
	CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error)
}

type workflowStateStore struct {
	db *gorm.DB
}

// NewWorkflowStateStore 创建一个新的 WorkflowStateStore 实例
func NewWorkflowStateStore(db *gorm.DB) WorkflowStateStore {
	return &workflowStateStore{db: db}
}

func (s *workflowStateStore) Create(ctx context.Context, state *model.WorkflowState) error {
	return s.db.WithContext(ctx).Create(state).Error
}

func (s *workflowStateStore) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowState, error) {
	var state model.WorkflowState
	if err := s.db.WithContext(ctx).First(&state, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *workflowStateStore) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]*model.WorkflowState, error) {
	var states []*model.WorkflowState
	// 按 Position 升序排序
	if err := s.db.WithContext(ctx).Where("team_id = ?", teamID).Order("position ASC").Find(&states).Error; err != nil {
		return nil, err
	}
	return states, nil
}

func (s *workflowStateStore) Update(ctx context.Context, state *model.WorkflowState) error {
	// 使用 Updates 更新非零值，如果需要更新零值（如 IsDefault=false），需要 careful
	// 这里假设 state 是完整的对象，或者使用 Omit/Select
	// 通常 Update 接收的是 entity，直接 Save 或者 Updates
	return s.db.WithContext(ctx).Save(state).Error
}

func (s *workflowStateStore) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&model.WorkflowState{}, "id = ?", id).Error
}

// CountIssues 统计状态下的 Issue 数量
func (s *workflowStateStore) CountIssues(ctx context.Context, stateID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Issue{}).Where("status_id = ?", stateID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计 Issue 数量失败: %w", err)
	}
	return count, nil
}

func (s *workflowStateStore) GetMaxPosition(ctx context.Context, teamID uuid.UUID) (float64, error) {
	var result struct {
		MaxPos float64
	}
	// 处理该团队没有状态的情况
	err := s.db.WithContext(ctx).Model(&model.WorkflowState{}).
		Where("team_id = ?", teamID).
		Select("COALESCE(MAX(position), 0) as max_pos").
		Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.MaxPos, nil
}

func (s *workflowStateStore) CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.WorkflowState{}).
		Where("team_id = ?", teamID).
		Count(&count).Error
	return count, err
}
