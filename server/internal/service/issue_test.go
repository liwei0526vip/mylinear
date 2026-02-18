package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"gorm.io/gorm"
)

// TestIssueService_Interface 测试 IssueService 接口定义存在
func TestIssueService_Interface(t *testing.T) {
	var _ IssueService = (*issueService)(nil)
}

// =============================================================================
// CreateIssue 测试
// =============================================================================

func TestIssueService_CreateIssue(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueServiceFixtures(t, tx)
	ctx := fixtures.ctx

	tests := []struct {
		name        string
		params      *CreateIssueParams
		wantErr     bool
		errContains string
		checkFunc   func(*testing.T, *model.Issue)
	}{
		{
			name: "正常创建 Issue",
			params: &CreateIssueParams{
				TeamID:   fixtures.team.ID,
				Title:    "测试 Issue",
				StatusID: fixtures.status.ID,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, issue *model.Issue) {
				if issue.Title != "测试 Issue" {
					t.Errorf("Expected title '测试 Issue', got '%s'", issue.Title)
				}
				if issue.Number != 1 {
					t.Errorf("Expected number 1, got %d", issue.Number)
				}
				// 验证创建者自动订阅
				subscribed, _ := fixtures.subscriptionStore.IsSubscribed(ctx, issue.ID, fixtures.userID)
				if !subscribed {
					t.Error("Creator should be auto-subscribed")
				}
			},
		},
		{
			name: "创建带描述的 Issue",
			params: &CreateIssueParams{
				TeamID:      fixtures.team.ID,
				Title:       "带描述的 Issue",
				Description: strPtr("这是一个描述"),
				StatusID:    fixtures.status.ID,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, issue *model.Issue) {
				if issue.Description == nil || *issue.Description != "这是一个描述" {
					t.Error("Description should be set")
				}
			},
		},
		{
			name: "未认证用户创建 Issue",
			params: &CreateIssueParams{
				TeamID:   fixtures.team.ID,
				Title:    "未认证 Issue",
				StatusID: fixtures.status.ID,
			},
			wantErr:     true,
			errContains: "未认证",
		},
		{
			name: "空标题创建 Issue",
			params: &CreateIssueParams{
				TeamID:   fixtures.team.ID,
				Title:    "",
				StatusID: fixtures.status.ID,
			},
			wantErr:     true,
			errContains: "标题",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为未认证测试创建无用户信息的 context
			testCtx := ctx
			if tt.name == "未认证用户创建 Issue" {
				testCtx = context.Background()
			}

			issue, err := fixtures.issueService.CreateIssue(testCtx, tt.params)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Error should contain '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("CreateIssue() error = %v", err)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, issue)
			}
		})
	}
}

// =============================================================================
// GetIssue 测试
// =============================================================================

func TestIssueService_GetIssue(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueServiceFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建测试 Issue
	issue, _ := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试获取 Issue",
		StatusID: fixtures.status.ID,
	})

	tests := []struct {
		name        string
		issueID     uuid.UUID
		wantErr     bool
		errContains string
	}{
		{
			name:    "正常获取 Issue",
			issueID: issue.ID,
			wantErr: false,
		},
		{
			name:        "获取不存在的 Issue",
			issueID:     uuid.New(),
			wantErr:     true,
			errContains: "不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fixtures.issueService.GetIssue(ctx, tt.issueID.String())
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetIssue() error = %v", err)
				return
			}
			if got.ID != tt.issueID {
				t.Errorf("GetIssue() ID = %v, want %v", got.ID, tt.issueID)
			}
		})
	}
}

// =============================================================================
// ListIssues 测试
// =============================================================================

