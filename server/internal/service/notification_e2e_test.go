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
// E2E 测试：通知完整流程验证
// =============================================================================

type e2eTestFixtures struct {
	ctx               context.Context
	db                *gorm.DB
	workspaceID       uuid.UUID
	user1ID           uuid.UUID
	user2ID           uuid.UUID
	teamID            uuid.UUID
	statusID          uuid.UUID
	status2ID         uuid.UUID

	issueService         IssueService
	commentService       CommentService
	notificationService  NotificationService
	preferenceService    NotificationPreferenceService

	issueStore        store.IssueStore
	commentStore      store.CommentStore
	notificationStore store.NotificationStore
	preferenceStore   store.NotificationPreferenceStore
	subscriptionStore store.IssueSubscriptionStore
	userStore         store.UserStore
}

func setupE2ETestFixtures(t *testing.T, db *gorm.DB) *e2eTestFixtures {
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_E2E Workspace",
		Slug: prefix + "_e2e_ws",
	}
	require.NoError(t, db.Create(workspace).Error)

	// 创建用户1（操作者）
	user1ID := uuid.New()
	user1 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user1@example.com",
		Username:     prefix + "_user1",
		Name:         "User 1",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	user1.ID = user1ID
	require.NoError(t, db.Create(user1).Error)

	// 创建用户2（被指派者/被提及者）
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
		Name:        prefix + "_E2E Team",
		Key:         "E2E",
	}
	require.NoError(t, db.Create(team).Error)

	// 添加团队成员
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: team.ID,
		UserID: user1ID,
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
	commentStore := store.NewCommentStore(db)
	notificationStore := store.NewNotificationStore(db)
	preferenceStore := store.NewNotificationPreferenceStore(db)
	subscriptionStore := store.NewIssueSubscriptionStore(db)
	userStore := store.NewUserStore(db)

	// 创建 services
	notificationService := NewNotificationService(notificationStore, preferenceStore, userStore)
	preferenceService := NewNotificationPreferenceService(preferenceStore)
	issueService := NewIssueServiceWithNotification(issueStore, subscriptionStore, nil, notificationService)
	commentService := NewCommentServiceWithNotification(commentStore, issueStore, subscriptionStore, userStore, notificationService)

	// 创建带用户信息的 context
	userCtx := context.WithValue(ctx, "user_id", user1ID)
	userCtx = context.WithValue(userCtx, "workspace_id", workspace.ID)
	userCtx = context.WithValue(userCtx, "user_role", model.RoleAdmin)

	return &e2eTestFixtures{
		ctx:               userCtx,
		db:                db,
		workspaceID:       workspace.ID,
		user1ID:           user1ID,
		user2ID:           user2ID,
		teamID:            team.ID,
		statusID:          status.ID,
		status2ID:         status2.ID,
		issueService:      issueService,
		commentService:    commentService,
		notificationService: notificationService,
		preferenceService: preferenceService,
		issueStore:        issueStore,
		commentStore:      commentStore,
		notificationStore: notificationStore,
		preferenceStore:   preferenceStore,
		subscriptionStore: subscriptionStore,
		userStore:         userStore,
	}
}

// =============================================================================
// 7.1.1 E2E 测试：Issue 指派 → 生成通知 → 查询通知列表
// =============================================================================

func TestE2E_IssueAssignment_NotificationFlow(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupE2ETestFixtures(t, tx)
	ctx := f.ctx

	// Step 1: 创建 Issue（无指派）
	issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   f.teamID,
		Title:    "E2E 测试 Issue",
		StatusID: f.statusID,
	})
	require.NoError(t, err)
	require.NotNil(t, issue)

	// Step 2: 指派给 user2
	updated, err := f.issueService.UpdateIssue(ctx, issue.ID.String(), map[string]interface{}{
		"assignee_id": f.user2ID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, f.user2ID, *updated.AssigneeID)

	// Step 3: 验证 user2 收到通知
	notifications, total, err := f.notificationService.ListNotifications(ctx, f.user2ID, 1, 10, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, notifications, 1)

	notification := notifications[0]
	assert.Equal(t, model.NotificationTypeIssueAssigned, notification.Type)
	assert.Equal(t, issue.ID, *notification.ResourceID)
	assert.Contains(t, notification.Title, "分配")

	// Step 4: 验证 user1（操作者）没有收到通知
	user1Notifications, user1Total, err := f.notificationService.ListNotifications(ctx, f.user1ID, 1, 10, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(0), user1Total)
	assert.Len(t, user1Notifications, 0)

	t.Log("✅ E2E 测试通过：Issue 指派 → 生成通知 → 查询通知列表")
}

// =============================================================================
// 7.1.2 E2E 测试：评论 @mention → 生成通知 → 标记已读
// =============================================================================

