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

// notificationTestFixtures 测试 fixtures
type notificationTestFixtures struct {
	service     NotificationService
	db          *gorm.DB
	workspaceID uuid.UUID
}

// setupNotificationTest 初始化通知服务测试环境
func setupNotificationTest(t *testing.T) *notificationTestFixtures {
	require.NotNil(t, testSvcDB, "testSvcDB 未初始化")

	// 清理通知相关表
	require.NoError(t, testSvcDB.Exec("DELETE FROM notifications").Error)
	require.NoError(t, testSvcDB.Exec("DELETE FROM notification_preferences").Error)

	// 创建测试 workspace（使用唯一 slug）
	prefix := uuid.New().String()[:8]
	workspace := &model.Workspace{
		Name: prefix + "_notification_ws",
		Slug: prefix + "-notif-ws",
	}
	require.NoError(t, testSvcDB.Create(workspace).Error)

	notificationStore := store.NewNotificationStore(testSvcDB)
	preferenceStore := store.NewNotificationPreferenceStore(testSvcDB)
	userStore := store.NewUserStore(testSvcDB)

	service := NewNotificationService(notificationStore, preferenceStore, userStore)
	return &notificationTestFixtures{
		service:     service,
		db:          testSvcDB,
		workspaceID: workspace.ID,
	}
}

// createTestUserForNotification 创建测试用户
func (f *notificationTestFixtures) createTestUser(t *testing.T, username string) *model.User {
	// 使用前缀确保用户名唯一
	prefix := f.workspaceID.String()[:8]
	user := &model.User{
		WorkspaceID:  f.workspaceID,
		Email:        fmt.Sprintf("%s_%s@test.com", prefix, username),
		Name:         username,
		Username:     fmt.Sprintf("%s_%s", prefix, username),
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	require.NoError(t, f.db.Create(user).Error)
	return user
}

// ptrString 辅助函数，返回字符串指针
func ptrString(s string) *string {
	return &s
}

// TestNotificationService_CreateNotification 测试创建通知
func TestNotificationService_CreateNotification(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	user := f.createTestUser(t, "testuser")
	issueID := uuid.New()

	tests := []struct {
		name        string
		notification *model.Notification
		wantErr     bool
		errContains string
	}{
		{
			name: "Issue 指派通知",
			notification: &model.Notification{
				UserID:       user.ID,
				Type:         model.NotificationTypeIssueAssigned,
				Title:        "您被分配了一个 Issue",
				Body:         ptrString("Issue #123 已分配给您"),
				ResourceType: "issue",
				ResourceID:   &issueID,
			},
			wantErr: false,
		},
		{
			name: "@mention 通知",
			notification: &model.Notification{
				UserID:       user.ID,
				Type:         model.NotificationTypeIssueMentioned,
				Title:        "您在 Issue 中被提及",
				Body:         ptrString("@testuser 请查看这个问题"),
				ResourceType: "issue",
				ResourceID:   &issueID,
			},
			wantErr: false,
		},
		{
			name: "订阅变更通知",
			notification: &model.Notification{
				UserID:       user.ID,
				Type:         model.NotificationTypeIssueStatusChanged,
				Title:        "订阅的 Issue 状态已更新",
				Body:         ptrString("Issue 状态已变更为已完成"),
				ResourceType: "issue",
				ResourceID:   &issueID,
			},
			wantErr: false,
		},
		{
			name: "无效的通知类型",
			notification: &model.Notification{
				UserID: user.ID,
				Type:   model.NotificationType("invalid_type"),
				Title:  "测试标题",
			},
			wantErr:     true,
			errContains: "无效的通知类型",
		},
		{
			name: "标题不能为空",
			notification: &model.Notification{
				UserID: user.ID,
				Type:   model.NotificationTypeIssueAssigned,
				Title:  "",
			},
			wantErr:     true,
			errContains: "标题不能为空",
		},
		{
			name: "用户 ID 不能为空",
			notification: &model.Notification{
				UserID: uuid.Nil,
				Type:   model.NotificationTypeIssueAssigned,
				Title:  "测试标题",
			},
			wantErr:     true,
			errContains: "用户 ID 不能为空",
		},
		{
			name: "nil 通知返回错误",
			notification: nil,
			wantErr:      true,
			errContains:  "通知不能为 nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.service.CreateNotification(ctx, tt.notification)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// 验证通知已创建
			assert.NotEqual(t, uuid.Nil, tt.notification.ID)
			assert.False(t, tt.notification.CreatedAt.IsZero())

			// 验证数据库中的记录
			var found model.Notification
			err = f.db.First(&found, "id = ?", tt.notification.ID).Error
			require.NoError(t, err)
			assert.Equal(t, tt.notification.UserID, found.UserID)
			assert.Equal(t, tt.notification.Type, found.Type)
			assert.Equal(t, tt.notification.Title, found.Title)
		})
	}
}

