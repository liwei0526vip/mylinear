package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TestActivityStore_Interface 测试 ActivityStore 接口定义存在
func TestActivityStore_Interface(t *testing.T) {
	var _ ActivityStore = (*activityStore)(nil)
}

// =============================================================================
// CreateActivity 测试
// =============================================================================

func TestActivityStore_CreateActivity(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewActivityStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupActivityTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	tests := []struct {
		name    string
		activity *model.Activity
		wantErr  bool
	}{
		{
			name: "正常创建 issue_created 活动",
			activity: &model.Activity{
				IssueID: issue1.ID,
				Type:    model.ActivityIssueCreated,
				ActorID: user1.ID,
				Payload: nil,
			},
			wantErr: false,
		},
		{
			name: "正常创建 title_changed 活动（带 Payload）",
			activity: &model.Activity{
				IssueID: issue1.ID,
				Type:    model.ActivityTitleChanged,
				ActorID: user1.ID,
				Payload: mustMarshalJSON(model.ActivityPayloadTitle{
					OldValue: "旧标题",
					NewValue: "新标题",
				}),
			},
			wantErr: false,
		},
		{
			name: "正常创建 status_changed 活动（带复杂 Payload）",
			activity: &model.Activity{
				IssueID: issue1.ID,
				Type:    model.ActivityStatusChanged,
				ActorID: user1.ID,
				Payload: mustMarshalJSON(model.ActivityPayloadStatus{
					OldStatus: &model.ActivityStatusRef{
						ID:    fixtures.status.ID,
						Name:  "Backlog",
						Color: "#808080",
					},
					NewStatus: &model.ActivityStatusRef{
						ID:    fixtures.status.ID,
						Name:  "Todo",
						Color: "#6366f1",
					},
				}),
			},
			wantErr: false,
		},
		{
			name: "正常创建 comment_added 活动",
			activity: &model.Activity{
				IssueID: issue1.ID,
				Type:    model.ActivityCommentAdded,
				ActorID: user1.ID,
				Payload: mustMarshalJSON(model.ActivityPayloadComment{
					CommentID:      uuid.New(),
					CommentPreview: "这是一条评论的预览内容...",
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateActivity(ctx, tt.activity)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateActivity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// 验证 ID 已生成
				if tt.activity.ID == uuid.Nil {
					t.Error("CreateActivity() 未生成 ID")
				}
				// 验证 CreatedAt 已设置
				if tt.activity.CreatedAt.IsZero() {
					t.Error("CreateActivity() 未设置 CreatedAt")
				}
			}
		})
	}
}

// =============================================================================
// GetActivityByID 测试
// =============================================================================

func TestActivityStore_GetActivityByID(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewActivityStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupActivityTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	// 创建一个活动
	activity := &model.Activity{
		IssueID: issue1.ID,
		Type:    model.ActivityTitleChanged,
		ActorID: user1.ID,
		Payload: mustMarshalJSON(model.ActivityPayloadTitle{
			OldValue: "旧标题",
			NewValue: "新标题",
		}),
	}
	if err := store.CreateActivity(ctx, activity); err != nil {
		t.Fatalf("创建测试活动失败: %v", err)
	}

	tests := []struct {
		name       string
		id         uuid.UUID
		wantErr    bool
		errContain string
	}{
		{
			name:    "正常获取",
			id:      activity.ID,
			wantErr: false,
		},
		{
			name:       "不存在的 ID",
			id:         uuid.New(),
			wantErr:    true,
			errContain: "未找到",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetActivityByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActivityByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != activity.ID {
					t.Errorf("GetActivityByID() ID = %v, want %v", got.ID, activity.ID)
				}
				if got.Type != activity.Type {
					t.Errorf("GetActivityByID() Type = %v, want %v", got.Type, activity.Type)
				}
			}
		})
	}
}

// =============================================================================
// GetActivitiesByIssueID 测试
// =============================================================================

