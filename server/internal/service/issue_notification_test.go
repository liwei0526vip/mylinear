package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// =============================================================================
// 5.1 IssueService 通知触发测试
// =============================================================================

type issueNotificationTestFixtures struct {
	ctx                context.Context
	userID             uuid.UUID
	user2ID            uuid.UUID
	workspaceID        uuid.UUID
	team               *model.Team
	status             *model.WorkflowState
	status2            *model.WorkflowState
	issueService       IssueService
	notificationService NotificationService
	notificationStore   store.NotificationStore
	subscriptionStore  store.IssueSubscriptionStore
	db                 *gorm.DB
}

func setupIssueNotificationTestFixtures(t *testing.T, db *gorm.DB) *issueNotificationTestFixtures {
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Notif Workspace",
		Slug: prefix + "_notif_ws",
	}
	require.NoError(t, db.Create(workspace).Error)

	// 创建用户1
	userID := uuid.New()
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user1@example.com",
		Username:     prefix + "_user1",
		Name:         "User 1",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	user.ID = userID
	require.NoError(t, db.Create(user).Error)

	// 创建用户2
	user2ID := uuid.New()
	user2 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user2@example.com",
		Username:     prefix + "_user2",
		Name:         "User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2.ID = user2ID
	require.NoError(t, db.Create(user2).Error)

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        prefix + "_Notif Team",
		Key:         "NTF",
	}
	require.NoError(t, db.Create(team).Error)

	// 添加团队成员
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   model.RoleAdmin,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: team.ID,
		UserID: user2ID,
		Role:   model.RoleMember,
	}).Error)

	// 创建工作流状态
	status := &model.WorkflowState{
		TeamID:   team.ID,
		Name:     "Backlog",
		Type:     model.StateTypeBacklog,
		Color:    "#808080",
		Position: 0,
	}
	require.NoError(t, db.Create(status).Error)

	status2 := &model.WorkflowState{
		TeamID:   team.ID,
		Name:     "In Progress",
		Type:     model.StateTypeStarted,
		Color:    "#008080",
		Position: 1,
	}
	require.NoError(t, db.Create(status2).Error)

	// 创建 stores
	issueStore := store.NewIssueStore(db)
	subscriptionStore := store.NewIssueSubscriptionStore(db)
	notificationStore := store.NewNotificationStore(db)
	preferenceStore := store.NewNotificationPreferenceStore(db)
	userStore := store.NewUserStore(db)

	// 创建 services
	notificationService := NewNotificationService(notificationStore, preferenceStore, userStore)
	issueService := NewIssueServiceWithNotification(issueStore, subscriptionStore, nil, notificationService)

	// 创建带用户信息的 context
	userCtx := context.WithValue(ctx, "user_id", userID)
	userCtx = context.WithValue(userCtx, "workspace_id", workspace.ID)
	userCtx = context.WithValue(userCtx, "user_role", model.RoleAdmin)

	return &issueNotificationTestFixtures{
		ctx:                userCtx,
		userID:             userID,
		user2ID:            user2ID,
		workspaceID:        workspace.ID,
		team:               team,
		status:             status,
		status2:            status2,
		issueService:       issueService,
		notificationService: notificationService,
		notificationStore:   notificationStore,
		subscriptionStore:  subscriptionStore,
		db:                 db,
	}
}

// TestIssueService_Update_Notification_Trigger 测试 Issue 更新触发通知
func TestIssueService_Update_Notification_Trigger(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupIssueNotificationTestFixtures(t, tx)
	ctx := f.ctx

	// 创建 Issue（由 user1 创建，自动订阅）
	issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   f.team.ID,
		Title:    "测试通知触发",
		StatusID: f.status.ID,
	})
	require.NoError(t, err)

	t.Run("指派通知", func(t *testing.T) {
		// 清理通知
		tx.Exec("DELETE FROM notifications")

		// 指派给 user2
		updates := map[string]interface{}{
			"assignee_id": f.user2ID.String(),
		}
		_, err := f.issueService.UpdateIssue(ctx, issue.ID.String(), updates)
		require.NoError(t, err)

		// 验证 user2 收到指派通知
		var count int64
		tx.Model(&model.Notification{}).
			Where("user_id = ? AND type = ?", f.user2ID, model.NotificationTypeIssueAssigned).
			Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("状态变更通知订阅者", func(t *testing.T) {
		// 清理通知
		tx.Exec("DELETE FROM notifications")

		// 手动订阅 user2
		require.NoError(t, f.subscriptionStore.Subscribe(ctx, issue.ID, f.user2ID))

		// 更新状态
		updates := map[string]interface{}{
			"status_id": f.status2.ID.String(),
		}
		_, err := f.issueService.UpdateIssue(ctx, issue.ID.String(), updates)
		require.NoError(t, err)

		// 验证 user2 收到状态变更通知（user1 是操作者，不应收到通知）
		var count int64
		tx.Model(&model.Notification{}).
			Where("user_id = ? AND type = ?", f.user2ID, model.NotificationTypeIssueStatusChanged).
			Count(&count)
		assert.Equal(t, int64(1), count)

		// 验证 user1（操作者）没有收到通知
		var user1Count int64
		tx.Model(&model.Notification{}).
			Where("user_id = ? AND type = ?", f.userID, model.NotificationTypeIssueStatusChanged).
			Count(&user1Count)
		assert.Equal(t, int64(0), user1Count)
	})

	t.Run("优先级变更通知订阅者", func(t *testing.T) {
		// 清理通知
		tx.Exec("DELETE FROM notifications")

		// 更新优先级
		updates := map[string]interface{}{
			"priority": 2,
		}
		_, err := f.issueService.UpdateIssue(ctx, issue.ID.String(), updates)
		require.NoError(t, err)

		// 验证 user2 收到优先级变更通知
		var count int64
		tx.Model(&model.Notification{}).
			Where("user_id = ? AND type = ?", f.user2ID, model.NotificationTypeIssuePriorityChanged).
			Count(&count)
		assert.Equal(t, int64(1), count)
	})
}

// TestIssueService_Create_AutoSubscribe 测试创建 Issue 时自动订阅
func TestIssueService_Create_AutoSubscribe(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupIssueNotificationTestFixtures(t, tx)
	ctx := f.ctx

	// 创建 Issue
	issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   f.team.ID,
		Title:    "测试自动订阅",
		StatusID: f.status.ID,
	})
	require.NoError(t, err)

	// 验证创建者已自动订阅
	subscribed, err := f.subscriptionStore.IsSubscribed(ctx, issue.ID, f.userID)
	require.NoError(t, err)
	assert.True(t, subscribed, "创建者应自动订阅 Issue")
}

