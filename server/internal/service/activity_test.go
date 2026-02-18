package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TestActivityService_Interface 测试 ActivityService 接口定义存在
func TestActivityService_Interface(t *testing.T) {
	var _ ActivityService = (*activityService)(nil)
}

// =============================================================================
// RecordActivity 测试
// =============================================================================

func TestActivityService_RecordActivity(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	activityStore := store.NewActivityStore(tx)
	svc := NewActivityService(activityStore)
	ctx := context.Background()

	fixtures := setupActivityServiceTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	tests := []struct {
		name    string
		activity *model.Activity
		wantErr  bool
	}{
		{
			name: "记录 issue_created 活动",
			activity: &model.Activity{
				IssueID: issue1.ID,
				Type:    model.ActivityIssueCreated,
				ActorID: user1.ID,
			},
			wantErr: false,
		},
		{
			name: "记录 title_changed 活动（带 payload）",
			activity: &model.Activity{
				IssueID: issue1.ID,
				Type:    model.ActivityTitleChanged,
				ActorID: user1.ID,
				Payload: datatypes.JSON(`{"old_value":"旧标题","new_value":"新标题"}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.RecordActivity(ctx, tt.activity)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordActivity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.activity.ID == uuid.Nil {
					t.Error("RecordActivity() 未生成 ID")
				}
			}
		})
	}
}

// =============================================================================
// GetIssueActivities 测试
// =============================================================================

func TestActivityService_GetIssueActivities(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	activityStore := store.NewActivityStore(tx)
	svc := NewActivityService(activityStore)
	ctx := context.Background()

	fixtures := setupActivityServiceTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	issue2 := fixtures.issues[1]
	user1 := fixtures.users[0]

	// 创建多个活动
	for i := 0; i < 5; i++ {
		_ = svc.RecordActivity(ctx, &model.Activity{
			IssueID: issue1.ID,
			Type:    model.ActivityCommentAdded,
			ActorID: user1.ID,
		})
	}

	tests := []struct {
		name      string
		issueID   uuid.UUID
		page      int
		pageSize  int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "获取所有活动",
			issueID:   issue1.ID,
			page:      1,
			pageSize:  50,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name:      "无活动的 Issue",
			issueID:   issue2.ID,
			page:      1,
			pageSize:  50,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "分页测试",
			issueID:   issue1.ID,
			page:      1,
			pageSize:  2,
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activities, total, err := svc.GetIssueActivities(ctx, tt.issueID, tt.page, tt.pageSize, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIssueActivities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(activities) != tt.wantCount {
				t.Errorf("GetIssueActivities() got %d activities, want %d", len(activities), tt.wantCount)
			}
			if tt.name == "获取所有活动" && total != 5 {
				t.Errorf("GetIssueActivities() got total %d, want 5", total)
			}
		})
	}
}

// =============================================================================
// 测试辅助函数
// =============================================================================

type activityServiceTestFixtures struct {
	users  []*model.User
	issues []*model.Issue
	team   *model.Team
	status *model.WorkflowState
}

func setupActivityServiceTestFixtures(t *testing.T, db *gorm.DB) *activityServiceTestFixtures {
	ctx := context.Background()
	userStore := store.NewUserStore(db)
	teamStore := store.NewTeamStore(db)
	issueStore := store.NewIssueStore(db)

	prefix := uuid.New().String()[:8]

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Activity Service Workspace",
		Slug: prefix + "_act_svc",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("创建工作区失败: %v", err)
	}

	// 创建用户
	users := make([]*model.User, 1)
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_act_svc_user@example.com",
		Username:     prefix + "_actuser",
		Name:         "Test Activity User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	users[0] = user

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        prefix + "_Act Team",
		Key:         "ACT",
	}
	if err := teamStore.Create(ctx, team); err != nil {
		t.Fatalf("创建团队失败: %v", err)
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

	// 创建 Issue
	issues := make([]*model.Issue, 2)
	for i := 0; i < 2; i++ {
		issue := &model.Issue{
			TeamID:      team.ID,
			Title:       prefix + "_Act Issue " + string(rune('1'+i)),
			StatusID:    status.ID,
			Priority:    0,
			CreatedByID: user.ID,
		}
		if err := issueStore.Create(ctx, issue); err != nil {
			t.Fatalf("创建 Issue 失败: %v", err)
		}
		issues[i] = issue
	}

	return &activityServiceTestFixtures{
		users:  users,
		issues: issues,
		team:   team,
		status: status,
	}
}
