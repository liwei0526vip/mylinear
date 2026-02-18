package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
)

// TestNotificationStore_Interface 测试 NotificationStore 接口定义存在
func TestNotificationStore_Interface(t *testing.T) {
	var _ NotificationStore = (*notificationStore)(nil)
}

// =============================================================================
// CreateNotification 测试
// =============================================================================

func TestNotificationStore_CreateNotification(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify@example.com",
		Username:     prefix + "_notifyuser",
		Name:         "Notify Test User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	tests := []struct {
		name         string
		notification *model.Notification
		wantErr      bool
	}{
		{
			name: "正常创建 issue_assigned 通知",
			notification: &model.Notification{
				UserID:       user.ID,
				Type:         model.NotificationTypeIssueAssigned,
				Title:        "你被指派了一个 Issue",
				Body:         ptrString("测试 Issue 标题"),
				ResourceType: "issue",
			},
			wantErr: false,
		},
		{
			name: "正常创建 issue_mentioned 通知",
			notification: &model.Notification{
				UserID:       user.ID,
				Type:         model.NotificationTypeIssueMentioned,
				Title:        "你在评论中被提及",
				ResourceType: "issue",
			},
			wantErr: false,
		},
		{
			name: "nil 参数返回错误",
			notification: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateNotification(ctx, tt.notification)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// GetNotificationByID 测试
// =============================================================================

func TestNotificationStore_GetNotificationByID(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify2@example.com",
		Username:     prefix + "_notifyuser2",
		Name:         "Notify Test User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 创建测试通知
	notification := &model.Notification{
		UserID:       user.ID,
		Type:         model.NotificationTypeIssueAssigned,
		Title:        "测试通知",
		ResourceType: "issue",
	}
	if err := store.CreateNotification(ctx, notification); err != nil {
		t.Fatalf("创建测试通知失败: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "存在的通知",
			id:      notification.ID,
			wantErr: false,
		},
		{
			name:    "不存在的通知",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetNotificationByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNotificationByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != tt.id {
					t.Errorf("GetNotificationByID() got.ID = %v, want %v", got.ID, tt.id)
				}
			}
		})
	}
}

// =============================================================================
// ListNotifications 测试
// =============================================================================