// TestIssueService_Update_NoSelfNotification 测试操作者不会收到自己的操作通知
func TestIssueService_Update_NoSelfNotification(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupIssueNotificationTestFixtures(t, tx)
	ctx := f.ctx

	// 创建 Issue
	issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   f.team.ID,
		Title:    "测试不通知自己",
		StatusID: f.status.ID,
	})
	require.NoError(t, err)

	// 清理通知
	tx.Exec("DELETE FROM notifications")

	// 更新状态（创建者已自动订阅）
	updates := map[string]interface{}{
		"status_id": f.status2.ID.String(),
	}
	_, err = f.issueService.UpdateIssue(ctx, issue.ID.String(), updates)
	require.NoError(t, err)

	// 验证操作者没有收到通知
	var count int64
	tx.Model(&model.Notification{}).
		Where("user_id = ? AND type = ?", f.userID, model.NotificationTypeIssueStatusChanged).
		Count(&count)
	assert.Equal(t, int64(0), count, "操作者不应收到自己操作的通知")
}

// =============================================================================
// 5.2 CommentService 通知触发测试
// =============================================================================

type commentNotificationTestFixtures struct {
	ctx                 context.Context
	userID              uuid.UUID
	user2ID             uuid.UUID
	workspaceID         uuid.UUID
	team                *model.Team
	status              *model.WorkflowState
	issue               *model.Issue
	commentService      CommentService
	notificationService NotificationService
	subscriptionStore   store.IssueSubscriptionStore
	db                  *gorm.DB
}

func setupCommentNotificationTestFixtures(t *testing.T, db *gorm.DB) *commentNotificationTestFixtures {
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_CommentNotif Workspace",
		Slug: prefix + "_cmt_ws",
	}
	require.NoError(t, db.Create(workspace).Error)

	// 创建用户1（评论者）
	userID := uuid.New()
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        fmt.Sprintf("%s_commenter@example.com", prefix),
		Username:     fmt.Sprintf("%s_commenter", prefix),
		Name:         "Commenter",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user.ID = userID
	require.NoError(t, db.Create(user).Error)

	// 创建用户2（被提及者）
	user2ID := uuid.New()
	user2 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        fmt.Sprintf("%s_mentioned@example.com", prefix),
		Username:     fmt.Sprintf("%s_mentioned", prefix),
		Name:         "Mentioned User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2.ID = user2ID
	require.NoError(t, db.Create(user2).Error)

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        prefix + "_Cmt Team",
		Key:         "CMT",
	}
	require.NoError(t, db.Create(team).Error)

	// 添加团队成员
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   model.RoleMember,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: team.ID,
		UserID: user2ID,
		Role:   model.RoleMember,
	}).Error)

	// 创建工作流状态
	status := &model.WorkflowState{
		TeamID:   team.ID,
		Name:     "Backlog",
		Type:     model.StateTypeBacklog,
		Color:    "#808080",
		Position: 0,
	}
	require.NoError(t, db.Create(status).Error)

	// 创建 stores
	issueStore := store.NewIssueStore(db)
	commentStore := store.NewCommentStore(db)
	subscriptionStore := store.NewIssueSubscriptionStore(db)
	notificationStore := store.NewNotificationStore(db)
	preferenceStore := store.NewNotificationPreferenceStore(db)
	userStore := store.NewUserStore(db)

	// 创建 services
	notificationService := NewNotificationService(notificationStore, preferenceStore, userStore)
	commentService := NewCommentServiceWithNotification(commentStore, issueStore, subscriptionStore, userStore, notificationService)

	// 创建带用户信息的 context
	userCtx := context.WithValue(ctx, "user_id", userID)
	userCtx = context.WithValue(userCtx, "workspace_id", workspace.ID)
	userCtx = context.WithValue(userCtx, "user_role", model.RoleMember)

	// 创建一个测试 Issue
	issueStore.Create(ctx, &model.Issue{
		TeamID:   team.ID,
		Title:    "测试评论通知",
		StatusID: status.ID,
		CreatedByID: userID,
	})

	// 获取创建的 Issue
	issues, _, _ := issueStore.List(ctx, team.ID, nil, 1, 1)
	var testIssue *model.Issue
	if len(issues) > 0 {
		testIssue = &issues[0]
	}

	return &commentNotificationTestFixtures{
		ctx:                 userCtx,
		userID:              userID,
		user2ID:             user2ID,
		workspaceID:         workspace.ID,
		team:                team,
		status:              status,
		issue:               testIssue,
		commentService:      commentService,
		notificationService: notificationService,
		subscriptionStore:   subscriptionStore,
		db:                  db,
	}
}

