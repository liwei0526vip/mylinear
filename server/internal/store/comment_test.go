package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// TestCommentStore_Interface 测试 CommentStore 接口定义存在
func TestCommentStore_Interface(t *testing.T) {
	var _ CommentStore = (*commentStore)(nil)
}

// =============================================================================
// CreateComment 测试
// =============================================================================

func TestCommentStore_CreateComment(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewCommentStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	tests := []struct {
		name    string
		comment *model.Comment
		wantErr bool
	}{
		{
			name: "正常创建评论",
			comment: &model.Comment{
				IssueID: issue1.ID,
				UserID:  user1.ID,
				Body:    "这是一条测试评论",
			},
			wantErr: false,
		},
		{
			name: "创建嵌套回复",
			comment: &model.Comment{
				IssueID: issue1.ID,
				ParentID: &fixtures.parentCommentID,
				UserID:  user1.ID,
				Body:    "这是一条回复",
			},
			wantErr: false,
		},
	}

	// 先创建一个父评论用于嵌套回复测试
	parentComment := &model.Comment{
		IssueID: issue1.ID,
		UserID:  user1.ID,
		Body:    "这是父评论",
	}
	if err := store.CreateComment(ctx, parentComment); err != nil {
		t.Fatalf("创建父评论失败: %v", err)
	}
	fixtures.parentCommentID = parentComment.ID

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateComment(ctx, tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.comment.ID == uuid.Nil {
					t.Error("CreateComment() 未生成 ID")
				}
				if tt.comment.CreatedAt.IsZero() {
					t.Error("CreateComment() 未设置 CreatedAt")
				}
			}
		})
	}
}

// =============================================================================
// GetCommentByID 测试
// =============================================================================

func TestCommentStore_GetCommentByID(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewCommentStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	// 创建测试评论
	comment := &model.Comment{
		IssueID: issue1.ID,
		UserID:  user1.ID,
		Body:    "测试评论",
	}
	if err := store.CreateComment(ctx, comment); err != nil {
		t.Fatalf("创建测试评论失败: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "正常获取",
			id:      comment.ID,
			wantErr: false,
		},
		{
			name:    "不存在的 ID",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetCommentByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommentByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != comment.ID {
					t.Errorf("GetCommentByID() ID = %v, want %v", got.ID, comment.ID)
				}
			}
		})
	}
}

// =============================================================================
// GetCommentsByIssueID 测试
// =============================================================================

func TestCommentStore_GetCommentsByIssueID(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewCommentStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	issue2 := fixtures.issues[1]
	user1 := fixtures.users[0]
	user2 := fixtures.users[1]

	// 创建多个评论（含嵌套回复）
	parent1 := &model.Comment{
		IssueID: issue1.ID,
		UserID:  user1.ID,
		Body:    "父评论1",
	}
	_ = store.CreateComment(ctx, parent1)

	reply1 := &model.Comment{
		IssueID: issue1.ID,
		ParentID: &parent1.ID,
		UserID:  user2.ID,
		Body:    "回复1",
	}
	_ = store.CreateComment(ctx, reply1)

	parent2 := &model.Comment{
		IssueID: issue1.ID,
		UserID:  user1.ID,
		Body:    "父评论2",
	}
	_ = store.CreateComment(ctx, parent2)

	tests := []struct {
		name      string
		issueID   uuid.UUID
		opts      *ListCommentsOptions
		wantCount int
		wantErr   bool
	}{
		{
			name:      "获取 Issue 的所有评论（树形结构）",
			issueID:   issue1.ID,
			opts:      nil,
			wantCount: 2, // 两个父评论
			wantErr:   false,
		},
		{
			name:      "没有评论的 Issue",
			issueID:   issue2.ID,
			opts:      nil,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:    "分页测试",
			issueID: issue1.ID,
			opts: &ListCommentsOptions{
				Page:     1,
				PageSize: 1,
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetCommentsByIssueID(ctx, tt.issueID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommentsByIssueID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetCommentsByIssueID() got %d comments, want %d", len(got), tt.wantCount)
			}
			// 验证树形结构：第一个父评论应该有 1 个回复
			if tt.name == "获取 Issue 的所有评论（树形结构）" && len(got) > 0 {
				if len(got[0].Replies) != 1 {
					t.Errorf("GetCommentsByIssueID() 第一个父评论应有 1 个回复，实际有 %d 个", len(got[0].Replies))
				}
			}
		})
	}
}

// =============================================================================
// UpdateComment 测试
// =============================================================================

func TestCommentStore_UpdateComment(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewCommentStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	// 创建测试评论
	comment := &model.Comment{
		IssueID: issue1.ID,
		UserID:  user1.ID,
		Body:    "原始评论内容",
	}
	if err := store.CreateComment(ctx, comment); err != nil {
		t.Fatalf("创建测试评论失败: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		body    string
		wantErr bool
	}{
		{
			name:    "正常更新",
			id:      comment.ID,
			body:    "更新后的评论内容",
			wantErr: false,
		},
		{
			name:    "不存在的 ID",
			id:      uuid.New(),
			body:    "测试内容",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.UpdateComment(ctx, tt.id, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// 验证更新后的内容和 edited_at
				got, err := store.GetCommentByID(ctx, tt.id)
				if err != nil {
					t.Fatalf("获取更新后的评论失败: %v", err)
				}
				if got.Body != tt.body {
					t.Errorf("UpdateComment() Body = %v, want %v", got.Body, tt.body)
				}
				if got.EditedAt == nil {
					t.Error("UpdateComment() 未设置 EditedAt")
				}
			}
		})
	}
}

// =============================================================================
// DeleteComment 测试
// =============================================================================

