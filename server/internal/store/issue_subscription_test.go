package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// TestIssueSubscriptionStore_Interface 测试 IssueSubscriptionStore 接口定义存在
func TestIssueSubscriptionStore_Interface(t *testing.T) {
	var _ IssueSubscriptionStore = (*issueSubscriptionStore)(nil)
}

// =============================================================================
// Subscribe 测试
// =============================================================================

func TestIssueSubscriptionStore_Subscribe(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueSubscriptionStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupSubscriptionTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]

	tests := []struct {
		name    string
		issueID uuid.UUID
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "正常订阅",
			issueID: issue1.ID,
			userID:  user1.ID,
			wantErr: false,
		},
		{
			name:    "重复订阅（幂等）",
			issueID: issue1.ID,
			userID:  user1.ID,
			wantErr: false,
		},
		{
			name:    "另一个用户订阅同一 Issue",
			issueID: issue1.ID,
			userID:  user2.ID,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Subscribe(ctx, tt.issueID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// 验证订阅数量
	var count int64
	tx.Model(&model.IssueSubscription{}).Where("issue_id = ?", issue1.ID).Count(&count)
	if count != 2 { // user1 和 user2 都订阅了
		t.Errorf("Expected 2 subscriptions, got %d", count)
	}

	// 验证可以通过 ListSubscribers 获取
	subscribers, err := store.ListSubscribers(ctx, issue1.ID)
	if err != nil {
		t.Fatalf("ListSubscribers() error = %v", err)
	}
	if len(subscribers) != 2 {
		t.Errorf("Expected 2 subscribers, got %d", len(subscribers))
	}
}

// =============================================================================
// Unsubscribe 测试
// =============================================================================

func TestIssueSubscriptionStore_Unsubscribe(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueSubscriptionStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupSubscriptionTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]

	// 先订阅
	_ = store.Subscribe(ctx, issue1.ID, user1.ID)
	_ = store.Subscribe(ctx, issue1.ID, user2.ID)

	tests := []struct {
		name    string
		issueID uuid.UUID
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "正常取消订阅",
			issueID: issue1.ID,
			userID:  user1.ID,
			wantErr: false,
		},
		{
			name:    "取消不存在的订阅（幂等）",
			issueID: issue1.ID,
			userID:  user1.ID,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Unsubscribe(ctx, tt.issueID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unsubscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// 验证只剩一个订阅者
	subscribers, err := store.ListSubscribers(ctx, issue1.ID)
	if err != nil {
		t.Fatalf("ListSubscribers() error = %v", err)
	}
	if len(subscribers) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(subscribers))
	}
	if len(subscribers) > 0 && subscribers[0].ID != user2.ID {
		t.Errorf("Expected user2 to be the remaining subscriber")
	}
}

// =============================================================================
// ListSubscribers 测试
// =============================================================================

func TestIssueSubscriptionStore_ListSubscribers(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueSubscriptionStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupSubscriptionTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	issue2 := fixtures.issues[1]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]
	user3 := fixtures.users[2]

	// 订阅多个用户
	_ = store.Subscribe(ctx, issue1.ID, user1.ID)
	_ = store.Subscribe(ctx, issue1.ID, user2.ID)
	_ = store.Subscribe(ctx, issue1.ID, user3.ID)

	tests := []struct {
		name          string
		issueID       uuid.UUID
		wantCount     int
		checkContains bool
		containsUserIDs []uuid.UUID
	}{
		{
			name:        "获取多个订阅者",
			issueID:     issue1.ID,
			wantCount:   3,
			checkContains: true,
			containsUserIDs: []uuid.UUID{user1.ID, user2.ID, user3.ID},
		},
		{
			name:       "没有订阅者的 Issue",
			issueID:    issue2.ID,
			wantCount:  0,
			checkContains: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscribers, err := store.ListSubscribers(ctx, tt.issueID)
			if err != nil {
				t.Errorf("ListSubscribers() error = %v", err)
				return
			}
			if len(subscribers) != tt.wantCount {
				t.Errorf("ListSubscribers() got %d subscribers, want %d", len(subscribers), tt.wantCount)
			}
			if tt.checkContains {
				// 验证所有用户都在订阅者列表中
				subscriberIDs := make(map[uuid.UUID]bool)
				for _, s := range subscribers {
					subscriberIDs[s.ID] = true
				}
				for _, uid := range tt.containsUserIDs {
					if !subscriberIDs[uid] {
						t.Errorf("ListSubscribers() missing user %v", uid)
					}
				}
			}
		})
	}
}