func TestActivityStore_GetActivitiesByIssueID(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewActivityStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupActivityTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	issue2 := fixtures.issues[1]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]

	// 为 issue1 创建多个活动
	activities := []*model.Activity{
		{
			IssueID: issue1.ID,
			Type:    model.ActivityIssueCreated,
			ActorID: user1.ID,
		},
		{
			IssueID: issue1.ID,
			Type:    model.ActivityTitleChanged,
			ActorID: user1.ID,
			Payload: mustMarshalJSON(model.ActivityPayloadTitle{
				OldValue: "旧标题",
				NewValue: "新标题",
			}),
		},
		{
			IssueID: issue1.ID,
			Type:    model.ActivityStatusChanged,
			ActorID: user2.ID,
		},
		{
			IssueID: issue1.ID,
			Type:    model.ActivityCommentAdded,
			ActorID: user1.ID,
		},
	}
	for _, a := range activities {
		if err := store.CreateActivity(ctx, a); err != nil {
			t.Fatalf("创建测试活动失败: %v", err)
		}
	}

	tests := []struct {
		name      string
		issueID   uuid.UUID
		opts      *ListActivitiesOptions
		wantCount int
		wantErr   bool
	}{
		{
			name:      "获取 Issue 的所有活动",
			issueID:   issue1.ID,
			opts:      nil,
			wantCount: 4,
			wantErr:   false,
		},
		{
			name:      "没有活动的 Issue",
			issueID:   issue2.ID,
			opts:      nil,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:    "按类型过滤 - 单个类型",
			issueID: issue1.ID,
			opts: &ListActivitiesOptions{
				Types: []model.ActivityType{model.ActivityStatusChanged},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:    "按类型过滤 - 多个类型",
			issueID: issue1.ID,
			opts: &ListActivitiesOptions{
				Types: []model.ActivityType{model.ActivityTitleChanged, model.ActivityCommentAdded},
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:    "分页 - 第一页",
			issueID: issue1.ID,
			opts: &ListActivitiesOptions{
				Page:     1,
				PageSize: 2,
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:    "分页 - 第二页",
			issueID: issue1.ID,
			opts: &ListActivitiesOptions{
				Page:     2,
				PageSize: 2,
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetActivitiesByIssueID(ctx, tt.issueID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActivitiesByIssueID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetActivitiesByIssueID() got %d activities, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// =============================================================================
// GetActivitiesByIssueIDWithTotal 测试
// =============================================================================

func TestActivityStore_GetActivitiesByIssueIDWithTotal(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewActivityStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupActivityTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	// 创建 5 个活动
	for i := 0; i < 5; i++ {
		activity := &model.Activity{
			IssueID: issue1.ID,
			Type:    model.ActivityCommentAdded,
			ActorID: user1.ID,
		}
		if err := store.CreateActivity(ctx, activity); err != nil {
			t.Fatalf("创建测试活动失败: %v", err)
		}
	}

	tests := []struct {
		name         string
		issueID      uuid.UUID
		opts         *ListActivitiesOptions
		wantCount    int
		wantTotal    int64
		wantErr      bool
	}{
		{
			name:      "分页查询带总数",
			issueID:   issue1.ID,
			opts:      &ListActivitiesOptions{Page: 1, PageSize: 3},
			wantCount: 3,
			wantTotal: 5,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := store.GetActivitiesByIssueIDWithTotal(ctx, tt.issueID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActivitiesByIssueIDWithTotal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetActivitiesByIssueIDWithTotal() got %d activities, want %d", len(got), tt.wantCount)
			}
			if total != tt.wantTotal {
				t.Errorf("GetActivitiesByIssueIDWithTotal() got total %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

// =============================================================================
// 测试辅助函数
// =============================================================================

type activityTestFixtures struct {
	users  []*model.User
	issues []*model.Issue
	team   *model.Team
	status *model.WorkflowState
}

func setupActivityTestFixtures(t *testing.T, db *gorm.DB) *activityTestFixtures {
	ctx := context.Background()
	userStore := NewUserStore(db)
	teamStore := NewTeamStore(db)
	issueStore := NewIssueStore(db)

	// 创建前缀避免冲突
	prefix := uuid.New().String()[:8]

	// 创建用户
	users := make([]*model.User, 2)
	for i := 0; i < 2; i++ {
		user := &model.User{
			WorkspaceID:  testWorkspaceID,
			Email:        prefix + "_activity_user" + string(rune('1'+i)) + "@example.com",
			Username:     prefix + "_actuser" + string(rune('1'+i)),
			Name:         "Test Activity User " + string(rune('1'+i)),
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
		Name:        prefix + "_Activity Team",
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
			Title:       prefix + "_Activity Issue " + string(rune('1'+i)),
			StatusID:    status.ID,
			Priority:    0,
			CreatedByID: users[0].ID,
		}
		if err := issueStore.Create(ctx, issue); err != nil {
			t.Fatalf("创建 Issue 失败: %v", err)
		}
		issues[i] = issue
	}

	return &activityTestFixtures{
		users:  users,
		issues: issues,
		team:   team,
		status: status,
	}
}

// mustMarshalJSON 辅助函数
func mustMarshalJSON(v interface{}) datatypes.JSON {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return datatypes.JSON(data)
}