func TestE2E_CommentMention_NotificationFlow(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupE2ETestFixtures(t, tx)
	ctx := f.ctx

	// Step 1: 创建 Issue
	issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   f.teamID,
		Title:    "E2E 测试 @mention",
		StatusID: f.statusID,
	})
	require.NoError(t, err)

	// Step 2: 获取 user2 的用户名
	user2, err := f.userStore.GetUserByID(ctx, f.user2ID.String())
	require.NoError(t, err)

	// Step 3: 创建带有 @mention 的评论
	comment, err := f.commentService.CreateComment(ctx, issue.ID, f.user1ID, fmt.Sprintf("@%s 请查看这个问题", user2.Username), nil)
	require.NoError(t, err)
	require.NotNil(t, comment)

	// Step 4: 验证 user2 收到 mention 通知
	notifications, total, err := f.notificationService.ListNotifications(ctx, f.user2ID, 1, 10, nil, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))

	// 找到 mention 通知
	var mentionNotification *model.Notification
	for _, n := range notifications {
		if n.Type == model.NotificationTypeIssueMentioned {
			mentionNotification = &n
			break
		}
	}
	require.NotNil(t, mentionNotification, "应该收到 mention 通知")

	// Step 5: 验证未读数量
	unreadCount, err := f.notificationService.GetUnreadCount(ctx, f.user2ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, unreadCount, int64(1))

	// Step 6: 标记已读
	err = f.notificationService.MarkAsRead(ctx, mentionNotification.ID, f.user2ID)
	require.NoError(t, err)

	// Step 7: 验证已标记已读
	updatedNotification, err := f.notificationStore.GetNotificationByID(ctx, mentionNotification.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedNotification.ReadAt)

	t.Log("✅ E2E 测试通过：评论 @mention → 生成通知 → 标记已读")
}

// =============================================================================
// 7.1.3 E2E 测试：Issue 状态变更 → 通知订阅者 → 未读计数更新
// =============================================================================

func TestE2E_StatusChange_SubscriberNotification(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupE2ETestFixtures(t, tx)
	ctx := f.ctx

	// Step 1: 创建 Issue
	issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   f.teamID,
		Title:    "E2E 测试状态变更",
		StatusID: f.statusID,
	})
	require.NoError(t, err)

	// Step 2: user2 订阅该 Issue
	err = f.subscriptionStore.Subscribe(ctx, issue.ID, f.user2ID)
	require.NoError(t, err)

	// Step 3: 验证初始未读数量
	initialUnreadCount, err := f.notificationService.GetUnreadCount(ctx, f.user2ID)
	require.NoError(t, err)

	// Step 4: user1 更新 Issue 状态
	_, err = f.issueService.UpdateIssue(ctx, issue.ID.String(), map[string]interface{}{
		"status_id": f.status2ID.String(),
	})
	require.NoError(t, err)

	// Step 5: 验证 user2 收到状态变更通知
	notifications, total, err := f.notificationService.ListNotifications(ctx, f.user2ID, 1, 10, nil, []model.NotificationType{model.NotificationTypeIssueStatusChanged})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	require.GreaterOrEqual(t, len(notifications), 1)

	// Step 6: 验证未读计数更新
	updatedUnreadCount, err := f.notificationService.GetUnreadCount(ctx, f.user2ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, updatedUnreadCount, initialUnreadCount+1)

	// Step 7: 验证 user1（操作者）没有收到通知
	user1Notifications, _, err := f.notificationService.ListNotifications(ctx, f.user1ID, 1, 10, nil, []model.NotificationType{model.NotificationTypeIssueStatusChanged})
	require.NoError(t, err)
	assert.Len(t, user1Notifications, 0)

	t.Log("✅ E2E 测试通过：Issue 状态变更 → 通知订阅者 → 未读计数更新")
}

// =============================================================================
// 7.1.4 E2E 测试：通知配置禁用 → 不生成通知
// =============================================================================

func TestE2E_PreferenceDisabled_NoNotification(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupE2ETestFixtures(t, tx)
	ctx := f.ctx

	// Step 1: 禁用 user2 的指派通知
	falseVal := false
	pref := &model.NotificationPreference{
		UserID:  f.user2ID,
		Channel: model.NotificationChannelInApp,
		Type:    model.NotificationTypeIssueAssigned,
		Enabled: &falseVal,
	}
	err := f.preferenceStore.Upsert(ctx, pref)
	require.NoError(t, err)

	// Step 2: 验证配置已禁用
	enabled, err := f.preferenceStore.IsEnabled(ctx, f.user2ID, model.NotificationChannelInApp, model.NotificationTypeIssueAssigned)
	require.NoError(t, err)
	assert.False(t, enabled)

	// Step 3: 创建 Issue
	issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   f.teamID,
		Title:    "E2E 测试禁用通知",
		StatusID: f.statusID,
	})
	require.NoError(t, err)

	// Step 4: 指派给 user2
	_, err = f.issueService.UpdateIssue(ctx, issue.ID.String(), map[string]interface{}{
		"assignee_id": f.user2ID.String(),
	})
	require.NoError(t, err)

	// Step 5: 验证 user2 没有收到指派通知
	notifications, total, err := f.notificationService.ListNotifications(ctx, f.user2ID, 1, 10, nil, []model.NotificationType{model.NotificationTypeIssueAssigned})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, notifications, 0)

	t.Log("✅ E2E 测试通过：通知配置禁用 → 不生成通知")
}