// TestNotificationService_NotifyIssueAssigned 测试 Issue 指派通知
func TestNotificationService_NotifyIssueAssigned(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	actor := f.createTestUser(t, "actor")
	assignee := f.createTestUser(t, "assignee")
	otherUser := f.createTestUser(t, "other")
	issueID := uuid.New()

	tests := []struct {
		name            string
		actorID         uuid.UUID
		assigneeID      *uuid.UUID
		issueTitle      string
		wantNotification bool
		wantErr         bool
	}{
		{
			name:            "正常指派",
			actorID:         actor.ID,
			assigneeID:      &assignee.ID,
			issueTitle:      "测试 Issue",
			wantNotification: true,
			wantErr:         false,
		},
		{
			name:            "指派给自己不通知",
			actorID:         assignee.ID,
			assigneeID:      &assignee.ID,
			issueTitle:      "测试 Issue",
			wantNotification: false,
			wantErr:         false,
		},
		{
			name:            "取消指派不通知",
			actorID:         actor.ID,
			assigneeID:      nil,
			issueTitle:      "测试 Issue",
			wantNotification: false,
			wantErr:         false,
		},
		{
			name:            "指派给其他用户",
			actorID:         actor.ID,
			assigneeID:      &otherUser.ID,
			issueTitle:      "测试 Issue",
			wantNotification: true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理
			f.db.Exec("DELETE FROM notifications")

			err := f.service.NotifyIssueAssigned(ctx, tt.actorID, tt.assigneeID, issueID, tt.issueTitle)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// 验证通知是否生成
			var count int64
			f.db.Model(&model.Notification{}).Where("type = ?", model.NotificationTypeIssueAssigned).Count(&count)

			if tt.wantNotification {
				assert.Equal(t, int64(1), count)

				// 验证通知内容
				var notification model.Notification
				err := f.db.Where("type = ?", model.NotificationTypeIssueAssigned).First(&notification).Error
				require.NoError(t, err)
				assert.Equal(t, *tt.assigneeID, notification.UserID)
				assert.Contains(t, notification.Title, "分配")
				assert.Equal(t, "issue", notification.ResourceType)
				assert.Equal(t, issueID, *notification.ResourceID)
			} else {
				assert.Equal(t, int64(0), count)
			}
		})
	}
}

// TestNotificationService_NotifyIssueMentioned 测试 @mention 通知
func TestNotificationService_NotifyIssueMentioned(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	actor := f.createTestUser(t, "actor")
	mentioned1 := f.createTestUser(t, "mentioned1")
	mentioned2 := f.createTestUser(t, "mentioned2")
	invalidUser := f.createTestUser(t, "invalid") // 会删除，模拟不存在
	f.db.Delete(invalidUser) // 删除用户，模拟无效用户名

	issueID := uuid.New()

	tests := []struct {
		name              string
		actorID           uuid.UUID
		mentionedUsernames []string
		issueTitle        string
		wantCount         int
		wantErr           bool
	}{
		{
			name:              "单个 mention",
			actorID:           actor.ID,
			mentionedUsernames: []string{mentioned1.Username},
			issueTitle:        "测试 Issue",
			wantCount:         1,
			wantErr:           false,
		},
		{
			name:              "多个 mention",
			actorID:           actor.ID,
			mentionedUsernames: []string{mentioned1.Username, mentioned2.Username},
			issueTitle:        "测试 Issue",
			wantCount:         2,
			wantErr:           false,
		},
		{
			name:              "自己 mention 自己不通知",
			actorID:           mentioned1.ID,
			mentionedUsernames: []string{mentioned1.Username},
			issueTitle:        "测试 Issue",
			wantCount:         0,
			wantErr:           false,
		},
		{
			name:              "无效 username 忽略",
			actorID:           actor.ID,
			mentionedUsernames: []string{"nonexistent"},
			issueTitle:        "测试 Issue",
			wantCount:         0,
			wantErr:           false,
		},
		{
			name:              "空 mention 列表",
			actorID:           actor.ID,
			mentionedUsernames: []string{},
			issueTitle:        "测试 Issue",
			wantCount:         0,
			wantErr:           false,
		},
		{
			name:              "混合有效和无效",
			actorID:           actor.ID,
			mentionedUsernames: []string{mentioned1.Username, "nonexistent", mentioned2.Username},
			issueTitle:        "测试 Issue",
			wantCount:         2,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理
			f.db.Exec("DELETE FROM notifications")

			err := f.service.NotifyIssueMentioned(ctx, tt.actorID, tt.mentionedUsernames, issueID, tt.issueTitle)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// 验证通知数量
			var count int64
			f.db.Model(&model.Notification{}).Where("type = ?", model.NotificationTypeIssueMentioned).Count(&count)
			assert.Equal(t, int64(tt.wantCount), count)
		})
	}
}