// TestCommentService_Mention_Parse 测试 @mention 解析
func TestCommentService_Mention_Parse(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantMentions  []string
	}{
		{
			name:         "单个 mention",
			body:         "这是给 @alice 的评论",
			wantMentions: []string{"alice"},
		},
		{
			name:         "多个 mention",
			body:         "@alice 和 @bob 请看一下",
			wantMentions: []string{"alice", "bob"},
		},
		{
			name:         "无效 username",
			body:         "@nonexistent_user 请看一下",
			wantMentions: []string{"nonexistent_user"},
		},
		{
			name:         "无 mention",
			body:         "这是普通评论",
			wantMentions: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractUniqueMentions(tt.body)
			assert.Equal(t, len(tt.wantMentions), len(got))
		})
	}
}

// TestCommentService_Create_MentionNotification 测试评论 @mention 触发通知
func TestCommentService_Create_MentionNotification(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupCommentNotificationTestFixtures(t, tx)
	ctx := f.ctx
	require.NotNil(t, f.issue, "测试 Issue 应该存在")

	// 清理通知
	tx.Exec("DELETE FROM notifications")

	// 创建带有 @mention 的评论（使用 user2 的用户名）
	user2, err := store.NewUserStore(tx).GetUserByID(ctx, f.user2ID.String())
	require.NoError(t, err)

	body := fmt.Sprintf("@%s 请查看这个 Issue", user2.Username)
	_, err = f.commentService.CreateComment(ctx, f.issue.ID, f.userID, body, nil)
	require.NoError(t, err)

	// 验证 user2 收到 mention 通知
	var count int64
	tx.Model(&model.Notification{}).
		Where("user_id = ? AND type = ?", f.user2ID, model.NotificationTypeIssueMentioned).
		Count(&count)
	assert.Equal(t, int64(1), count)

	// 验证评论者（user1）没有收到通知
	var user1Count int64
	tx.Model(&model.Notification{}).
		Where("user_id = ? AND type = ?", f.userID, model.NotificationTypeIssueMentioned).
		Count(&user1Count)
	assert.Equal(t, int64(0), user1Count)
}

// TestCommentService_Create_SubscriberNotification 测试评论通知订阅者
func TestCommentService_Create_SubscriberNotification(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupCommentNotificationTestFixtures(t, tx)
	ctx := f.ctx
	require.NotNil(t, f.issue, "测试 Issue 应该存在")

	// 手动订阅 user2
	require.NoError(t, f.subscriptionStore.Subscribe(ctx, f.issue.ID, f.user2ID))

	// 清理通知
	tx.Exec("DELETE FROM notifications")

	// 创建评论（不带 mention）
	_, err := f.commentService.CreateComment(ctx, f.issue.ID, f.userID, "这是一条普通评论", nil)
	require.NoError(t, err)

	// 验证 user2 收到评论通知
	var count int64
	tx.Model(&model.Notification{}).
		Where("user_id = ? AND type = ?", f.user2ID, model.NotificationTypeIssueCommented).
		Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestCommentService_Create_AutoSubscribeCommenter 测试评论者自动订阅
func TestCommentService_Create_AutoSubscribeCommenter(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupCommentNotificationTestFixtures(t, tx)
	ctx := f.ctx
	require.NotNil(t, f.issue, "测试 Issue 应该存在")

	// 确保评论者尚未订阅
	f.subscriptionStore.Unsubscribe(ctx, f.issue.ID, f.userID)

	// 创建评论
	_, err := f.commentService.CreateComment(ctx, f.issue.ID, f.userID, "我发表了一条评论", nil)
	require.NoError(t, err)

	// 验证评论者已自动订阅
	subscribed, err := f.subscriptionStore.IsSubscribed(ctx, f.issue.ID, f.userID)
	require.NoError(t, err)
	assert.True(t, subscribed, "评论者应自动订阅 Issue")
}