// =============================================================================
// IsSubscribed 测试
// =============================================================================

func TestIssueSubscriptionStore_IsSubscribed(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueSubscriptionStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupSubscriptionTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	issue2 := fixtures.issues[1]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]

	// user1 订阅 issue1
	_ = store.Subscribe(ctx, issue1.ID, user1.ID)

	tests := []struct {
		name        string
		issueID     uuid.UUID
		userID      uuid.UUID
		wantSubscribed bool
		wantErr     bool
	}{
		{
			name:        "已订阅",
			issueID:     issue1.ID,
			userID:      user1.ID,
			wantSubscribed: true,
			wantErr:     false,
		},
		{
			name:        "未订阅",
			issueID:     issue1.ID,
			userID:      user2.ID,
			wantSubscribed: false,
			wantErr:     false,
		},
		{
			name:        "另一个 Issue 未订阅",
			issueID:     issue2.ID,
			userID:      user1.ID,
			wantSubscribed: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscribed, err := store.IsSubscribed(ctx, tt.issueID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSubscribed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if subscribed != tt.wantSubscribed {
				t.Errorf("IsSubscribed() = %v, want %v", subscribed, tt.wantSubscribed)
			}
		})
	}
}

// =============================================================================
// 测试辅助函数
// =============================================================================

type subscriptionTestFixtures struct {
	users   []*model.User
	issues  []*model.Issue
	team    *model.Team
	status  *model.WorkflowState
}

func setupSubscriptionTestFixtures(t *testing.T, db *gorm.DB) *subscriptionTestFixtures {
	ctx := context.Background()
	userStore := NewUserStore(db)
	teamStore := NewTeamStore(db)
	issueStore := NewIssueStore(db)

	// 创建前缀避免冲突
	prefix := uuid.New().String()[:8]

	// 创建用户
	users := make([]*model.User, 3)
	for i := 0; i < 3; i++ {
		user := &model.User{
			WorkspaceID:  testWorkspaceID,
			Email:        prefix + "_user" + string(rune('1'+i)) + "@example.com",
			Username:     prefix + "_user" + string(rune('1'+i)),
			Name:         "Test User " + string(rune('1'+i)),
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
		WorkspaceID: testWorkspaceID,
		Name:        prefix + "_Team",
		Key:         "TSK", // 团队 Key 必须以大写字母开头
	}
	if err := teamStore.Create(ctx, team); err != nil {
		t.Fatalf("创建团队失败: %v", err)
	}

	// 创建工作流状态
	status := &model.WorkflowState{
		TeamID:    team.ID,
		Name:      "Backlog",
		Type:      model.StateTypeBacklog,
		Color:     "#808080",
		Position:  0,
	}
	if err := db.Create(status).Error; err != nil {
		t.Fatalf("创建工作流状态失败: %v", err)
	}

	// 创建 Issue（使用 IssueStore.Create 来自动生成 Number）
	issues := make([]*model.Issue, 2)
	for i := 0; i < 2; i++ {
		issue := &model.Issue{
			TeamID:      team.ID,
			Title:       prefix + "_Issue " + string(rune('1'+i)),
			StatusID:    status.ID,
			Priority:    0,
			CreatedByID: users[0].ID,
		}
		if err := issueStore.Create(ctx, issue); err != nil {
			t.Fatalf("创建 Issue 失败: %v", err)
		}
		issues[i] = issue
	}

	return &subscriptionTestFixtures{
		users:  users,
		issues: issues,
		team:   team,
		status: status,
	}
}