func TestNotificationStore_ListNotifications(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user1 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify3@example.com",
		Username:     prefix + "_notifyuser3",
		Name:         "Notify Test User 3",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify4@example.com",
		Username:     prefix + "_notifyuser4",
		Name:         "Notify Test User 4",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user1); err != nil {
		t.Fatalf("创建测试用户1失败: %v", err)
	}
	if err := userStore.CreateUser(ctx, user2); err != nil {
		t.Fatalf("创建测试用户2失败: %v", err)
	}

	// 创建测试通知
	notifications := []*model.Notification{
		{UserID: user1.ID, Type: model.NotificationTypeIssueAssigned, Title: "通知1"},
		{UserID: user1.ID, Type: model.NotificationTypeIssueMentioned, Title: "通知2"},
		{UserID: user1.ID, Type: model.NotificationTypeIssueCommented, Title: "通知3"},
		{UserID: user2.ID, Type: model.NotificationTypeIssueAssigned, Title: "其他用户通知"},
	}
	for _, n := range notifications {
		if err := store.CreateNotification(ctx, n); err != nil {
			t.Fatalf("创建测试通知失败: %v", err)
		}
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		opts       *ListNotificationsOptions
		wantCount  int
		wantTotal  int64
	}{
		{
			name:      "获取用户1的所有通知",
			userID:    user1.ID,
			opts:      nil,
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:      "获取用户2的通知",
			userID:    user2.ID,
			opts:      nil,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "分页查询",
			userID:    user1.ID,
			opts:      &ListNotificationsOptions{Page: 1, PageSize: 2},
			wantCount: 2,
			wantTotal: 3,
		},
		{
			name:      "按类型过滤",
			userID:    user1.ID,
			opts:      &ListNotificationsOptions{Types: []model.NotificationType{model.NotificationTypeIssueAssigned}},
			wantCount: 1,
			wantTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := store.ListNotifications(ctx, tt.userID, tt.opts)
			if err != nil {
				t.Errorf("ListNotifications() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("ListNotifications() got %d notifications, want %d", len(got), tt.wantCount)
			}
			if total != tt.wantTotal {
				t.Errorf("ListNotifications() got total %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

// =============================================================================
// CountUnread 测试
// =============================================================================

func TestNotificationStore_CountUnread(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify5@example.com",
		Username:     prefix + "_notifyuser5",
		Name:         "Notify Test User 5",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 创建测试通知（全部未读）
	for i := 0; i < 3; i++ {
		if err := store.CreateNotification(ctx, &model.Notification{
			UserID: user.ID,
			Type:   model.NotificationTypeIssueAssigned,
			Title:  "测试通知",
		}); err != nil {
			t.Fatalf("创建测试通知失败: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		wantCount int64
	}{
		{
			name:      "有未读通知",
			userID:    user.ID,
			wantCount: 3,
		},
		{
			name:      "无未读通知",
			userID:    uuid.New(),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.CountUnread(ctx, tt.userID)
			if err != nil {
				t.Errorf("CountUnread() error = %v", err)
				return
			}
			if got != tt.wantCount {
				t.Errorf("CountUnread() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

// =============================================================================
// MarkAsRead 测试
// =============================================================================

func TestNotificationStore_MarkAsRead(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user1 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify6@example.com",
		Username:     prefix + "_notifyuser6",
		Name:         "Notify Test User 6",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify7@example.com",
		Username:     prefix + "_notifyuser7",
		Name:         "Notify Test User 7",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user1); err != nil {
		t.Fatalf("创建测试用户1失败: %v", err)
	}
	if err := userStore.CreateUser(ctx, user2); err != nil {
		t.Fatalf("创建测试用户2失败: %v", err)
	}

	// 创建测试通知
	notification := &model.Notification{
		UserID: user1.ID,
		Type:   model.NotificationTypeIssueAssigned,
		Title:  "测试通知",
	}
	if err := store.CreateNotification(ctx, notification); err != nil {
		t.Fatalf("创建测试通知失败: %v", err)
	}

	tests := []struct {
		name        string
		id          uuid.UUID
		userID      uuid.UUID
		wantErr     bool
	}{
		{
			name:    "正常标记已读",
			id:      notification.ID,
			userID:  user1.ID,
			wantErr: false,
		},
		{
			name:    "标记他人通知",
			id:      notification.ID,
			userID:  user2.ID,
			wantErr: true,
		},
		{
			name:    "不存在的通知",
			id:      uuid.New(),
			userID:  user1.ID,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.MarkAsRead(ctx, tt.id, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkAsRead() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// MarkAllAsRead 测试
// =============================================================================

func TestNotificationStore_MarkAllAsRead(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify8@example.com",
		Username:     prefix + "_notifyuser8",
		Name:         "Notify Test User 8",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 创建测试通知
	for i := 0; i < 3; i++ {
		if err := store.CreateNotification(ctx, &model.Notification{
			UserID: user.ID,
			Type:   model.NotificationTypeIssueAssigned,
			Title:  "测试通知",
		}); err != nil {
			t.Fatalf("创建测试通知失败: %v", err)
		}
	}

	// 标记全部已读
	count, err := store.MarkAllAsRead(ctx, user.ID)
	if err != nil {
		t.Errorf("MarkAllAsRead() error = %v", err)
	}
	if count != 3 {
		t.Errorf("MarkAllAsRead() count = %v, want 3", count)
	}

	// 验证无未读
	unread, _ := store.CountUnread(ctx, user.ID)
	if unread != 0 {
		t.Errorf("标记全部已读后应有 0 未读，got %d", unread)
	}
}

// =============================================================================
// MarkBatchAsRead 测试
// =============================================================================

func TestNotificationStore_MarkBatchAsRead(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_notify9@example.com",
		Username:     prefix + "_notifyuser9",
		Name:         "Notify Test User 9",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 创建测试通知
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		n := &model.Notification{
			UserID: user.ID,
			Type:   model.NotificationTypeIssueAssigned,
			Title:  "测试通知",
		}
		if err := store.CreateNotification(ctx, n); err != nil {
			t.Fatalf("创建测试通知失败: %v", err)
		}
		ids = append(ids, n.ID)
	}

	// 批量标记前 3 个已读
	count, err := store.MarkBatchAsRead(ctx, ids[:3], user.ID)
	if err != nil {
		t.Errorf("MarkBatchAsRead() error = %v", err)
	}
	if count != 3 {
		t.Errorf("MarkBatchAsRead() count = %v, want 3", count)
	}

	// 验证剩余 2 个未读
	unread, _ := store.CountUnread(ctx, user.ID)
	if unread != 2 {
		t.Errorf("批量标记后应有 2 未读，got %d", unread)
	}

	// 空数组
	count, _ = store.MarkBatchAsRead(ctx, []uuid.UUID{}, user.ID)
	if count != 0 {
		t.Errorf("空数组应返回 0，got %d", count)
	}
}

// =============================================================================
// 辅助函数
// =============================================================================

func ptrString(s string) *string {
	return &s
}
