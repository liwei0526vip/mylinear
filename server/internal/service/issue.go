// Package service 提供业务逻辑层
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

// IssueFilter Issue 过滤条件
type IssueFilter struct {
	StatusID    *string
	Priority    *int
	AssigneeID  *string
	ProjectID   *string
	CycleID     *string
	LabelIDs    []string
	CreatedByID *string
}

// CreateIssueParams 创建 Issue 参数
type CreateIssueParams struct {
	TeamID      uuid.UUID
	Title       string
	Description *string
	StatusID    uuid.UUID
	Priority    int
	AssigneeID  *uuid.UUID
	ProjectID   *uuid.UUID
	Labels      []uuid.UUID
	DueDate     *string
}

// IssueService 定义 Issue 服务接口
type IssueService interface {
	// CreateIssue 创建 Issue
	CreateIssue(ctx context.Context, params *CreateIssueParams) (*model.Issue, error)
	// GetIssue 获取 Issue
	GetIssue(ctx context.Context, issueID string) (*model.Issue, error)
	// ListIssues 获取 Issue 列表
	ListIssues(ctx context.Context, teamID string, filter *IssueFilter, page, pageSize int) ([]model.Issue, int64, error)
	// UpdateIssue 更新 Issue
	UpdateIssue(ctx context.Context, issueID string, updates map[string]interface{}) (*model.Issue, error)
	// DeleteIssue 删除 Issue
	DeleteIssue(ctx context.Context, issueID string) error
	// RestoreIssue 恢复已删除的 Issue
	RestoreIssue(ctx context.Context, issueID string) error
	// Subscribe 订阅 Issue
	Subscribe(ctx context.Context, issueID string) error
	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, issueID string) error
	// ListSubscribers 获取订阅者列表
	ListSubscribers(ctx context.Context, issueID string) ([]model.User, error)
	// UpdatePosition 更新 Issue 位置
	UpdatePosition(ctx context.Context, issueID string, position float64, statusID *string) error
}

// issueService 实现 IssueService 接口
type issueService struct {
	issueStore        store.IssueStore
	subscriptionStore store.IssueSubscriptionStore
	teamMemberStore   store.TeamMemberStore
}

// NewIssueService 创建 Issue 服务实例
func NewIssueService(issueStore store.IssueStore, subscriptionStore store.IssueSubscriptionStore, teamMemberStore store.TeamMemberStore) IssueService {
	return &issueService{
		issueStore:        issueStore,
		subscriptionStore: subscriptionStore,
		teamMemberStore:   teamMemberStore,
	}
}

// CreateIssue 创建 Issue
func (s *issueService) CreateIssue(ctx context.Context, params *CreateIssueParams) (*model.Issue, error) {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("未认证")
	}

	// 验证标题
	if params.Title == "" {
		return nil, fmt.Errorf("标题不能为空")
	}

	// 如果没有指定状态，使用默认状态（这里简化处理，实际应该查询默认状态）
	var statusID uuid.UUID
	if params.StatusID != uuid.Nil {
		statusID = params.StatusID
	} else {
		// TODO: 查询团队的默认状态
		return nil, fmt.Errorf("必须指定状态")
	}

	// 创建 Issue
	issue := &model.Issue{
		TeamID:      params.TeamID,
		Title:       params.Title,
		Description: params.Description,
		StatusID:    statusID,
		Priority:    params.Priority,
		AssigneeID:  params.AssigneeID,
		ProjectID:   params.ProjectID,
		CreatedByID: userID,
	}

	if err := s.issueStore.Create(ctx, issue); err != nil {
		return nil, fmt.Errorf("创建 Issue 失败: %w", err)
	}

	// 创建者自动订阅
	if err := s.subscriptionStore.Subscribe(ctx, issue.ID, userID); err != nil {
		// 订阅失败不影响创建，记录日志即可
	}

	return issue, nil
}

// GetIssue 获取 Issue
func (s *issueService) GetIssue(ctx context.Context, issueID string) (*model.Issue, error) {
	// 解析 Issue ID
	id, err := uuid.Parse(issueID)
	if err != nil {
		return nil, fmt.Errorf("无效的 Issue ID")
	}

	issue, err := s.issueStore.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Issue 不存在")
	}

	return issue, nil
}

