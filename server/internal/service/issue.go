// Package service 提供业务逻辑层
package service

import (
	"context"
	"encoding/json"
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
	issueStore         store.IssueStore
	subscriptionStore  store.IssueSubscriptionStore
	teamMemberStore    store.TeamMemberStore
	activityService    ActivityService
	notificationService NotificationService
}

// NewIssueService 创建 Issue 服务实例
func NewIssueService(issueStore store.IssueStore, subscriptionStore store.IssueSubscriptionStore, teamMemberStore store.TeamMemberStore) IssueService {
	return &issueService{
		issueStore:        issueStore,
		subscriptionStore: subscriptionStore,
		teamMemberStore:   teamMemberStore,
		activityService:   nil, // 不记录活动
	}
}

// NewIssueServiceWithActivity 创建带活动记录的 Issue 服务实例
func NewIssueServiceWithActivity(issueStore store.IssueStore, subscriptionStore store.IssueSubscriptionStore, teamMemberStore store.TeamMemberStore, activityService ActivityService) IssueService {
	return &issueService{
		issueStore:        issueStore,
		subscriptionStore: subscriptionStore,
		teamMemberStore:   teamMemberStore,
		activityService:   activityService,
	}
}

// NewIssueServiceWithNotification 创建带通知功能的 Issue 服务实例
func NewIssueServiceWithNotification(issueStore store.IssueStore, subscriptionStore store.IssueSubscriptionStore, teamMemberStore store.TeamMemberStore, notificationService NotificationService) IssueService {
	return &issueService{
		issueStore:         issueStore,
		subscriptionStore:  subscriptionStore,
		teamMemberStore:    teamMemberStore,
		notificationService: notificationService,
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

	// 记录 Issue 创建活动
	if s.activityService != nil {
		activity := &model.Activity{
			IssueID: issue.ID,
			Type:    model.ActivityIssueCreated,
			ActorID: userID,
		}
		if err := s.activityService.RecordActivity(ctx, activity); err != nil {
			// 活动记录失败不影响创建
		}
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

	// 获取当前用户信息
	userID, _ := ctx.Value("user_id").(uuid.UUID)

	// 获取现有 Issue
	issue, err := s.issueStore.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Issue 不存在")
	}

	// 记录变更前的值用于活动记录
	oldTitle := issue.Title
	oldDescription := issue.Description
	oldStatusID := issue.StatusID
	oldPriority := issue.Priority
	oldAssigneeID := issue.AssigneeID

	// 应用更新
	hasTitleChange := false
	hasDescriptionChange := false
	hasStatusChange := false
	hasPriorityChange := false
	hasAssigneeChange := false

	if title, ok := updates["title"].(string); ok {
		if title != oldTitle {
			issue.Title = title
			hasTitleChange = true
		}
	}
	if description, ok := updates["description"].(string); ok {
		oldDesc := ""
		if oldDescription != nil {
			oldDesc = *oldDescription
		}
		if description != oldDesc {
			issue.Description = &description
			hasDescriptionChange = true
		}
	}
	if priority, ok := updates["priority"].(int); ok {
		if priority != oldPriority {
			issue.Priority = priority
			hasPriorityChange = true
		}
	}
	if priorityFloat, ok := updates["priority"].(float64); ok {
		priority := int(priorityFloat)
		if priority != oldPriority {
			issue.Priority = priority
			hasPriorityChange = true
		}
	}
	if assigneeID, ok := updates["assignee_id"].(string); ok {
		var newAssigneeID *uuid.UUID
		if assigneeID != "" {
			parsed, _ := uuid.Parse(assigneeID)
			newAssigneeID = &parsed
		}
		// 比较是否变更
		if !uuidPtrEqual(oldAssigneeID, newAssigneeID) {
			issue.AssigneeID = newAssigneeID
			hasAssigneeChange = true
		}
	}
	if statusID, ok := updates["status_id"].(string); ok {
		statusUUID, _ := uuid.Parse(statusID)
		if statusUUID != oldStatusID {
			issue.StatusID = statusUUID
			hasStatusChange = true
		}
	}

	// 保存更新
	if err := s.issueStore.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("更新 Issue 失败: %w", err)
	}

	// 记录活动
	if s.activityService != nil {
		// 标题变更
		if hasTitleChange {
			s.recordActivity(ctx, issue.ID, userID, model.ActivityTitleChanged, &model.ActivityPayloadTitle{
				OldValue: oldTitle,
				NewValue: issue.Title,
			})
		}

		// 描述变更
		if hasDescriptionChange {
			oldDesc := ""
			if oldDescription != nil {
				oldDesc = *oldDescription
			}
			newDesc := ""
			if issue.Description != nil {
				newDesc = *issue.Description
			}
			s.recordActivity(ctx, issue.ID, userID, model.ActivityDescriptionChanged, &model.ActivityPayloadDescription{
				OldValue: oldDesc,
				NewValue: newDesc,
			})
		}

		// 状态变更
		if hasStatusChange {
			// TODO: 查询状态详情以填充名称和颜色
			s.recordActivity(ctx, issue.ID, userID, model.ActivityStatusChanged, &model.ActivityPayloadStatus{
				NewStatus: &model.ActivityStatusRef{
					ID: issue.StatusID,
				},
			})
		}

		// 优先级变更
		if hasPriorityChange {
			s.recordActivity(ctx, issue.ID, userID, model.ActivityPriorityChanged, &model.ActivityPayloadPriority{
				OldValue: oldPriority,
				NewValue: issue.Priority,
			})
		}

		// 负责人变更
		if hasAssigneeChange {
			payload := &model.ActivityPayloadAssignee{}
			if oldAssigneeID != nil {
				payload.OldAssignee = &model.ActivityPayloadUser{
					ID: *oldAssigneeID,
				}
			}
			if issue.AssigneeID != nil {
				payload.NewAssignee = &model.ActivityPayloadUser{
					ID: *issue.AssigneeID,
				}
			}
			s.recordActivity(ctx, issue.ID, userID, model.ActivityAssigneeChanged, payload)
		}
	}

	// 触发通知
	if s.notificationService != nil {
		// 获取订阅者列表
		subscribers, _ := s.subscriptionStore.ListSubscribers(ctx, issue.ID)
		subscriberIDs := make([]uuid.UUID, len(subscribers))
		for i, sub := range subscribers {
			subscriberIDs[i] = sub.ID
		}

		// 指派通知
		if hasAssigneeChange {
			_ = s.notificationService.NotifyIssueAssigned(ctx, userID, issue.AssigneeID, issue.ID, issue.Title)
		}

		// 状态变更通知订阅者
		if hasStatusChange {
			_ = s.notificationService.NotifySubscribers(ctx, userID, subscriberIDs, model.NotificationTypeIssueStatusChanged, issue.ID, "Issue 状态已更新", fmt.Sprintf("Issue «%s» 状态已变更", issue.Title))
		}

		// 优先级变更通知订阅者
		if hasPriorityChange {
			_ = s.notificationService.NotifySubscribers(ctx, userID, subscriberIDs, model.NotificationTypeIssuePriorityChanged, issue.ID, "Issue 优先级已更新", fmt.Sprintf("Issue «%s» 优先级已变更", issue.Title))
		}
	}

	return issue, nil
}

// recordActivity 记录活动的辅助方法
func (s *issueService) recordActivity(ctx context.Context, issueID, actorID uuid.UUID, activityType model.ActivityType, payload interface{}) {
	if s.activityService == nil {
		return
	}

	activity := &model.Activity{
		IssueID: issueID,
		Type:    activityType,
		ActorID: actorID,
	}

	if payload != nil {
		payloadBytes, err := jsonMarshal(payload)
		if err == nil {
			activity.Payload = payloadBytes
		}
	}

	_ = s.activityService.RecordActivity(ctx, activity)
}

// uuidPtrEqual 比较两个 uuid 指针是否相等
func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// jsonMarshal 序列化 JSON
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
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
