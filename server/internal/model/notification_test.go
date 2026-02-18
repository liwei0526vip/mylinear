package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNotification_TableName 测试 Notification 表名
func TestNotification_TableName(t *testing.T) {
	n := Notification{}
	if got := n.TableName(); got != "notifications" {
		t.Errorf("Notification.TableName() = %v, want %v", got, "notifications")
	}
}

// TestNotification_Fields 测试 Notification 字段
func TestNotification_Fields(t *testing.T) {
	userID := uuid.New()
	resourceID := uuid.New()
	now := time.Now()
	body := "测试通知内容"

	tests := []struct {
		name         string
		notification Notification
		checkFunc    func(n Notification) bool
	}{
		{
			name: "所有字段都有值",
			notification: Notification{
				ID:           uuid.New(),
				UserID:       userID,
				Type:         NotificationTypeIssueAssigned,
				Title:        "你被指派了一个 Issue",
				Body:         &body,
				ResourceType: "issue",
				ResourceID:   &resourceID,
				ReadAt:       &now,
				CreatedAt:    now,
			},
			checkFunc: func(n Notification) bool {
				return n.ID != uuid.Nil &&
					n.UserID == userID &&
					n.Type == NotificationTypeIssueAssigned &&
					n.Title == "你被指派了一个 Issue" &&
					n.Body != nil && *n.Body == "测试通知内容" &&
					n.ResourceType == "issue" &&
					n.ResourceID != nil && *n.ResourceID == resourceID &&
					n.ReadAt != nil &&
					!n.CreatedAt.IsZero()
			},
		},
		{
			name: "可选字段为空",
			notification: Notification{
				ID:           uuid.New(),
				UserID:       userID,
				Type:         NotificationTypeIssueMentioned,
				Title:        "你在评论中被提及",
				ResourceType: "issue",
				CreatedAt:    now,
			},
			checkFunc: func(n Notification) bool {
				return n.Body == nil &&
					n.ResourceID == nil &&
					n.ReadAt == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.checkFunc(tt.notification) {
				t.Errorf("Notification 字段检查失败: %+v", tt.notification)
			}
		})
	}
}

// TestNotification_IsRead 测试 IsRead 方法
func TestNotification_IsRead(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		notify   Notification
		wantRead bool
	}{
		{
			name:     "未读通知",
			notify:   Notification{ReadAt: nil},
			wantRead: false,
		},
		{
			name:     "已读通知",
			notify:   Notification{ReadAt: &now},
			wantRead: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.notify.IsRead(); got != tt.wantRead {
				t.Errorf("Notification.IsRead() = %v, want %v", got, tt.wantRead)
			}
		})
	}
}

// TestNotification_IsUnread 测试 IsUnread 方法
func TestNotification_IsUnread(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		notify     Notification
		wantUnread bool
	}{
		{
			name:       "未读通知",
			notify:     Notification{ReadAt: nil},
			wantUnread: true,
		},
		{
			name:       "已读通知",
			notify:     Notification{ReadAt: &now},
			wantUnread: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.notify.IsUnread(); got != tt.wantUnread {
				t.Errorf("Notification.IsUnread() = %v, want %v", got, tt.wantUnread)
			}
		})
	}
}

// TestNotification_MarkAsRead 测试 MarkAsRead 方法
func TestNotification_MarkAsRead(t *testing.T) {
	notify := Notification{ReadAt: nil}

	if !notify.IsUnread() {
		t.Error("初始状态应该是未读")
	}

	notify.MarkAsRead()

	if notify.IsUnread() {
		t.Error("MarkAsRead 后应该是已读")
	}

	if notify.ReadAt == nil {
		t.Error("ReadAt 不应该为 nil")
	}
}

// TestNotification_HasResource 测试 HasResource 方法
func TestNotification_HasResource(t *testing.T) {
	resourceID := uuid.New()

	tests := []struct {
		name         string
		notify       Notification
		wantResource bool
	}{
		{
			name:         "无关联资源",
			notify:       Notification{ResourceType: "", ResourceID: nil},
			wantResource: false,
		},
		{
			name:         "只有 ResourceType",
			notify:       Notification{ResourceType: "issue", ResourceID: nil},
			wantResource: false,
		},
		{
			name:         "只有 ResourceID",
			notify:       Notification{ResourceType: "", ResourceID: &resourceID},
			wantResource: false,
		},
		{
			name:         "有关联资源",
			notify:       Notification{ResourceType: "issue", ResourceID: &resourceID},
			wantResource: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.notify.HasResource(); got != tt.wantResource {
				t.Errorf("Notification.HasResource() = %v, want %v", got, tt.wantResource)
			}
		})
	}
}

// TestNotification_BeforeCreate 测试 BeforeCreate 钩子
func TestNotification_BeforeCreate(t *testing.T) {
	t.Run("UUID 为空时自动生成", func(t *testing.T) {
		notify := Notification{}
		err := notify.BeforeCreate(nil)
		if err != nil {
			t.Errorf("BeforeCreate() error = %v", err)
		}
		if notify.ID == uuid.Nil {
			t.Error("BeforeCreate 应该自动生成 UUID")
		}
	})

	t.Run("UUID 已存在时保持不变", func(t *testing.T) {
		existingID := uuid.New()
		notify := Notification{ID: existingID}
		err := notify.BeforeCreate(nil)
		if err != nil {
			t.Errorf("BeforeCreate() error = %v", err)
		}
		if notify.ID != existingID {
			t.Errorf("BeforeCreate 不应该修改已存在的 UUID")
		}
	})
}

// TestNotification_UserRelation 测试 User 关联关系
func TestNotification_UserRelation(t *testing.T) {
	userID := uuid.New()
	user := &User{}
	user.ID = userID

	notify := Notification{
		UserID: userID,
		User:   user,
	}

	if notify.UserID != userID {
		t.Errorf("Notification.UserID = %v, want %v", notify.UserID, userID)
	}

	if notify.User == nil || notify.User.ID != userID {
		t.Error("Notification.User 关联关系不正确")
	}
}
