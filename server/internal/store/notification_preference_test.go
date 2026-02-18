package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
)

// TestNotificationPreferenceStore_Interface 测试接口定义存在
func TestNotificationPreferenceStore_Interface(t *testing.T) {
	var _ NotificationPreferenceStore = (*notificationPreferenceStore)(nil)
}

// =============================================================================
// GetByUser 测试
// =============================================================================

func TestNotificationPreferenceStore_GetByUser(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationPreferenceStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_pref@example.com",
		Username:     prefix + "_prefuser",
		Name:         "Pref Test User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 创建测试配置
	trueVal, falseVal := true, false
	prefs := []*model.NotificationPreference{
		{UserID: user.ID, Channel: model.NotificationChannelInApp, Type: model.NotificationTypeIssueAssigned, Enabled: &trueVal},
		{UserID: user.ID, Channel: model.NotificationChannelInApp, Type: model.NotificationTypeIssueMentioned, Enabled: &falseVal},
		{UserID: user.ID, Channel: model.NotificationChannelEmail, Type: model.NotificationTypeIssueAssigned, Enabled: &trueVal},
	}
	for _, p := range prefs {
		if err := store.Upsert(ctx, p); err != nil {
			t.Fatalf("创建测试配置失败: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		channel   *model.NotificationChannel
		wantCount int
	}{
		{
			name:      "获取所有配置",
			userID:    user.ID,
			channel:   nil,
			wantCount: 3,
		},
		{
			name:      "按 in_app 渠道过滤",
			userID:    user.ID,
			channel:   ptrChannel(model.NotificationChannelInApp),
			wantCount: 2,
		},
		{
			name:      "按 email 渠道过滤",
			userID:    user.ID,
			channel:   ptrChannel(model.NotificationChannelEmail),
			wantCount: 1,
		},
		{
			name:      "无配置的用户",
			userID:    uuid.New(),
			channel:   nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByUser(ctx, tt.userID, tt.channel)
			if err != nil {
				t.Errorf("GetByUser() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetByUser() got %d preferences, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// =============================================================================
// Upsert 测试
// =============================================================================

func TestNotificationPreferenceStore_Upsert(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationPreferenceStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_pref2@example.com",
		Username:     prefix + "_prefuser2",
		Name:         "Pref Test User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 测试新建
	trueVal := true
	pref := &model.NotificationPreference{
		UserID:  user.ID,
		Channel: model.NotificationChannelInApp,
		Type:    model.NotificationTypeIssueAssigned,
		Enabled: &trueVal,
	}
	if err := store.Upsert(ctx, pref); err != nil {
		t.Errorf("Upsert() 新建 error = %v", err)
	}

	// 测试更新
	falseVal := false
	pref.Enabled = &falseVal
	if err := store.Upsert(ctx, pref); err != nil {
		t.Errorf("Upsert() 更新 error = %v", err)
	}

	// 验证更新
	got, _ := store.GetByUser(ctx, user.ID, nil)
	if len(got) != 1 {
		t.Fatalf("期望 1 条配置，got %d", len(got))
	}
	if got[0].Enabled == nil || *got[0].Enabled != false {
		t.Errorf("Enabled 应为 false, got %v", got[0].Enabled)
	}

	// 测试 nil 参数
	if err := store.Upsert(ctx, nil); err == nil {
		t.Error("Upsert(nil) 应返回错误")
	}
}

// =============================================================================
// BatchUpsert 测试
// =============================================================================

func TestNotificationPreferenceStore_BatchUpsert(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationPreferenceStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_pref3@example.com",
		Username:     prefix + "_prefuser3",
		Name:         "Pref Test User 3",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 批量创建
	trueVal, falseVal := true, false
	prefs := []model.NotificationPreference{
		{UserID: user.ID, Channel: model.NotificationChannelInApp, Type: model.NotificationTypeIssueAssigned, Enabled: &trueVal},
		{UserID: user.ID, Channel: model.NotificationChannelInApp, Type: model.NotificationTypeIssueMentioned, Enabled: &trueVal},
		{UserID: user.ID, Channel: model.NotificationChannelInApp, Type: model.NotificationTypeIssueCommented, Enabled: &falseVal},
	}
	if err := store.BatchUpsert(ctx, prefs); err != nil {
		t.Errorf("BatchUpsert() error = %v", err)
	}

	// 验证
	got, _ := store.GetByUser(ctx, user.ID, nil)
	if len(got) != 3 {
		t.Errorf("期望 3 条配置，got %d", len(got))
	}

	// 批量更新（部分启用，部分禁用）
	updatedPrefs := []model.NotificationPreference{
		{UserID: user.ID, Channel: model.NotificationChannelInApp, Type: model.NotificationTypeIssueAssigned, Enabled: &falseVal},
		{UserID: user.ID, Channel: model.NotificationChannelInApp, Type: model.NotificationTypeIssueMentioned, Enabled: &falseVal},
	}
	if err := store.BatchUpsert(ctx, updatedPrefs); err != nil {
		t.Errorf("BatchUpsert() 更新 error = %v", err)
	}

	// 空数组
	if err := store.BatchUpsert(ctx, []model.NotificationPreference{}); err != nil {
		t.Errorf("BatchUpsert() 空数组 error = %v", err)
	}
}

// =============================================================================
// IsEnabled 测试
// =============================================================================

func TestNotificationPreferenceStore_IsEnabled(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewNotificationPreferenceStore(tx)
	userStore := NewUserStore(tx)
	ctx := context.Background()

	// 创建测试用户
	prefix := uuid.New().String()[:8]
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_pref4@example.com",
		Username:     prefix + "_prefuser4",
		Name:         "Pref Test User 4",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := userStore.CreateUser(ctx, user); err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 创建禁用的配置
	falseVal := false
	pref := &model.NotificationPreference{
		UserID:  user.ID,
		Channel: model.NotificationChannelInApp,
		Type:    model.NotificationTypeIssueAssigned,
		Enabled: &falseVal,
	}
	if err := store.Upsert(ctx, pref); err != nil {
		t.Fatalf("创建测试配置失败: %v", err)
	}

	// 验证禁用的配置返回 false
	got, err := store.IsEnabled(ctx, user.ID, model.NotificationChannelInApp, model.NotificationTypeIssueAssigned)
	if err != nil {
		t.Errorf("IsEnabled() 禁用配置 error = %v", err)
	}
	if got != false {
		t.Errorf("IsEnabled() 禁用配置 = %v, want false", got)
	}

	// 验证无配置的类型返回 true（默认启用）
	got, err = store.IsEnabled(ctx, user.ID, model.NotificationChannelInApp, model.NotificationTypeIssueMentioned)
	if err != nil {
		t.Errorf("IsEnabled() 无配置 error = %v", err)
	}
	if got != true {
		t.Errorf("IsEnabled() 无配置 = %v, want true", got)
	}

	// 验证其他用户返回 true（默认启用）
	got, err = store.IsEnabled(ctx, uuid.New(), model.NotificationChannelInApp, model.NotificationTypeIssueAssigned)
	if err != nil {
		t.Errorf("IsEnabled() 其他用户 error = %v", err)
	}
	if got != true {
		t.Errorf("IsEnabled() 其他用户 = %v, want true", got)
	}
}

// =============================================================================
// 辅助函数
// =============================================================================

func ptrChannel(c model.NotificationChannel) *model.NotificationChannel {
	return &c
}
