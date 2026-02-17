// Package store 提供数据访问层
package store

import (
	"context"
	"fmt"

	"github.com/mylinear/server/internal/model"
	"gorm.io/gorm"
)

// WorkspaceStore 定义工作区数据访问接口
type WorkspaceStore interface {
	// GetByID 通过 ID 获取工作区
	GetByID(ctx context.Context, id string) (*model.Workspace, error)
	// Update 更新工作区信息
	Update(ctx context.Context, workspace *model.Workspace) error
	// GetStats 获取工作区统计信息
	GetStats(ctx context.Context, id string) (*WorkspaceStats, error)
}

// WorkspaceStats 工作区统计信息
type WorkspaceStats struct {
	TeamsCount   int64 `json:"teams_count"`
	MembersCount int64 `json:"members_count"`
	IssuesCount  int64 `json:"issues_count"`
}

// workspaceStore 实现 WorkspaceStore 接口
type workspaceStore struct {
	db *gorm.DB
}

// NewWorkspaceStore 创建工作区存储实例
func NewWorkspaceStore(db *gorm.DB) WorkspaceStore {
	return &workspaceStore{db: db}
}

// GetByID 通过 ID 获取工作区
func (s *workspaceStore) GetByID(ctx context.Context, id string) (*model.Workspace, error) {
	var workspace model.Workspace
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// Update 更新工作区信息
func (s *workspaceStore) Update(ctx context.Context, workspace *model.Workspace) error {
	return s.db.WithContext(ctx).Save(workspace).Error
}

// GetStats 获取工作区统计信息
func (s *workspaceStore) GetStats(ctx context.Context, id string) (*WorkspaceStats, error) {
	stats := &WorkspaceStats{}

	// 统计团队数量
	if err := s.db.WithContext(ctx).
		Model(&model.Team{}).
		Where("workspace_id = ?", id).
		Count(&stats.TeamsCount).Error; err != nil {
		return nil, fmt.Errorf("统计团队数量失败: %w", err)
	}

	// 统计成员数量
	if err := s.db.WithContext(ctx).
		Model(&model.User{}).
		Where("workspace_id = ?", id).
		Count(&stats.MembersCount).Error; err != nil {
		return nil, fmt.Errorf("统计成员数量失败: %w", err)
	}

	// 统计 Issue 数量（通过团队关联）
	if err := s.db.WithContext(ctx).
		Table("issues").
		Joins("JOIN teams ON teams.id = issues.team_id").
		Where("teams.workspace_id = ?", id).
		Count(&stats.IssuesCount).Error; err != nil {
		return nil, fmt.Errorf("统计 Issue 数量失败: %w", err)
	}

	return stats, nil
}
