package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"gorm.io/gorm"
)

// TestCommentService_Interface 测试 CommentService 接口定义存在
func TestCommentService_Interface(t *testing.T) {
	var _ CommentService = (*commentService)(nil)
}

// =============================================================================
// CreateComment 测试
// =============================================================================

func TestCommentService_CreateComment(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	// 创建 stores
	commentStore := store.NewCommentStore(tx)
	issueStore := store.NewIssueStore(tx)
	subscriptionStore := store.NewIssueSubscriptionStore(tx)
	userStore := store.NewUserStore(tx)

	svc := NewCommentService(commentStore, issueStore, subscriptionStore, userStore)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentServiceTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]

	tests := []struct {
		name     string
		issueID  uuid.UUID
		userID   uuid.UUID
		body     string
		parentID *uuid.UUID
		wantErr  bool
	}{
		{
			name:     "正常创建评论",
			issueID:  issue1.ID,
			userID:   user1.ID,
			body:     "这是测试评论",
			parentID: nil,
			wantErr:  false,
		},
		{
			name:     "评论者自动订阅",
			issueID:  issue1.ID,
			userID:   user2.ID,
			body:     "新用户的评论",
			parentID: nil,
			wantErr:  false,
		},
		{
			name:     "评论带 @mention（使用动态用户名）",
			issueID:  issue1.ID,
			userID:   user1.ID,
			body:     "@" + user2.Username + " 看看这个",
			parentID: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment, err := svc.CreateComment(ctx, tt.issueID, tt.userID, tt.body, tt.parentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if comment == nil {
					t.Error("CreateComment() 返回 nil")
					return
				}
				if comment.Body != tt.body {
					t.Errorf("CreateComment() Body = %v, want %v", comment.Body, tt.body)
				}
				// 验证评论者自动订阅
				subscribed, _ := subscriptionStore.IsSubscribed(ctx, tt.issueID, tt.userID)
				if !subscribed {
					t.Error("CreateComment() 评论者应该自动订阅")
				}
			}
		})
	}
}

// =============================================================================
// UpdateComment 测试
// =============================================================================

func TestCommentService_UpdateComment(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	commentStore := store.NewCommentStore(tx)
	issueStore := store.NewIssueStore(tx)
	subscriptionStore := store.NewIssueSubscriptionStore(tx)
	userStore := store.NewUserStore(tx)

	svc := NewCommentService(commentStore, issueStore, subscriptionStore, userStore)
	ctx := context.Background()

	fixtures := setupCommentServiceTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]

	// 创建测试评论
	comment, _ := svc.CreateComment(ctx, issue1.ID, user1.ID, "原始内容", nil)

	tests := []struct {
		name        string
		commentID   uuid.UUID
		userID      uuid.UUID
		newBody     string
		wantErr     bool
		errContains string
	}{
		{
			name:      "作者更新评论",
			commentID: comment.ID,
			userID:    user1.ID,
			newBody:   "更新后的内容",
			wantErr:   false,
		},
		{
			name:        "非作者无法更新",
			commentID:   comment.ID,
			userID:      user2.ID,
			newBody:     "尝试更新",
			wantErr:     true,
			errContains: "权限",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := svc.UpdateComment(ctx, tt.commentID, tt.userID, tt.newBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if updated.Body != tt.newBody {
					t.Errorf("UpdateComment() Body = %v, want %v", updated.Body, tt.newBody)
				}
				if updated.EditedAt == nil {
					t.Error("UpdateComment() EditedAt 应该被设置")
				}
			}
		})
	}
}

// =============================================================================
// DeleteComment 测试
// =============================================================================

func TestCommentService_DeleteComment(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	commentStore := store.NewCommentStore(tx)
	issueStore := store.NewIssueStore(tx)
	subscriptionStore := store.NewIssueSubscriptionStore(tx)
	userStore := store.NewUserStore(tx)

	svc := NewCommentService(commentStore, issueStore, subscriptionStore, userStore)
	ctx := context.Background()

	fixtures := setupCommentServiceTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]

	// 创建测试评论
	comment, _ := svc.CreateComment(ctx, issue1.ID, user1.ID, "测试评论", nil)

	tests := []struct {
		name    string
		comment *model.Comment
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "非作者无法删除",
			comment: comment,
			userID:  user2.ID,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteComment(ctx, tt.comment.ID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteComment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// 测试辅助函数
// =============================================================================

type commentServiceTestFixtures struct {
	users     []*model.User
	issues    []*model.Issue
	team      *model.Team
	status    *model.WorkflowState
	workspace *model.Workspace
}

func setupCommentServiceTestFixtures(t *testing.T, db *gorm.DB) *commentServiceTestFixtures {
	ctx := context.Background()
	userStore := store.NewUserStore(db)
	teamStore := store.NewTeamStore(db)
	issueStore := store.NewIssueStore(db)

	prefix := uuid.New().String()[:8]

	// 先创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Comment Service Workspace",
		Slug: prefix + "_cmt_svc",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("创建工作区失败: %v", err)
	}

	// 创建用户（用户名用于 @mention 测试）
	users := make([]*model.User, 2)
	for i := 0; i < 2; i++ {
		user := &model.User{
			WorkspaceID:  workspace.ID,
			Email:        prefix + "_svc_user" + string(rune('1'+i)) + "@example.com",
			Username:     prefix + "_testuser" + string(rune('1'+i)), // 添加前缀避免冲突
			Name:         "Test Service User " + string(rune('1'+i)),
			PasswordHash: "hash",
			Role:         model.RoleMember,
		}
		if err := userStore.CreateUser(ctx, user); err != nil {
			t.Fatalf("创建用户失败: %v", err)
		}
		users[i] = user
	}

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        prefix + "_Svc Team",
		Key:         "SVC",
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
	issues := make([]*model.Issue, 1)
	issue := &model.Issue{
		TeamID:      team.ID,
		Title:       prefix + "_Svc Issue",
		StatusID:    status.ID,
		Priority:    0,
		CreatedByID: users[0].ID,
	}
	if err := issueStore.Create(ctx, issue); err != nil {
		t.Fatalf("创建 Issue 失败: %v", err)
	}
	issues[0] = issue

	return &commentServiceTestFixtures{
		users:     users,
		issues:    issues,
		team:      team,
		status:    status,
		workspace: workspace,
	}
}
