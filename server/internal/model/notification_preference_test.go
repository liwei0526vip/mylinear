package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNotificationPreference_TableName 测试表名
func TestNotificationPreference_TableName(t *testing.T) {
	p := NotificationPreference{}
	if got := p.TableName(); got != "notification_preferences" {
		t.Errorf("NotificationPreference.TableName() = %v, want %v", got, "notification_preferences")
	}
}

// TestNotificationChannel_Valid 测试通知渠道验证
func TestNotificationChannel_Valid(t *testing.T) {
	tests := []struct {
		name    string
		channel NotificationChannel
		want    bool
	}{
		{"in_app 有效", NotificationChannelInApp, true},
		{"email 有效", NotificationChannelEmail, true},
		{"slack 有效", NotificationChannelSlack, true},
		{"无效渠道", NotificationChannel("invalid"), false},
		{"空渠道", NotificationChannel(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.channel.Valid(); got != tt.want {
				t.Errorf("NotificationChannel.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNotificationPreference_Fields 测试字段
func TestNotificationPreference_Fields(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	trueVal := true

	p := NotificationPreference{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   NotificationChannelInApp,
		Type:      NotificationTypeIssueAssigned,
		Enabled:   &trueVal,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if p.ID == uuid.Nil {
		t.Error("ID 不应为空")
	}
	if p.UserID != userID {
		t.Errorf("UserID = %v, want %v", p.UserID, userID)
	}
	if p.Channel != NotificationChannelInApp {
		t.Errorf("Channel = %v, want %v", p.Channel, NotificationChannelInApp)
	}
	if p.Type != NotificationTypeIssueAssigned {
		t.Errorf("Type = %v, want %v", p.Type, NotificationTypeIssueAssigned)
	}
	if p.Enabled == nil || !*p.Enabled {
		t.Error("Enabled 应该为 true")
	}
}

// TestNotificationPreference_BeforeCreate 测试 BeforeCreate 钩子
func TestNotificationPreference_BeforeCreate(t *testing.T) {
	t.Run("UUID 为空时自动生成", func(t *testing.T) {
		p := NotificationPreference{}
		err := p.BeforeCreate(nil)
		if err != nil {
			t.Errorf("BeforeCreate() error = %v", err)
		}
		if p.ID == uuid.Nil {
			t.Error("BeforeCreate 应该自动生成 UUID")
		}
	})

	t.Run("UUID 已存在时保持不变", func(t *testing.T) {
		existingID := uuid.New()
		p := NotificationPreference{ID: existingID}
		err := p.BeforeCreate(nil)
		if err != nil {
			t.Errorf("BeforeCreate() error = %v", err)
		}
		if p.ID != existingID {
			t.Errorf("BeforeCreate 不应该修改已存在的 UUID")
		}
	})
}

// TestNotificationPreference_BeforeUpdate 测试 BeforeUpdate 钩子
func TestNotificationPreference_BeforeUpdate(t *testing.T) {
	p := NotificationPreference{
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	oldTime := p.UpdatedAt

	err := p.BeforeUpdate(nil)
	if err != nil {
		t.Errorf("BeforeUpdate() error = %v", err)
	}

	if !p.UpdatedAt.After(oldTime) {
		t.Error("BeforeUpdate 应该更新 UpdatedAt")
	}
}

// TestDefaultPreferences 测试默认偏好配置
func TestDefaultPreferences(t *testing.T) {
	userID := uuid.New()
	preferences := DefaultPreferences(userID)

	expectedTypes := []NotificationType{
		NotificationTypeIssueAssigned,
		NotificationTypeIssueMentioned,
		NotificationTypeIssueCommented,
		NotificationTypeIssueStatusChanged,
		NotificationTypeIssuePriorityChanged,
	}

	if len(preferences) != len(expectedTypes) {
		t.Errorf("DefaultPreferences 返回 %d 个配置, want %d", len(preferences), len(expectedTypes))
	}

	for i, p := range preferences {
		if p.UserID != userID {
			t.Errorf("preferences[%d].UserID = %v, want %v", i, p.UserID, userID)
		}
		if p.Channel != NotificationChannelInApp {
			t.Errorf("preferences[%d].Channel = %v, want %v", i, p.Channel, NotificationChannelInApp)
		}
		if p.Type != expectedTypes[i] {
			t.Errorf("preferences[%d].Type = %v, want %v", i, p.Type, expectedTypes[i])
		}
		if p.Enabled == nil || !*p.Enabled {
			t.Errorf("preferences[%d].Enabled 应该为 true", i)
		}
	}
}

// TestNotificationPreference_UserRelation 测试 User 关联关系
func TestNotificationPreference_UserRelation(t *testing.T) {
	userID := uuid.New()
	user := &User{}
	user.ID = userID

	p := NotificationPreference{
		UserID: userID,
		User:   user,
	}

	if p.UserID != userID {
		t.Errorf("NotificationPreference.UserID = %v, want %v", p.UserID, userID)
	}

	if p.User == nil || p.User.ID != userID {
		t.Error("NotificationPreference.User 关联关系不正确")
	}
}