func TestIssueService_ListIssues(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueServiceFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建多个测试 Issue
	for i := 0; i < 5; i++ {
		_, _ = fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
			TeamID:   fixtures.team.ID,
			Title:    "测试列表 Issue",
			StatusID: fixtures.status.ID,
			Priority: i % 3,
		})
	}

	tests := []struct {
		name        string
		teamID      uuid.UUID
		filter      *IssueFilter
		page        int
		pageSize    int
		wantCount   int
		wantTotal   int64
		wantErr     bool
	}{
		{
			name:      "获取所有 Issue",
			teamID:    fixtures.team.ID,
			filter:    nil,
			page:      1,
			pageSize:  10,
			wantCount: 5,
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:   "按优先级过滤",
			teamID: fixtures.team.ID,
			filter: &IssueFilter{Priority: intPtr(0)},
			page:   1,
			pageSize: 10,
			wantCount: 2, // i % 3 == 0 的有 i=0, i=3
			wantTotal: 2,
			wantErr: false,
		},
		{
			name:   "分页测试",
			teamID: fixtures.team.ID,
			filter: nil,
			page:   1,
			pageSize: 2,
			wantCount: 2,
			wantTotal: 5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues, total, err := fixtures.issueService.ListIssues(ctx, tt.teamID.String(), tt.filter, tt.page, tt.pageSize)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListIssues() error = %v", err)
				return
			}
			if len(issues) != tt.wantCount {
				t.Errorf("ListIssues() got %d issues, want %d", len(issues), tt.wantCount)
			}
			if total != tt.wantTotal {
				t.Errorf("ListIssues() total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

// =============================================================================
// UpdateIssue 测试
// =============================================================================

func TestIssueService_UpdateIssue(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueServiceFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建测试 Issue
	issue, _ := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试更新 Issue",
		StatusID: fixtures.status.ID,
	})

	tests := []struct {
		name        string
		issueID     uuid.UUID
		updates     map[string]interface{}
		wantErr     bool
		errContains string
		checkFunc   func(*testing.T, *model.Issue)
	}{
		{
			name:    "更新标题",
			issueID: issue.ID,
			updates: map[string]interface{}{"title": "更新后的标题"},
			wantErr: false,
			checkFunc: func(t *testing.T, got *model.Issue) {
				if got.Title != "更新后的标题" {
					t.Errorf("Expected title '更新后的标题', got '%s'", got.Title)
				}
			},
		},
		{
			name:    "更新优先级",
			issueID: issue.ID,
			updates: map[string]interface{}{"priority": 4},
			wantErr: false,
			checkFunc: func(t *testing.T, got *model.Issue) {
				if got.Priority != 4 {
					t.Errorf("Expected priority 4, got %d", got.Priority)
				}
			},
		},
		{
			name:        "更新不存在的 Issue",
			issueID:     uuid.New(),
			updates:     map[string]interface{}{"title": "不存在"},
			wantErr:     true,
			errContains: "不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fixtures.issueService.UpdateIssue(ctx, tt.issueID.String(), tt.updates)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateIssue() error = %v", err)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

// =============================================================================
// DeleteIssue 测试
// =============================================================================

func TestIssueService_DeleteIssue(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueServiceFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建测试 Issue
	issue, _ := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试删除 Issue",
		StatusID: fixtures.status.ID,
	})

	tests := []struct {
		name        string
		issueID     uuid.UUID
		wantErr     bool
		errContains string
	}{
		{
			name:    "正常删除 Issue",
			issueID: issue.ID,
			wantErr: false,
		},
		{
			name:        "删除不存在的 Issue",
			issueID:     uuid.New(),
			wantErr:     true,
			errContains: "不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fixtures.issueService.DeleteIssue(ctx, tt.issueID.String())
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteIssue() error = %v", err)
				return
			}
			// 验证软删除
			_, err = fixtures.issueService.GetIssue(ctx, tt.issueID.String())
			if err == nil {
				t.Error("Issue should be soft deleted")
			}
		})
	}
}

// =============================================================================
// Subscribe/Unsubscribe 测试
// =============================================================================

