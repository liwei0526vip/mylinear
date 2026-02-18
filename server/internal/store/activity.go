// Package store 提供数据访问层
package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// ListActivitiesOptions 活动列表查询选项
type ListActivitiesOptions struct {
	Page     int   // 页码，从 1 开始
	PageSize int   // 每页数量，默认 50
	Types    []model.ActivityType // 按类型过滤
}

// ActivityStore 定义活动数据访问接口
type ActivityStore interface {
	// CreateActivity 创建活动记录
	CreateActivity(ctx context.Context, activity *model.Activity) error
	// GetActivityByID 根据 ID 获取活动
	GetActivityByID(ctx context.Context, id uuid.UUID) (*model.Activity, error)
	// GetActivitiesByIssueID 获取 Issue 的活动列表
	GetActivitiesByIssueID(ctx context.Context, issueID uuid.UUID, opts *ListActivitiesOptions) ([]model.Activity, error)
	// GetActivitiesByIssueIDWithTotal 获取 Issue 的活动列表（带总数）
	GetActivitiesByIssueIDWithTotal(ctx context.Context, issueID uuid.UUID, opts *ListActivitiesOptions) ([]model.Activity, int64, error)
}

// activityStore 实现 ActivityStore 接口
type activityStore struct {
	db *gorm.DB
}

// NewActivityStore 创建活动存储实例
func NewActivityStore(db *gorm.DB) ActivityStore {
	return &activityStore{db: db}
}

// CreateActivity 创建活动记录
func (s *activityStore) CreateActivity(ctx context.Context, activity *model.Activity) error {
	if activity == nil {
		return fmt.Errorf("activity 不能为 nil")
	}

	if err := s.db.WithContext(ctx).Create(activity).Error; err != nil {
		return fmt.Errorf("创建活动记录失败: %w", err)
	}

	return nil
}

// GetActivityByID 根据 ID 获取活动
func (s *activityStore) GetActivityByID(ctx context.Context, id uuid.UUID) (*model.Activity, error) {
	var activity model.Activity

	err := s.db.WithContext(ctx).
		Preload("Actor").
		First(&activity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到活动记录: %w", err)
		}
		return nil, fmt.Errorf("查询活动记录失败: %w", err)
	}

	return &activity, nil
}

// GetActivitiesByIssueID 获取 Issue 的活动列表
func (s *activityStore) GetActivitiesByIssueID(ctx context.Context, issueID uuid.UUID, opts *ListActivitiesOptions) ([]model.Activity, error) {
	query := s.db.WithContext(ctx).
		Preload("Actor").
		Where("issue_id = ?", issueID).
		Order("created_at DESC")

	// 应用类型过滤
	if opts != nil && len(opts.Types) > 0 {
		query = query.Where("type IN ?", opts.Types)
	}

	// 应用分页
	if opts != nil && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		if offset < 0 {
			offset = 0
		}
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	var activities []model.Activity
	if err := query.Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("查询活动列表失败: %w", err)
	}

	return activities, nil
}

// GetActivitiesByIssueIDWithTotal 获取 Issue 的活动列表（带总数）
func (s *activityStore) GetActivitiesByIssueIDWithTotal(ctx context.Context, issueID uuid.UUID, opts *ListActivitiesOptions) ([]model.Activity, int64, error) {
	// 先获取总数
	var total int64
	countQuery := s.db.WithContext(ctx).
		Model(&model.Activity{}).
		Where("issue_id = ?", issueID)

	if opts != nil && len(opts.Types) > 0 {
		countQuery = countQuery.Where("type IN ?", opts.Types)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计活动数量失败: %w", err)
	}

	// 获取列表
	activities, err := s.GetActivitiesByIssueID(ctx, issueID, opts)
	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}
