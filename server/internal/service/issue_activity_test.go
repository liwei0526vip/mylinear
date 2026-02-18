package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"gorm.io/gorm"
)

// =============================================================================
// Issue 创建时记录活动测试 (任务 7.1)
// =============================================================================

func TestIssueService_CreateIssue_RecordsActivity(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueActivityTestFixtures(t, tx)
	ctx := fixtures.ctx

	tests := []struct {
		name         string
		params       *CreateIssueParams
		wantErr      bool
		activityType model.ActivityType
		checkFunc    func(*testing.T, *model.Issue, []model.Activity)
	}{
		{
			name: "创建 Issue 时应记录 issue_created 活动",
			params: &CreateIssueParams{
				TeamID:   fixtures.team.ID,
				Title:    "测试活动记录",
				StatusID: fixtures.status.ID,
			},
			wantErr:      false,
			activityType: model.ActivityIssueCreated,
			checkFunc: func(t *testing.T, issue *model.Issue, activities []model.Activity) {
				if len(activities) == 0 {
					t.Error("创建 Issue 时应记录活动")
					return
				}
				activity := activities[0]
				if activity.Type != model.ActivityIssueCreated {
					t.Errorf("期望活动类型 %s, 得到 %s", model.ActivityIssueCreated, activity.Type)
				}
				if activity.IssueID != issue.ID {
					t.Error("活动记录的 IssueID 不匹配")
				}
				if activity.ActorID != fixtures.userID {
					t.Error("活动记录的 ActorID 不匹配")
				}
			},
		},
		{
			name: "创建带描述的 Issue 时应记录 issue_created 活动",
			params: &CreateIssueParams{
				TeamID:      fixtures.team.ID,
				Title:       "带描述的 Issue",
				Description: strPtr("这是描述内容"),
				StatusID:    fixtures.status.ID,
			},
			wantErr:      false,
			activityType: model.ActivityIssueCreated,
			checkFunc: func(t *testing.T, issue *model.Issue, activities []model.Activity) {
				if len(activities) == 0 {
					t.Error("创建 Issue 时应记录活动")
					return
				}
				if activities[0].Type != model.ActivityIssueCreated {
					t.Errorf("期望活动类型 %s, 得到 %s", model.ActivityIssueCreated, activities[0].Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue, err := fixtures.issueService.CreateIssue(ctx, tt.params)
			if tt.wantErr {
				if err == nil {
					t.Error("期望错误但得到 nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateIssue() 错误 = %v", err)
				return
			}

			// 验证活动记录
			activities, _, err := fixtures.activityService.GetIssueActivities(ctx, issue.ID, 1, 10, nil)
			if err != nil {
				t.Errorf("GetIssueActivities() 错误 = %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, issue, activities)
			}
		})
	}
}

// =============================================================================
// Issue 更新标题/描述时记录活动测试 (任务 7.3)
// =============================================================================

func TestIssueService_UpdateIssue_TitleAndDescription(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueActivityTestFixtures(t, tx)
	ctx := fixtures.ctx

	// 先创建一个 Issue
	issue, err := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "原始标题",
		StatusID: fixtures.status.ID,
	})
	if err != nil {
		t.Fatalf("创建 Issue 失败: %v", err)
	}

	tests := []struct {
		name         string
		issueID      uuid.UUID
		updates      map[string]interface{}
		wantErr      bool
		activityType model.ActivityType
		checkFunc    func(*testing.T, *model.Issue, []model.Activity)
	}{
		{
			name:    "更新标题时应记录 title_changed 活动",
			issueID: issue.ID,
			updates: map[string]interface{}{
				"title": "更新后的标题",
			},
			wantErr:      false,
			activityType: model.ActivityTitleChanged,
			checkFunc: func(t *testing.T, updated *model.Issue, activities []model.Activity) {
				// 过滤出 title_changed 类型的活动
				var titleActivities []model.Activity
				for _, a := range activities {
					if a.Type == model.ActivityTitleChanged {
						titleActivities = append(titleActivities, a)
					}
				}
				if len(titleActivities) == 0 {
					t.Error("更新标题时应记录 title_changed 活动")
					return
				}
				activity := titleActivities[0]
				// 验证 payload
				var payload model.ActivityPayloadTitle
				if err := json.Unmarshal(activity.Payload, &payload); err != nil {
					t.Errorf("解析 payload 失败: %v", err)
					return
				}
				if payload.OldValue != "原始标题" {
					t.Errorf("期望 oldValue '原始标题', 得到 '%s'", payload.OldValue)
				}
				if payload.NewValue != "更新后的标题" {
					t.Errorf("期望 newValue '更新后的标题', 得到 '%s'", payload.NewValue)
				}
			},
		},
		{
			name:    "标题相同时不应记录活动",
			issueID: issue.ID,
			updates: map[string]interface{}{
				"title": "更新后的标题", // 与上一次相同
			},
			wantErr: false,
			checkFunc: func(t *testing.T, updated *model.Issue, activities []model.Activity) {
				// 计算所有 title_changed 活动
				var titleActivities []model.Activity
				for _, a := range activities {
					if a.Type == model.ActivityTitleChanged {
						titleActivities = append(titleActivities, a)
					}
				}
				// 应该只有一条（来自第一个测试）
				if len(titleActivities) > 1 {
					t.Errorf("标题未变更时不应记录活动，得到 %d 条", len(titleActivities))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := fixtures.issueService.UpdateIssue(ctx, tt.issueID.String(), tt.updates)
			if tt.wantErr {
				if err == nil {
					t.Error("期望错误但得到 nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateIssue() 错误 = %v", err)
				return
			}

			// 获取活动列表
			activities, _, err := fixtures.activityService.GetIssueActivities(ctx, tt.issueID, 1, 50, nil)
			if err != nil {
				t.Errorf("GetIssueActivities() 错误 = %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, updated, activities)
			}
		})
	}
}

func TestIssueService_UpdateIssue_Description(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueActivityTestFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建带描述的 Issue
	issue, err := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:      fixtures.team.ID,
		Title:       "测试描述更新",
		Description: strPtr("原始描述"),
		StatusID:    fixtures.status.ID,
	})
	if err != nil {
		t.Fatalf("创建 Issue 失败: %v", err)
	}

	tests := []struct {
		name      string
		issueID   uuid.UUID
		updates   map[string]interface{}
		wantErr   bool
		checkFunc func(*testing.T, *model.Issue, []model.Activity)
	}{
		{
			name:    "更新描述时应记录 description_changed 活动",
			issueID: issue.ID,
			updates: map[string]interface{}{
				"description": "新的描述内容",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, updated *model.Issue, activities []model.Activity) {
				var descActivities []model.Activity
				for _, a := range activities {
					if a.Type == model.ActivityDescriptionChanged {
						descActivities = append(descActivities, a)
					}
				}
				if len(descActivities) == 0 {
					t.Error("更新描述时应记录 description_changed 活动")
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := fixtures.issueService.UpdateIssue(ctx, tt.issueID.String(), tt.updates)
			if tt.wantErr {
				if err == nil {
					t.Error("期望错误但得到 nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateIssue() 错误 = %v", err)
				return
			}

			activities, _, err := fixtures.activityService.GetIssueActivities(ctx, tt.issueID, 1, 50, nil)
			if err != nil {
				t.Errorf("GetIssueActivities() 错误 = %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, updated, activities)
			}
		})
	}
}

// =============================================================================
// Issue 状态变更时记录活动和历史测试 (任务 7.5)
// =============================================================================

func TestIssueService_UpdateIssue_Status(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueActivityTestFixtures(t, tx)
	ctx := fixtures.ctx

	t.Run("更新状态时应记录 status_changed 活动", func(t *testing.T) {
		// 创建独立的 Issue 用于此测试
		issue, err := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
			TeamID:   fixtures.team.ID,
			Title:    "测试状态更新 - 第一次",
			StatusID: fixtures.status.ID,
		})
		if err != nil {
			t.Fatalf("创建 Issue 失败: %v", err)
		}

		updates := map[string]interface{}{
			"status_id": fixtures.status2.ID.String(),
		}

		updated, err := fixtures.issueService.UpdateIssue(ctx, issue.ID.String(), updates)
		if err != nil {
			t.Errorf("UpdateIssue() 错误 = %v", err)
			return
		}

		activities, _, err := fixtures.activityService.GetIssueActivities(ctx, issue.ID, 1, 50, nil)
		if err != nil {
			t.Errorf("GetIssueActivities() 错误 = %v", err)
			return
		}

		var statusActivities []model.Activity
		for _, a := range activities {
			if a.Type == model.ActivityStatusChanged {
				statusActivities = append(statusActivities, a)
			}
		}
		if len(statusActivities) == 0 {
			t.Error("更新状态时应记录 status_changed 活动")
			return
		}
		activity := statusActivities[0]
		var payload model.ActivityPayloadStatus
		if err := json.Unmarshal(activity.Payload, &payload); err != nil {
			t.Errorf("解析 payload 失败: %v", err)
			return
		}
		if payload.NewStatus == nil {
			t.Error("payload 应包含 new_status")
			return
		}
		if payload.NewStatus.ID != fixtures.status2.ID {
			t.Errorf("new_status ID 不匹配，期望 %s, 得到 %s", fixtures.status2.ID, payload.NewStatus.ID)
		}
		if updated.StatusID != fixtures.status2.ID {
			t.Errorf("updated.StatusID 不匹配，期望 %s, 得到 %s", fixtures.status2.ID, updated.StatusID)
		}
	})

	t.Run("状态相同时不应记录活动", func(t *testing.T) {
		// 创建独立的 Issue 用于此测试
		issue, err := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
			TeamID:   fixtures.team.ID,
			Title:    "测试状态更新 - 第二次",
			StatusID: fixtures.status.ID,
		})
		if err != nil {
			t.Fatalf("创建 Issue 失败: %v", err)
		}

		// 第一次更新：status -> status2
		_, err = fixtures.issueService.UpdateIssue(ctx, issue.ID.String(), map[string]interface{}{
			"status_id": fixtures.status2.ID.String(),
		})
		if err != nil {
			t.Fatalf("第一次更新失败: %v", err)
		}

		// 第二次更新：status2 -> status2 (相同)
		_, err = fixtures.issueService.UpdateIssue(ctx, issue.ID.String(), map[string]interface{}{
			"status_id": fixtures.status2.ID.String(),
		})
		if err != nil {
			t.Errorf("第二次更新失败: %v", err)
			return
		}

		activities, _, err := fixtures.activityService.GetIssueActivities(ctx, issue.ID, 1, 50, nil)
		if err != nil {
			t.Errorf("GetIssueActivities() 错误 = %v", err)
			return
		}

		var statusActivities []model.Activity
		for _, a := range activities {
			if a.Type == model.ActivityStatusChanged {
				statusActivities = append(statusActivities, a)
			}
		}
		// 应该只有第一次更新的那一条
		if len(statusActivities) != 1 {
			t.Errorf("状态未变更时不应记录活动，期望 1 条，得到 %d 条", len(statusActivities))
		}
	})
}

// =============================================================================
// Issue 负责人/优先级/截止日期变更时记录活动测试 (任务 7.7)
// =============================================================================

func TestIssueService_UpdateIssue_Priority(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueActivityTestFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建 Issue
	issue, err := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试优先级更新",
		StatusID: fixtures.status.ID,
		Priority: 0,
	})
	if err != nil {
		t.Fatalf("创建 Issue 失败: %v", err)
	}

	tests := []struct {
		name      string
		issueID   uuid.UUID
		updates   map[string]interface{}
		wantErr   bool
		checkFunc func(*testing.T, *model.Issue, []model.Activity)
	}{
		{
			name:    "更新优先级时应记录 priority_changed 活动",
			issueID: issue.ID,
			updates: map[string]interface{}{
				"priority": 2,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, updated *model.Issue, activities []model.Activity) {
				var priorityActivities []model.Activity
				for _, a := range activities {
					if a.Type == model.ActivityPriorityChanged {
						priorityActivities = append(priorityActivities, a)
					}
				}
				if len(priorityActivities) == 0 {
					t.Error("更新优先级时应记录 priority_changed 活动")
					return
				}
				activity := priorityActivities[0]
				var payload model.ActivityPayloadPriority
				if err := json.Unmarshal(activity.Payload, &payload); err != nil {
					t.Errorf("解析 payload 失败: %v", err)
					return
				}
				if payload.OldValue != 0 {
					t.Errorf("期望 oldValue 0, 得到 %d", payload.OldValue)
				}
				if payload.NewValue != 2 {
					t.Errorf("期望 newValue 2, 得到 %d", payload.NewValue)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := fixtures.issueService.UpdateIssue(ctx, tt.issueID.String(), tt.updates)
			if tt.wantErr {
				if err == nil {
					t.Error("期望错误但得到 nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateIssue() 错误 = %v", err)
				return
			}

			activities, _, err := fixtures.activityService.GetIssueActivities(ctx, tt.issueID, 1, 50, nil)
			if err != nil {
				t.Errorf("GetIssueActivities() 错误 = %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, updated, activities)
			}
		})
	}
}

func TestIssueService_UpdateIssue_Assignee(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueActivityTestFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建 Issue（无负责人）
	issue, err := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试负责人更新",
		StatusID: fixtures.status.ID,
	})
	if err != nil {
		t.Fatalf("创建 Issue 失败: %v", err)
	}

	tests := []struct {
		name      string
		issueID   uuid.UUID
		updates   map[string]interface{}
		wantErr   bool
		checkFunc func(*testing.T, *model.Issue, []model.Activity)
	}{
		{
			name:    "分配负责人时应记录 assignee_changed 活动",
			issueID: issue.ID,
			updates: map[string]interface{}{
				"assignee_id": fixtures.userID.String(),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, updated *model.Issue, activities []model.Activity) {
				var assigneeActivities []model.Activity
				for _, a := range activities {
					if a.Type == model.ActivityAssigneeChanged {
						assigneeActivities = append(assigneeActivities, a)
					}
				}
				if len(assigneeActivities) == 0 {
					t.Error("分配负责人时应记录 assignee_changed 活动")
					return
				}
			},
		},
		{
			name:    "取消分配时应记录 assignee_changed 活动",
			issueID: issue.ID,
			updates: map[string]interface{}{
				"assignee_id": "",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, updated *model.Issue, activities []model.Activity) {
				var assigneeActivities []model.Activity
				for _, a := range activities {
					if a.Type == model.ActivityAssigneeChanged {
						assigneeActivities = append(assigneeActivities, a)
					}
				}
				// 应该有 2 条（分配 + 取消分配）
				if len(assigneeActivities) < 2 {
					t.Errorf("取消分配时应记录活动，得到 %d 条", len(assigneeActivities))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := fixtures.issueService.UpdateIssue(ctx, tt.issueID.String(), tt.updates)
			if tt.wantErr {
				if err == nil {
					t.Error("期望错误但得到 nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateIssue() 错误 = %v", err)
				return
			}

			activities, _, err := fixtures.activityService.GetIssueActivities(ctx, tt.issueID, 1, 50, nil)
			if err != nil {
				t.Errorf("GetIssueActivities() 错误 = %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, updated, activities)
			}
		})
	}
}

// =============================================================================
// 测试辅助结构和函数
// =============================================================================

type issueActivityTestFixtures struct {
	ctx              context.Context
	userID           uuid.UUID
	workspaceID      uuid.UUID
	team             *model.Team
	status           *model.WorkflowState
	status2          *model.WorkflowState
	issueService     IssueService
	activityService  ActivityService
	activityStore    store.ActivityStore
	subscriptionStore store.IssueSubscriptionStore
}

func setupIssueActivityTestFixtures(t *testing.T, db *gorm.DB) *issueActivityTestFixtures {
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Activity Workspace",
		Slug: prefix + "_act_ws",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("创建工作区失败: %v", err)
	}

	// 创建用户
	userID := uuid.New()
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_act_user@example.com",
		Username:     prefix + "_actuser",
		Name:         "Activity Test User",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	user.ID = userID
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        prefix + "_Act Team",
		Key:         "ACT",
	}
	if err := db.Create(team).Error; err != nil {
		t.Fatalf("创建团队失败: %v", err)
	}

	// 添加团队成员
	member := &model.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   model.RoleAdmin,
	}
	if err := db.Create(member).Error; err != nil {
		t.Fatalf("添加团队成员失败: %v", err)
	}

	// 创建两个工作流状态
	status := &model.WorkflowState{
		TeamID:   team.ID,
		Name:     "Backlog",
		Type:     model.StateTypeBacklog,
		Color:    "#808080",
		Position: 0,
	}
	if err := db.Create(status).Error; err != nil {
		t.Fatalf("创建工作流状态失败: %v", err)
	}

	status2 := &model.WorkflowState{
		TeamID:   team.ID,
		Name:     "In Progress",
		Type:     model.StateTypeStarted,
		Color:    "#008080",
		Position: 1,
	}
	if err := db.Create(status2).Error; err != nil {
		t.Fatalf("创建工作流状态2失败: %v", err)
	}

	// 创建 stores
	issueStore := store.NewIssueStore(db)
	activityStore := store.NewActivityStore(db)
	subscriptionStore := store.NewIssueSubscriptionStore(db)
	teamMemberStore := store.NewTeamMemberStore(db)

	// 创建 services
	activityService := NewActivityService(activityStore)
	issueService := NewIssueServiceWithActivity(issueStore, subscriptionStore, teamMemberStore, activityService)

	// 创建带用户信息的 context
	userCtx := context.WithValue(ctx, "user_id", userID)
	userCtx = context.WithValue(userCtx, "workspace_id", workspace.ID)
	userCtx = context.WithValue(userCtx, "user_role", model.RoleAdmin)

	return &issueActivityTestFixtures{
		ctx:              userCtx,
		userID:           userID,
		workspaceID:      workspace.ID,
		team:             team,
		status:           status,
		status2:          status2,
		issueService:     issueService,
		activityService:  activityService,
		activityStore:    activityStore,
		subscriptionStore: subscriptionStore,
	}
}