func TestCommentStore_DeleteComment(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewCommentStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	// 创建父评论
	parentComment := &model.Comment{
		IssueID: issue1.ID,
		UserID:  user1.ID,
		Body:    "父评论",
	}
	if err := store.CreateComment(ctx, parentComment); err != nil {
		t.Fatalf("创建父评论失败: %v", err)
	}

	// 创建子回复
	reply := &model.Comment{
		IssueID: issue1.ID,
		ParentID: &parentComment.ID,
		UserID:  user1.ID,
		Body:    "子回复",
	}
	if err := store.CreateComment(ctx, reply); err != nil {
		t.Fatalf("创建子回复失败: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "删除父评论（级联删除子回复）",
			id:      parentComment.ID,
			wantErr: false,
		},
		{
			name:    "删除不存在的评论",
			id:      uuid.New(),
			wantErr: false, // GORM 删除不存在的记录不报错
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteComment(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

	// 验证子回复也被删除（级联）
	_, err := store.GetCommentByID(ctx, reply.ID)
	if err == nil {
		t.Error("DeleteComment() 子回复应该被级联删除")
	}
}

// =============================================================================
// GetCommentsByIssueIDWithTotal 测试
// =============================================================================

func TestCommentStore_GetCommentsByIssueIDWithTotal(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewCommentStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	// 创建 5 个评论
	for i := 0; i < 5; i++ {
		comment := &model.Comment{
			IssueID: issue1.ID,
			UserID:  user1.ID,
			Body:    "测试评论",
		}
		if err := store.CreateComment(ctx, comment); err != nil {
			t.Fatalf("创建测试评论失败: %v", err)
		}
	}

	tests := []struct {
		name      string
		issueID   uuid.UUID
		opts      *ListCommentsOptions
		wantCount int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "分页查询带总数",
			issueID:   issue1.ID,
			opts:      &ListCommentsOptions{Page: 1, PageSize: 3},
			wantCount: 3,
			wantTotal: 5,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := store.GetCommentsByIssueIDWithTotal(ctx, tt.issueID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommentsByIssueIDWithTotal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetCommentsByIssueIDWithTotal() got %d comments, want %d", len(got), tt.wantCount)
			}
			if total != tt.wantTotal {
				t.Errorf("GetCommentsByIssueIDWithTotal() got total %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

// =============================================================================
// 测试辅助函数
// =============================================================================

type commentTestFixtures struct {
	users           []*model.User
	issues          []*model.Issue
	team            *model.Team
	status          *model.WorkflowState
	parentCommentID uuid.UUID
}

func setupCommentTestFixtures(t *testing.T, db *gorm.DB) *commentTestFixtures {
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
			Email:        prefix + "_comment_user" + string(rune('1'+i)) + "@example.com",
			Username:     prefix + "_cmtuser" + string(rune('1'+i)),
			Name:         "Test Comment User " + string(rune('1'+i)),
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
		Name:        prefix + "_Comment Team",
		Key:         "CMT",
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
			Title:       prefix + "_Comment Issue " + string(rune('1'+i)),
			StatusID:    status.ID,
			Priority:    0,
			CreatedByID: users[0].ID,
		}
		if err := issueStore.Create(ctx, issue); err != nil {
			t.Fatalf("创建 Issue 失败: %v", err)
		}
		issues[i] = issue
	}

	return &commentTestFixtures{
		users:  users,
		issues: issues,
		team:   team,
		status: status,
	}
}

// GetCommentStoreWithPreload 辅助函数：获取评论并预加载用户信息
func TestCommentStore_GetCommentWithUser(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewCommentStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	// 创建评论
	comment := &model.Comment{
		IssueID: issue1.ID,
		UserID:  user1.ID,
		Body:    "测试评论",
	}
	if err := store.CreateComment(ctx, comment); err != nil {
		t.Fatalf("创建测试评论失败: %v", err)
	}

	// 获取带用户信息的评论
	got, err := store.GetCommentByID(ctx, comment.ID)
	if err != nil {
		t.Fatalf("获取评论失败: %v", err)
	}

	// 验证 ID 正确
	if got.ID != comment.ID {
		t.Errorf("GetCommentByID() ID = %v, want %v", got.ID, comment.ID)
	}

	// 检查时间戳
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt 应该被设置")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt 应该被设置")
	}
}

// TestCommentStore_EditedAt 测试编辑时间设置
func TestCommentStore_EditedAt(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewCommentStore(tx)
	ctx := context.Background()

	// 创建测试数据
	fixtures := setupCommentTestFixtures(t, tx)
	issue1 := fixtures.issues[0]
	user1 := fixtures.users[0]

	// 创建评论
	comment := &model.Comment{
		IssueID: issue1.ID,
		UserID:  user1.ID,
		Body:    "原始内容",
	}
	if err := store.CreateComment(ctx, comment); err != nil {
		t.Fatalf("创建测试评论失败: %v", err)
	}

	// 新创建的评论 EditedAt 应该为 nil
	if comment.EditedAt != nil {
		t.Error("新创建的评论 EditedAt 应该为 nil")
	}

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	// 更新评论
	if err := store.UpdateComment(ctx, comment.ID, "更新后的内容"); err != nil {
		t.Fatalf("更新评论失败: %v", err)
	}

	// 获取更新后的评论
	got, err := store.GetCommentByID(ctx, comment.ID)
	if err != nil {
		t.Fatalf("获取评论失败: %v", err)
	}

	// EditedAt 应该被设置
	if got.EditedAt == nil {
		t.Error("更新后的评论 EditedAt 不应该为 nil")
	}

	// EditedAt 应该晚于 CreatedAt
	if got.EditedAt.Before(got.CreatedAt) {
		t.Error("EditedAt 应该晚于 CreatedAt")
	}
}
