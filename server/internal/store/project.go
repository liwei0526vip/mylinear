// Package store 提供数据访问层
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// 常量定义
const (
	MaxDescriptionLength = 10000
	MaxNameLength        = 255
)

// 错误定义
var (
	ErrProjectNameEmpty       = errors.New("名称不能为空")
	ErrProjectNameTooLong     = errors.New("名称长度不能超过255字符")
	ErrProjectDescTooLong     = errors.New("描述长度不能超过10000字符")
	ErrProjectNotFound        = errors.New("项目不存在")
)

// ProjectFilter 项目列表过滤条件
type ProjectFilter struct {
	Status *model.ProjectStatus
}

// ProjectProgress 项目进度统计
type ProjectProgress struct {
	TotalIssues      int     `json:"total_issues"`
	CompletedIssues  int     `json:"completed_issues"`
	CancelledIssues  int     `json:"cancelled_issues"`
	ProgressPercent  float64 `json:"progress_percent"`
}

// ProjectStore 定义 Project 数据访问接口
type ProjectStore interface {
	// Create 创建项目
	Create(ctx context.Context, project *model.Project) error
	// GetByID 通过 ID 获取项目（预加载关联）
	GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	// Update 更新项目
	Update(ctx context.Context, project *model.Project) error
	// SoftDelete 软删除项目
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// Restore 恢复已删除的项目
	Restore(ctx context.Context, id uuid.UUID) error
	// ListByWorkspace 获取工作区下的项目列表
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, filter *ProjectFilter, page, pageSize int) ([]model.Project, int64, error)
	// ListByTeam 获取与团队关联的项目列表
	ListByTeam(ctx context.Context, teamID uuid.UUID, filter *ProjectFilter, page, pageSize int) ([]model.Project, int64, error)
	// GetProgress 获取项目进度统计
	GetProgress(ctx context.Context, projectID uuid.UUID) (*ProjectProgress, error)
	// ListIssues 获取项目关联的 Issue 列表
	ListIssues(ctx context.Context, projectID uuid.UUID, filter *IssueFilter, page, pageSize int) ([]model.Issue, int64, error)
}

// projectStore 实现 ProjectStore 接口
type projectStore struct {
	db *gorm.DB
}

// NewProjectStore 创建 Project 存储实例
func NewProjectStore(db *gorm.DB) ProjectStore {
	return &projectStore{db: db}
}

// Create 创建项目
func (s *projectStore) Create(ctx context.Context, project *model.Project) error {
	// 验证名称
	if project.Name == "" {
		return ErrProjectNameEmpty
	}
	if len(project.Name) > MaxNameLength {
		return ErrProjectNameTooLong
	}

	// 验证描述长度
	if project.Description != nil && len(*project.Description) > MaxDescriptionLength {
		return ErrProjectDescTooLong
	}

	// 设置默认状态
	if project.Status == "" {
		project.Status = model.ProjectStatusPlanned
	}

	// 初始化空数组
	if project.Teams == nil {
		project.Teams = pq.StringArray{}
	}
	if project.Labels == nil {
		project.Labels = pq.StringArray{}
	}

	if err := s.db.WithContext(ctx).Create(project).Error; err != nil {
		return fmt.Errorf("创建项目失败: %w", err)
	}

	return nil
}

// GetByID 通过 ID 获取项目（预加载关联）
func (s *projectStore) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var project model.Project
	err := s.db.WithContext(ctx).
		Preload("Workspace").
		Preload("Lead").
		Where("id = ?", id).
		First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("获取项目失败: %w", err)
	}
	return &project, nil
}

// Update 更新项目
func (s *projectStore) Update(ctx context.Context, project *model.Project) error {
	if err := s.db.WithContext(ctx).Save(project).Error; err != nil {
		return fmt.Errorf("更新项目失败: %w", err)
	}
	return nil
}

// SoftDelete 软删除项目
func (s *projectStore) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&model.Project{}, "id = ?", id).Error
}

// Restore 恢复已删除的项目
func (s *projectStore) Restore(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Unscoped().Model(&model.Project{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// ListByWorkspace 获取工作区下的项目列表
func (s *projectStore) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, filter *ProjectFilter, page, pageSize int) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64

	query := s.db.WithContext(ctx).Model(&model.Project{}).Where("workspace_id = ?", workspaceID)

	// 应用过滤条件
	if filter != nil && filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计项目数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Preload("Lead").
		Offset(offset).
		Limit(pageSize).
		Order("updated_at DESC").
		Find(&projects).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询项目列表失败: %w", err)
	}

	return projects, total, nil
}

// ListByTeam 获取与团队关联的项目列表
func (s *projectStore) ListByTeam(ctx context.Context, teamID uuid.UUID, filter *ProjectFilter, page, pageSize int) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64

	// 将 teamID 转换为字符串（用于数组包含查询）
	teamIDStr := teamID.String()

	query := s.db.WithContext(ctx).Model(&model.Project{}).
		Where("teams::text[] @> ARRAY[?]::text[]", teamIDStr)

	// 应用过滤条件
	if filter != nil && filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计项目数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Preload("Lead").
		Offset(offset).
		Limit(pageSize).
		Order("updated_at DESC").
		Find(&projects).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询项目列表失败: %w", err)
	}

	return projects, total, nil
}

// GetProgress 获取项目进度统计
func (s *projectStore) GetProgress(ctx context.Context, projectID uuid.UUID) (*ProjectProgress, error) {
	progress := &ProjectProgress{}

	// 统计各类型 Issue 数量
	type IssueCount struct {
		StateType string
		Count     int
	}

	var counts []IssueCount
	err := s.db.WithContext(ctx).
		Model(&model.Issue{}).
		Select("ws.type as state_type, COUNT(*) as count").
		Joins("JOIN workflow_states ws ON issues.status_id = ws.id").
		Where("issues.project_id = ?", projectID).
		Group("ws.type").
		Scan(&counts).Error
	if err != nil {
		return nil, fmt.Errorf("统计项目进度失败: %w", err)
	}

	// 计算统计数据
	for _, c := range counts {
		progress.TotalIssues += c.Count
		if c.StateType == string(model.StateTypeCompleted) {
			progress.CompletedIssues = c.Count
		}
		if c.StateType == string(model.StateTypeCanceled) {
			progress.CancelledIssues = c.Count
		}
	}

	// 计算进度百分比
	effectiveTotal := progress.TotalIssues - progress.CancelledIssues
	if effectiveTotal > 0 {
		progress.ProgressPercent = float64(progress.CompletedIssues) / float64(effectiveTotal) * 100
	}

	return progress, nil
}

// ListIssues 获取项目关联的 Issue 列表
func (s *projectStore) ListIssues(ctx context.Context, projectID uuid.UUID, filter *IssueFilter, page, pageSize int) ([]model.Issue, int64, error) {
	var issues []model.Issue
	var total int64

	query := s.db.WithContext(ctx).Model(&model.Issue{}).Where("project_id = ?", projectID)

	// 应用过滤条件
	if filter != nil {
		if filter.StatusID != nil {
			query = query.Where("status_id = ?", filter.StatusID)
		}
		if filter.Priority != nil {
			query = query.Where("priority = ?", filter.Priority)
		}
		if filter.AssigneeID != nil {
			query = query.Where("assignee_id = ?", filter.AssigneeID)
		}
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计 Issue 数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Preload("Team").
		Preload("Status").
		Preload("Assignee").
		Preload("CreatedBy").
		Offset(offset).
		Limit(pageSize).
		Order("position ASC").
		Find(&issues).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询 Issue 列表失败: %w", err)
	}

	return issues, total, nil
}
