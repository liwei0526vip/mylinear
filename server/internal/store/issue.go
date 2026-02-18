// Package store 提供数据访问层
package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// IssueFilter Issue 列表过滤条件
type IssueFilter struct {
	StatusID    *uuid.UUID
	Priority    *int
	AssigneeID  *uuid.UUID
	ProjectID   *uuid.UUID
	CycleID     *uuid.UUID
	LabelIDs    []uuid.UUID
	CreatedByID *uuid.UUID
}

// IssueStore 定义 Issue 数据访问接口
type IssueStore interface {
	// Create 创建 Issue（在事务中自动生成 Number）
	Create(ctx context.Context, issue *model.Issue) error
	// GetByID 通过 ID 获取 Issue（预加载关联）
	GetByID(ctx context.Context, id uuid.UUID) (*model.Issue, error)
	// List 获取 Issue 列表（支持过滤和分页）
	List(ctx context.Context, teamID uuid.UUID, filter *IssueFilter, page, pageSize int) ([]model.Issue, int64, error)
	// Update 更新 Issue
	Update(ctx context.Context, issue *model.Issue) error
	// SoftDelete 软删除 Issue
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// Restore 恢复已删除的 Issue
	Restore(ctx context.Context, id uuid.UUID) error
	// UpdatePosition 更新 Issue 排序位置
	UpdatePosition(ctx context.Context, id uuid.UUID, position float64, statusID *uuid.UUID) error
	// GetMaxNumber 获取团队内最大 Issue Number
	GetMaxNumber(ctx context.Context, teamID uuid.UUID) (int, error)
	// ListBySubscription 获取用户订阅的 Issue 列表
	ListBySubscription(ctx context.Context, userID uuid.UUID) ([]model.Issue, error)
}

// issueStore 实现 IssueStore 接口
type issueStore struct {
	db *gorm.DB
}

// NewIssueStore 创建 Issue 存储实例
func NewIssueStore(db *gorm.DB) IssueStore {
	return &issueStore{db: db}
}

// Create 创建 Issue（在事务中自动生成 Number）
func (s *issueStore) Create(ctx context.Context, issue *model.Issue) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取团队内最大 Number
		var maxNumber int
		err := tx.Model(&model.Issue{}).
			Where("team_id = ?", issue.TeamID).
			Select("COALESCE(MAX(number), 0)").
			Scan(&maxNumber).Error
		if err != nil {
			return fmt.Errorf("获取最大 Number 失败: %w", err)
		}

		// 设置新 Number
		issue.Number = maxNumber + 1

		// 如果 Position 为 0，设置默认值
		if issue.Position == 0 {
			issue.Position = float64(issue.Number * 1000)
		}

		// 创建 Issue
		if err := tx.Create(issue).Error; err != nil {
			return fmt.Errorf("创建 Issue 失败: %w", err)
		}

		return nil
	})
}

// GetByID 通过 ID 获取 Issue（预加载关联）
func (s *issueStore) GetByID(ctx context.Context, id uuid.UUID) (*model.Issue, error) {
	var issue model.Issue
	err := s.db.WithContext(ctx).
		Preload("Team").
		Preload("Status").
		Preload("Assignee").
		Preload("Project").
		Preload("CreatedBy").
		Where("id = ?", id).
		First(&issue).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// List 获取 Issue 列表（支持过滤和分页）
func (s *issueStore) List(ctx context.Context, teamID uuid.UUID, filter *IssueFilter, page, pageSize int) ([]model.Issue, int64, error) {
	var issues []model.Issue
	var total int64

	// 构建查询
	query := s.db.WithContext(ctx).Model(&model.Issue{}).Where("team_id = ?", teamID)

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
		if filter.ProjectID != nil {
			query = query.Where("project_id = ?", filter.ProjectID)
		}
		if filter.CycleID != nil {
			query = query.Where("cycle_id = ?", filter.CycleID)
		}
		if filter.CreatedByID != nil {
			query = query.Where("created_by_id = ?", filter.CreatedByID)
		}
		if len(filter.LabelIDs) > 0 {
			// 使用数组重叠查询
			query = query.Where("labels && ?", filter.LabelIDs)
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

// Update 更新 Issue
func (s *issueStore) Update(ctx context.Context, issue *model.Issue) error {
	// 使用 Session 配置跳过钩子更新指定字段，避免关联对象干扰
	return s.db.WithContext(ctx).Model(issue).Select(
		"Title",
		"Description",
		"StatusID",
		"Priority",
		"AssigneeID",
		"ProjectID",
		"MilestoneID",
		"CycleID",
		"ParentID",
		"Estimate",
		"DueDate",
		"SLADueAt",
		"Labels",
		"Position",
		"CompletedAt",
		"CancelledAt",
	).Updates(issue).Error
}

// SoftDelete 软删除 Issue
func (s *issueStore) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&model.Issue{}, "id = ?", id).Error
}

// Restore 恢复已删除的 Issue
func (s *issueStore) Restore(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Unscoped().Model(&model.Issue{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// UpdatePosition 更新 Issue 排序位置
func (s *issueStore) UpdatePosition(ctx context.Context, id uuid.UUID, position float64, statusID *uuid.UUID) error {
	updates := map[string]interface{}{
		"position": position,
	}
	if statusID != nil {
		updates["status_id"] = statusID
	}

	return s.db.WithContext(ctx).Model(&model.Issue{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetMaxNumber 获取团队内最大 Issue Number
func (s *issueStore) GetMaxNumber(ctx context.Context, teamID uuid.UUID) (int, error) {
	var maxNumber int
	err := s.db.WithContext(ctx).Model(&model.Issue{}).
		Where("team_id = ?", teamID).
		Select("COALESCE(MAX(number), 0)").
		Scan(&maxNumber).Error
	if err != nil {
		return 0, fmt.Errorf("获取最大 Number 失败: %w", err)
	}
	return maxNumber, nil
}

// ListBySubscription 获取用户订阅的 Issue 列表
func (s *issueStore) ListBySubscription(ctx context.Context, userID uuid.UUID) ([]model.Issue, error) {
	var issues []model.Issue

	err := s.db.WithContext(ctx).
		Joins("JOIN issue_subscriptions ON issues.id = issue_subscriptions.issue_id").
		Where("issue_subscriptions.user_id = ?", userID).
		Preload("Team").
		Preload("Status").
		Preload("Assignee").
		Preload("CreatedBy").
		Order("issues.created_at DESC").
		Find(&issues).Error
	if err != nil {
		return nil, fmt.Errorf("查询订阅的 Issue 列表失败: %w", err)
	}

	return issues, nil
}