// ListIssues 获取 Issue 列表
func (s *issueService) ListIssues(ctx context.Context, teamID string, filter *IssueFilter, page, pageSize int) ([]model.Issue, int64, error) {
	// 解析 Team ID
	teamUUID, err := uuid.Parse(teamID)
	if err != nil {
		return nil, 0, fmt.Errorf("无效的团队 ID")
	}

	// 转换过滤器
	storeFilter := &store.IssueFilter{}
	if filter != nil {
		if filter.StatusID != nil {
			statusUUID, _ := uuid.Parse(*filter.StatusID)
			storeFilter.StatusID = &statusUUID
		}
		if filter.Priority != nil {
			storeFilter.Priority = filter.Priority
		}
		if filter.AssigneeID != nil {
			assigneeUUID, _ := uuid.Parse(*filter.AssigneeID)
			storeFilter.AssigneeID = &assigneeUUID
		}
		if filter.ProjectID != nil {
			projectUUID, _ := uuid.Parse(*filter.ProjectID)
			storeFilter.ProjectID = &projectUUID
		}
		if filter.CycleID != nil {
			cycleUUID, _ := uuid.Parse(*filter.CycleID)
			storeFilter.CycleID = &cycleUUID
		}
		if filter.CreatedByID != nil {
			createdByUUID, _ := uuid.Parse(*filter.CreatedByID)
			storeFilter.CreatedByID = &createdByUUID
		}
		if len(filter.LabelIDs) > 0 {
			storeFilter.LabelIDs = make([]uuid.UUID, len(filter.LabelIDs))
			for i, l := range filter.LabelIDs {
				storeFilter.LabelIDs[i], _ = uuid.Parse(l)
			}
		}
	}

	return s.issueStore.List(ctx, teamUUID, storeFilter, page, pageSize)
}

// UpdateIssue 更新 Issue
func (s *issueService) UpdateIssue(ctx context.Context, issueID string, updates map[string]interface{}) (*model.Issue, error) {
	// 解析 Issue ID
	id, err := uuid.Parse(issueID)
	if err != nil {
		return nil, fmt.Errorf("无效的 Issue ID")
	}

	// 获取现有 Issue
	issue, err := s.issueStore.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Issue 不存在")
	}

	// 应用更新
	if title, ok := updates["title"].(string); ok {
		issue.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		issue.Description = &description
	}
	if priority, ok := updates["priority"].(int); ok {
		issue.Priority = priority
	}
	if priorityFloat, ok := updates["priority"].(float64); ok {
		issue.Priority = int(priorityFloat)
	}
	if assigneeID, ok := updates["assignee_id"].(string); ok {
		if assigneeID == "" {
			issue.AssigneeID = nil
		} else {
			assigneeUUID, _ := uuid.Parse(assigneeID)
			issue.AssigneeID = &assigneeUUID
		}
	}
	if statusID, ok := updates["status_id"].(string); ok {
		statusUUID, _ := uuid.Parse(statusID)
		issue.StatusID = statusUUID
	}

	// 保存更新
	if err := s.issueStore.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("更新 Issue 失败: %w", err)
	}

	return issue, nil
}

// DeleteIssue 删除 Issue
func (s *issueService) DeleteIssue(ctx context.Context, issueID string) error {
	// 解析 Issue ID
	id, err := uuid.Parse(issueID)
	if err != nil {
		return fmt.Errorf("无效的 Issue ID")
	}

	// 验证 Issue 存在
	_, err = s.issueStore.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("Issue 不存在")
	}

	return s.issueStore.SoftDelete(ctx, id)
}

// RestoreIssue 恢复已删除的 Issue
func (s *issueService) RestoreIssue(ctx context.Context, issueID string) error {
	// 解析 Issue ID
	id, err := uuid.Parse(issueID)
	if err != nil {
		return fmt.Errorf("无效的 Issue ID")
	}

	return s.issueStore.Restore(ctx, id)
}

// Subscribe 订阅 Issue
func (s *issueService) Subscribe(ctx context.Context, issueID string) error {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return fmt.Errorf("未认证")
	}

	// 解析 Issue ID
	id, err := uuid.Parse(issueID)
	if err != nil {
		return fmt.Errorf("无效的 Issue ID")
	}

	return s.subscriptionStore.Subscribe(ctx, id, userID)
}

// Unsubscribe 取消订阅
func (s *issueService) Unsubscribe(ctx context.Context, issueID string) error {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return fmt.Errorf("未认证")
	}

	// 解析 Issue ID
	id, err := uuid.Parse(issueID)
	if err != nil {
		return fmt.Errorf("无效的 Issue ID")
	}

	return s.subscriptionStore.Unsubscribe(ctx, id, userID)
}

// ListSubscribers 获取订阅者列表
func (s *issueService) ListSubscribers(ctx context.Context, issueID string) ([]model.User, error) {
	// 解析 Issue ID
	id, err := uuid.Parse(issueID)
	if err != nil {
		return nil, fmt.Errorf("无效的 Issue ID")
	}

	return s.subscriptionStore.ListSubscribers(ctx, id)
}

// UpdatePosition 更新 Issue 位置
func (s *issueService) UpdatePosition(ctx context.Context, issueID string, position float64, statusID *string) error {
	// 解析 Issue ID
	id, err := uuid.Parse(issueID)
	if err != nil {
		return fmt.Errorf("无效的 Issue ID")
	}

	var statusUUID *uuid.UUID
	if statusID != nil {
		parsed, err := uuid.Parse(*statusID)
		if err != nil {
			return fmt.Errorf("无效的状态 ID")
		}
		statusUUID = &parsed
	}

	return s.issueStore.UpdatePosition(ctx, id, position, statusUUID)
}