// TestNotificationService_NotifySubscribers 测试订阅者通知
func TestNotificationService_NotifySubscribers(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	actor := f.createTestUser(t, "actor")
	subscriber1 := f.createTestUser(t, "subscriber1")
	subscriber2 := f.createTestUser(t, "subscriber2")

	issueID := uuid.New()

	tests := []struct {
		name         string
		actorID      uuid.UUID
		subscribers  []uuid.UUID
		notifyType   model.NotificationType
		title        string
		body         string
		wantCount    int
		wantErr      bool
	}{
		{
			name:        "状态变更通知订阅者",
			actorID:     actor.ID,
			subscribers: []uuid.UUID{subscriber1.ID, subscriber2.ID},
			notifyType:  model.NotificationTypeIssueStatusChanged,
			title:       "Issue 状态已更新",
			body:        "状态变更为已完成",
			wantCount:   2,
			wantErr:     false,
		},
		{
			name:        "新评论通知订阅者",
			actorID:     actor.ID,
			subscribers: []uuid.UUID{subscriber1.ID, subscriber2.ID},
			notifyType:  model.NotificationTypeIssueCommented,
			title:       "Issue 有新评论",
			body:        "这是一条新评论",
			wantCount:   2,
			wantErr:     false,
		},
		{
			name:        "优先级变更通知订阅者",
			actorID:     actor.ID,
			subscribers: []uuid.UUID{subscriber1.ID},
			notifyType:  model.NotificationTypeIssuePriorityChanged,
			title:       "Issue 优先级已更新",
			body:        "优先级变更为高",
			wantCount:   1,
			wantErr:     false,
		},
		{
			name:        "排除操作者本人",
			actorID:     subscriber1.ID,
			subscribers: []uuid.UUID{subscriber1.ID, subscriber2.ID},
			notifyType:  model.NotificationTypeIssueStatusChanged,
			title:       "Issue 状态已更新",
			body:        "状态变更为已完成",
			wantCount:   1, // 只有 subscriber2 收到通知
			wantErr:     false,
		},
		{
			name:        "空订阅者列表",
			actorID:     actor.ID,
			subscribers: []uuid.UUID{},
			notifyType:  model.NotificationTypeIssueStatusChanged,
			title:       "Issue 状态已更新",
			body:        "",
			wantCount:   0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理
			f.db.Exec("DELETE FROM notifications")

			err := f.service.NotifySubscribers(ctx, tt.actorID, tt.subscribers, tt.notifyType, issueID, tt.title, tt.body)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// 验证通知数量
			var count int64
			f.db.Model(&model.Notification{}).Where("type = ?", tt.notifyType).Count(&count)
			assert.Equal(t, int64(tt.wantCount), count)

			// 验证操作者没有收到通知
			if tt.wantCount > 0 {
				var actorNotificationCount int64
				f.db.Model(&model.Notification{}).
					Where("type = ? AND user_id = ?", tt.notifyType, tt.actorID).
					Count(&actorNotificationCount)
				assert.Equal(t, int64(0), actorNotificationCount)
			}
		})
	}
}

// =============================================================================
// 3.2 通知查询服务测试
// =============================================================================

