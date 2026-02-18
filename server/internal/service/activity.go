// Package service 提供业务逻辑层
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

// ActivityService 定义活动业务逻辑接口
type ActivityService interface {
	// RecordActivity 记录活动
	RecordActivity(ctx context.Context, activity *model.Activity) error
	// GetIssueActivities 获取 Issue 的活动列表
	GetIssueActivities(ctx context.Context, issueID uuid.UUID, page, pageSize int, types []model.ActivityType) ([]model.Activity, int64, error)
}

// activityService 实现 ActivityService 接口
type activityService struct {
	activityStore store.ActivityStore
}

// NewActivityService 创建活动服务实例
func NewActivityService(activityStore store.ActivityStore) ActivityService {
	return &activityService{
		activityStore: activityStore,
	}
}

// RecordActivity 记录活动
func (s *activityService) RecordActivity(ctx context.Context, activity *model.Activity) error {
	return s.activityStore.CreateActivity(ctx, activity)
}

// GetIssueActivities 获取 Issue 的活动列表
func (s *activityService) GetIssueActivities(ctx context.Context, issueID uuid.UUID, page, pageSize int, types []model.ActivityType) ([]model.Activity, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	opts := &store.ListActivitiesOptions{
		Page:     page,
		PageSize: pageSize,
		Types:    types,
	}

	return s.activityStore.GetActivitiesByIssueIDWithTotal(ctx, issueID, opts)
}