func TestIssueService_Subscribe(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueServiceFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建测试 Issue
	issue, _ := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试订阅 Issue",
		StatusID: fixtures.status.ID,
	})

	// 使用另一个用户测试订阅（创建者已自动订阅）
	anotherUserID := fixtures.user2ID
	anotherCtx := context.WithValue(context.Background(), "user_id", anotherUserID)
	anotherCtx = context.WithValue(anotherCtx, "workspace_id", fixtures.workspaceID)
	anotherCtx = context.WithValue(anotherCtx, "user_role", model.RoleMember)

	// 订阅
	err := fixtures.issueService.Subscribe(anotherCtx, issue.ID.String())
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}

	// 验证订阅状态
	subscribed, _ := fixtures.subscriptionStore.IsSubscribed(ctx, issue.ID, anotherUserID)
	if !subscribed {
		t.Error("User should be subscribed")
	}

	// 取消订阅
	err = fixtures.issueService.Unsubscribe(anotherCtx, issue.ID.String())
	if err != nil {
		t.Errorf("Unsubscribe() error = %v", err)
	}

	// 验证取消订阅状态
	subscribed, _ = fixtures.subscriptionStore.IsSubscribed(ctx, issue.ID, anotherUserID)
	if subscribed {
		t.Error("User should be unsubscribed")
	}
}

// =============================================================================
// UpdatePosition 测试
// =============================================================================

func TestIssueService_UpdatePosition(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueServiceFixtures(t, tx)
	ctx := fixtures.ctx

	// 创建测试 Issue
	issue, _ := fixtures.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试位置更新 Issue",
		StatusID: fixtures.status.ID,
	})

	newPosition := 5000.0
	err := fixtures.issueService.UpdatePosition(ctx, issue.ID.String(), newPosition, nil)
	if err != nil {
		t.Errorf("UpdatePosition() error = %v", err)
		return
	}

	// 验证位置更新
	got, _ := fixtures.issueService.GetIssue(ctx, issue.ID.String())
	if got.Position != newPosition {
		t.Errorf("Expected position %f, got %f", newPosition, got.Position)
	}
}

// =============================================================================
// 测试辅助结构和函数
// =============================================================================

type issueServiceFixtures struct {
	ctx                 context.Context
	userID              uuid.UUID
	user2ID             uuid.UUID
	workspaceID         uuid.UUID
	team                *model.Team
	status              *model.WorkflowState
	issueService        IssueService
	issueStore          store.IssueStore
	subscriptionStore   store.IssueSubscriptionStore
}

func setupIssueServiceFixtures(t *testing.T, db *gorm.DB) *issueServiceFixtures {
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Workspace",
		Slug: prefix + "_workspace",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("创建工作区失败: %v", err)
	}

	// 创建用户
	userID := uuid.New()
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user@example.com",
		Username:     prefix + "_user",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	user.ID = userID
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 创建第二个用户
	user2ID := uuid.New()
	user2 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user2@example.com",
		Username:     prefix + "_user2",
		Name:         "Test User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2.ID = user2ID
	if err := db.Create(user2).Error; err != nil {
		t.Fatalf("创建第二个用户失败: %v", err)
	}

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        prefix + "_Team",
		Key:         "TST",
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

	// 创建工作流状态
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

	// 创建 stores
	issueStore := store.NewIssueStore(db)
	subscriptionStore := store.NewIssueSubscriptionStore(db)
	teamMemberStore := store.NewTeamMemberStore(db)

	// 创建 service
	issueService := NewIssueService(issueStore, subscriptionStore, teamMemberStore)

	// 创建带用户信息的 context
	userCtx := context.WithValue(ctx, "user_id", userID)
	userCtx = context.WithValue(userCtx, "workspace_id", workspace.ID)
	userCtx = context.WithValue(userCtx, "user_role", model.RoleAdmin)

	return &issueServiceFixtures{
		ctx:               userCtx,
		userID:            userID,
		user2ID:           user2ID,
		workspaceID:       workspace.ID,
		team:              team,
		status:            status,
		issueService:      issueService,
		issueStore:        issueStore,
		subscriptionStore: subscriptionStore,
	}
}

// 辅助函数
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s[1:], substr) || s[:len(substr)] == substr)
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}