// TestNotificationService_ListNotifications 测试获取通知列表
func TestNotificationService_ListNotifications(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	user1 := f.createTestUser(t, "user1")
	user2 := f.createTestUser(t, "user2")
	issueID := uuid.New()

	// 创建测试通知
	for i := 0; i < 5; i++ {
		notification := &model.Notification{
			UserID:       user1.ID,
			Type:         model.NotificationTypeIssueAssigned,
			Title:        fmt.Sprintf("通知 %d", i),
			ResourceType: "issue",
			ResourceID:   &issueID,
		}
		require.NoError(t, f.service.CreateNotification(ctx, notification))
	}

	// 创建已读通知
	readNotification := &model.Notification{
		UserID:       user1.ID,
		Type:         model.NotificationTypeIssueMentioned,
		Title:        "已读通知",
		ResourceType: "issue",
		ResourceID:   &issueID,
	}
	require.NoError(t, f.service.CreateNotification(ctx, readNotification))
	require.NoError(t, f.service.MarkAsRead(ctx, readNotification.ID, user1.ID))

	// 创建其他用户的通知
	otherNotification := &model.Notification{
		UserID:       user2.ID,
		Type:         model.NotificationTypeIssueAssigned,
		Title:        "其他用户通知",
		ResourceType: "issue",
		ResourceID:   &issueID,
	}
	require.NoError(t, f.service.CreateNotification(ctx, otherNotification))

	tests := []struct {
		name      string
		userID    uuid.UUID
		page      int
		pageSize  int
		read      *bool
		types     []model.NotificationType
		wantCount int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "获取所有通知",
			userID:    user1.ID,
			page:      1,
			pageSize:  10,
			read:      nil,
			wantCount: 6,
			wantTotal: 6,
			wantErr:   false,
		},
		{
			name:      "只获取未读通知",
			userID:    user1.ID,
			page:      1,
			pageSize:  10,
			read:      ptrBool(false),
			wantCount: 5,
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:      "只获取已读通知",
			userID:    user1.ID,
			page:      1,
			pageSize:  10,
			read:      ptrBool(true),
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "分页测试",
			userID:    user1.ID,
			page:      1,
			pageSize:  3,
			read:      nil,
			wantCount: 3,
			wantTotal: 6,
			wantErr:   false,
		},
		{
			name:      "第二页测试",
			userID:    user1.ID,
			page:      2,
			pageSize:  3,
			read:      nil,
			wantCount: 3,
			wantTotal: 6,
			wantErr:   false,
		},
		{
			name:   "按类型过滤",
			userID: user1.ID,
			page:   1,
			pageSize: 10,
			types:  []model.NotificationType{model.NotificationTypeIssueMentioned},
			wantCount: 1,
			wantTotal: 1,
			wantErr: false,
		},
		{
			name:      "其他用户有 1 条未读通知",
			userID:    user2.ID,
			page:      1,
			pageSize:  10,
			read:      ptrBool(false),
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifications, total, err := f.service.ListNotifications(ctx, tt.userID, tt.page, tt.pageSize, tt.read, tt.types)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(notifications))
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

// TestNotificationService_GetUnreadCount 测试获取未读数量
func TestNotificationService_GetUnreadCount(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	user := f.createTestUser(t, "user")
	issueID := uuid.New()

	// 初始未读数量为 0
	count, err := f.service.GetUnreadCount(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 创建 3 条通知
	for i := 0; i < 3; i++ {
		notification := &model.Notification{
			UserID:       user.ID,
			Type:         model.NotificationTypeIssueAssigned,
			Title:        fmt.Sprintf("通知 %d", i),
			ResourceType: "issue",
			ResourceID:   &issueID,
		}
		require.NoError(t, f.service.CreateNotification(ctx, notification))
	}

	// 验证未读数量
	count, err = f.service.GetUnreadCount(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// 标记一条已读
	var notification model.Notification
	require.NoError(t, f.db.Where("user_id = ?", user.ID).First(&notification).Error)
	require.NoError(t, f.service.MarkAsRead(ctx, notification.ID, user.ID))

	// 验证未读数量减少
	count, err = f.service.GetUnreadCount(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// ptrBool 辅助函数，返回布尔指针
func ptrBool(b bool) *bool {
	return &b
}

// =============================================================================
// 3.3 标记已读服务测试
// =============================================================================

// TestNotificationService_MarkAsRead 测试标记单条已读
func TestNotificationService_MarkAsRead(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	user := f.createTestUser(t, "user")
	otherUser := f.createTestUser(t, "other")
	issueID := uuid.New()

	// 创建通知
	notification := &model.Notification{
		UserID:       user.ID,
		Type:         model.NotificationTypeIssueAssigned,
		Title:        "测试通知",
		ResourceType: "issue",
		ResourceID:   &issueID,
	}
	require.NoError(t, f.service.CreateNotification(ctx, notification))

	tests := []struct {
		name          string
		notificationID uuid.UUID
		userID        uuid.UUID
		wantErr       bool
		errContains   string
	}{
		{
			name:          "正常标记已读",
			notificationID: notification.ID,
			userID:        user.ID,
			wantErr:       false,
		},
		{
			name:          "标记他人通知返回错误",
			notificationID: notification.ID,
			userID:        otherUser.ID,
			wantErr:       true,
			errContains:   "无权操作",
		},
		{
			name:          "不存在通知返回错误",
			notificationID: uuid.New(),
			userID:        user.ID,
			wantErr:       true,
			errContains:   "未找到",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 对于正常标记已读，需要重新创建通知（因为前面的测试可能已标记）
			if tt.name == "正常标记已读" {
				newNotification := &model.Notification{
					UserID:       user.ID,
					Type:         model.NotificationTypeIssueAssigned,
					Title:        "测试通知",
					ResourceType: "issue",
					ResourceID:   &issueID,
				}
				require.NoError(t, f.service.CreateNotification(ctx, newNotification))
				tt.notificationID = newNotification.ID
			}

			err := f.service.MarkAsRead(ctx, tt.notificationID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// 验证已标记已读
			var found model.Notification
			require.NoError(t, f.db.First(&found, "id = ?", tt.notificationID).Error)
			assert.NotNil(t, found.ReadAt)
		})
	}
}

// TestNotificationService_MarkAllAsRead 测试标记全部已读
func TestNotificationService_MarkAllAsRead(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	user := f.createTestUser(t, "user")
	issueID := uuid.New()

	// 创建 5 条通知
	for i := 0; i < 5; i++ {
		notification := &model.Notification{
			UserID:       user.ID,
			Type:         model.NotificationTypeIssueAssigned,
			Title:        fmt.Sprintf("通知 %d", i),
			ResourceType: "issue",
			ResourceID:   &issueID,
		}
		require.NoError(t, f.service.CreateNotification(ctx, notification))
	}

	// 标记全部已读
	count, err := f.service.MarkAllAsRead(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)

	// 验证全部已读
	unreadCount, err := f.service.GetUnreadCount(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), unreadCount)

	// 再次标记应返回 0
	count, err = f.service.MarkAllAsRead(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestNotificationService_MarkBatchAsRead 测试批量标记已读
func TestNotificationService_MarkBatchAsRead(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	user := f.createTestUser(t, "user")
	otherUser := f.createTestUser(t, "other")
	issueID := uuid.New()

	// 创建 5 条通知
	var notificationIDs []uuid.UUID
	for i := 0; i < 5; i++ {
		notification := &model.Notification{
			UserID:       user.ID,
			Type:         model.NotificationTypeIssueAssigned,
			Title:        fmt.Sprintf("通知 %d", i),
			ResourceType: "issue",
			ResourceID:   &issueID,
		}
		require.NoError(t, f.service.CreateNotification(ctx, notification))
		notificationIDs = append(notificationIDs, notification.ID)
	}

	// 创建其他用户的通知（不应被标记）
	otherNotification := &model.Notification{
		UserID:       otherUser.ID,
		Type:         model.NotificationTypeIssueAssigned,
		Title:        "其他用户通知",
		ResourceType: "issue",
		ResourceID:   &issueID,
	}
	require.NoError(t, f.service.CreateNotification(ctx, otherNotification))

	tests := []struct {
		name        string
		ids         []uuid.UUID
		userID      uuid.UUID
		wantCount   int64
		wantErr     bool
	}{
		{
			name:      "批量标记部分通知",
			ids:       notificationIDs[:3],
			userID:    user.ID,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "标记剩余通知",
			ids:       notificationIDs[3:],
			userID:    user.ID,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "空 ID 列表",
			ids:       []uuid.UUID{},
			userID:    user.ID,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "部分不存在的 ID",
			ids:       []uuid.UUID{uuid.New(), uuid.New()},
			userID:    user.ID,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := f.service.MarkBatchAsRead(ctx, tt.ids, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

// =============================================================================
// 3.4 通知配置服务测试
// =============================================================================

// TestNotificationPreferenceService_GetPreferences 测试获取通知配置
func TestNotificationPreferenceService_GetPreferences(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	user := f.createTestUser(t, "user")
	preferenceService := NewNotificationPreferenceService(store.NewNotificationPreferenceStore(f.db))

	tests := []struct {
		name          string
		setupPrefs    func()
		userID        uuid.UUID
		channel       *model.NotificationChannel
		wantCount     int
		wantErr       bool
	}{
		{
			name:       "无配置返回默认值",
			setupPrefs: func() {},
			userID:     user.ID,
			channel:    nil,
			wantCount:  5, // 默认 5 种通知类型
			wantErr:    false,
		},
		{
			name: "有配置返回配置值",
			setupPrefs: func() {
				falseVal := false
				pref := &model.NotificationPreference{
					UserID:  user.ID,
					Channel: model.NotificationChannelInApp,
					Type:    model.NotificationTypeIssueAssigned,
					Enabled: &falseVal,
				}
				require.NoError(t, f.db.Create(pref).Error)
			},
			userID:    user.ID,
			channel:   nil,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "按渠道过滤",
			setupPrefs: func() {
				falseVal := false
				pref := &model.NotificationPreference{
					UserID:  user.ID,
					Channel: model.NotificationChannelInApp,
					Type:    model.NotificationTypeIssueAssigned,
					Enabled: &falseVal,
				}
				require.NoError(t, f.db.Create(pref).Error)
			},
			userID:    user.ID,
			channel:   ptrChannel(model.NotificationChannelInApp),
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理
			f.db.Exec("DELETE FROM notification_preferences")

			tt.setupPrefs()

			prefs, err := preferenceService.GetPreferences(ctx, tt.userID, tt.channel)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(prefs))
		})
	}
}

// TestNotificationPreferenceService_UpdatePreferences 测试更新通知配置
func TestNotificationPreferenceService_UpdatePreferences(t *testing.T) {
	f := setupNotificationTest(t)
	ctx := context.Background()

	user := f.createTestUser(t, "user")
	preferenceService := NewNotificationPreferenceService(store.NewNotificationPreferenceStore(f.db))

	tests := []struct {
		name       string
		setupPrefs func()
		updates    []NotificationPreferenceUpdate
		wantErr    bool
		checkFunc  func(t *testing.T)
	}{
		{
			name:       "单个更新",
			setupPrefs: func() {},
			updates: []NotificationPreferenceUpdate{
				{
					Channel: model.NotificationChannelInApp,
					Type:    model.NotificationTypeIssueAssigned,
					Enabled: false,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T) {
				prefs, err := preferenceService.GetPreferences(ctx, user.ID, ptrChannel(model.NotificationChannelInApp))
				require.NoError(t, err)
				assert.Equal(t, 1, len(prefs))
				assert.False(t, *prefs[0].Enabled)
			},
		},
		{
			name: "批量更新",
			setupPrefs: func() {
				falseVal := false
				pref := &model.NotificationPreference{
					UserID:  user.ID,
					Channel: model.NotificationChannelInApp,
					Type:    model.NotificationTypeIssueAssigned,
					Enabled: &falseVal,
				}
				require.NoError(t, f.db.Create(pref).Error)
			},
			updates: []NotificationPreferenceUpdate{
				{
					Channel: model.NotificationChannelInApp,
					Type:    model.NotificationTypeIssueAssigned,
					Enabled: true,
				},
				{
					Channel: model.NotificationChannelInApp,
					Type:    model.NotificationTypeIssueMentioned,
					Enabled: false,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T) {
				prefs, err := preferenceService.GetPreferences(ctx, user.ID, ptrChannel(model.NotificationChannelInApp))
				require.NoError(t, err)
				assert.Equal(t, 2, len(prefs))
			},
		},
		{
			name:       "MVP 仅支持 in_app 渠道",
			setupPrefs: func() {},
			updates: []NotificationPreferenceUpdate{
				{
					Channel: model.NotificationChannelEmail,
					Type:    model.NotificationTypeIssueAssigned,
					Enabled: true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理
			f.db.Exec("DELETE FROM notification_preferences")

			tt.setupPrefs()

			err := preferenceService.UpdatePreferences(ctx, user.ID, tt.updates)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}

// ptrChannel 辅助函数，返回渠道指针
func ptrChannel(c model.NotificationChannel) *model.NotificationChannel {
	return &c
}