// =============================================================================
// 7.2 前端功能验证（通过 API 集成测试模拟）
// =============================================================================

// TestE2E_MarkAllAsRead 验证全部标记已读功能
func TestE2E_MarkAllAsRead(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupE2ETestFixtures(t, tx)
	ctx := f.ctx

	// Step 1: 创建多个 Issue 并指派给 user2
	for i := 0; i < 3; i++ {
		issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
			TeamID:   f.teamID,
			Title:    fmt.Sprintf("E2E 批量测试 %d", i),
			StatusID: f.statusID,
		})
		require.NoError(t, err)

		_, err = f.issueService.UpdateIssue(ctx, issue.ID.String(), map[string]interface{}{
			"assignee_id": f.user2ID.String(),
		})
		require.NoError(t, err)
	}

	// Step 2: 验证有 3 条未读通知
	unreadCount, err := f.notificationService.GetUnreadCount(ctx, f.user2ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), unreadCount)

	// Step 3: 全部标记已读
	marked, err := f.notificationService.MarkAllAsRead(ctx, f.user2ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), marked)

	// Step 4: 验证未读数量为 0
	unreadCount, err = f.notificationService.GetUnreadCount(ctx, f.user2ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), unreadCount)

	t.Log("✅ E2E 测试通过：全部标记已读功能正常工作")
}

// TestE2E_BatchMarkAsRead 验证批量标记已读功能
func TestE2E_BatchMarkAsRead(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupE2ETestFixtures(t, tx)
	ctx := f.ctx

	// Step 1: 创建 5 个通知
	var notificationIDs []uuid.UUID
	for i := 0; i < 5; i++ {
		issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
			TeamID:   f.teamID,
			Title:    fmt.Sprintf("E2E 批量已读测试 %d", i),
			StatusID: f.statusID,
		})
		require.NoError(t, err)

		_, err = f.issueService.UpdateIssue(ctx, issue.ID.String(), map[string]interface{}{
			"assignee_id": f.user2ID.String(),
		})
		require.NoError(t, err)

		// 获取通知 ID
		notifications, _, err := f.notificationService.ListNotifications(ctx, f.user2ID, 1, 10, nil, nil)
		require.NoError(t, err)
		for _, n := range notifications {
			found := false
			for _, id := range notificationIDs {
				if id == n.ID {
					found = true
					break
				}
			}
			if !found {
				notificationIDs = append(notificationIDs, n.ID)
			}
		}
	}

	require.GreaterOrEqual(t, len(notificationIDs), 5)

	// Step 2: 批量标记前 3 条已读
	batchIDs := notificationIDs[:3]
	marked, err := f.notificationService.MarkBatchAsRead(ctx, batchIDs, f.user2ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), marked)

	// Step 3: 验证未读数量
	unreadCount, err := f.notificationService.GetUnreadCount(ctx, f.user2ID)
	require.NoError(t, err)
	assert.LessOrEqual(t, unreadCount, int64(2))

	t.Log("✅ E2E 测试通过：批量标记已读功能正常工作")
}

// TestE2E_CommentAutoSubscribe 验证评论者自动订阅功能
func TestE2E_CommentAutoSubscribe(t *testing.T) {
	tx := testSvcDB.Begin()
	defer tx.Rollback()

	f := setupE2ETestFixtures(t, tx)
	ctx := f.ctx

	// Step 1: 创建 Issue（由 user1 创建，自动订阅）
	issue, err := f.issueService.CreateIssue(ctx, &CreateIssueParams{
		TeamID:   f.teamID,
		Title:    "E2E 自动订阅测试",
		StatusID: f.statusID,
	})
	require.NoError(t, err)

	// Step 2: user2 评论（自动订阅）
	_, err = f.commentService.CreateComment(ctx, issue.ID, f.user2ID, "我发表了一条评论", nil)
	require.NoError(t, err)

	// Step 3: 验证 user2 已订阅
	subscribed, err := f.subscriptionStore.IsSubscribed(ctx, issue.ID, f.user2ID)
	require.NoError(t, err)
	assert.True(t, subscribed)

	// Step 4: user1 更新状态，user2 应收到通知
	_, err = f.issueService.UpdateIssue(ctx, issue.ID.String(), map[string]interface{}{
		"status_id": f.status2ID.String(),
	})
	require.NoError(t, err)

	// Step 5: 验证 user2 收到状态变更通知
	notifications, total, err := f.notificationService.ListNotifications(ctx, f.user2ID, 1, 10, nil, []model.NotificationType{model.NotificationTypeIssueStatusChanged})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	require.GreaterOrEqual(t, len(notifications), 1)

	t.Log("✅ E2E 测试通过：评论者自动订阅功能正常工作")
}
