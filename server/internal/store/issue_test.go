package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestIssueStore_Interface 测试 IssueStore 接口定义存在
func TestIssueStore_Interface(t *testing.T) {
	var _ IssueStore = (*issueStore)(nil)
}

// =============================================================================
// Create 测试 (Task 2.1)
// =============================================================================

func TestIssueStore_Create(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, user, team, state := setupIssueTestFixtures(t, tx)

	tests := []struct {
		name    string
		issue   *model.Issue
		wantErr bool
	}{
		{
			name: "正常创建 Issue",
			issue: &model.Issue{
				TeamID:      team.ID,
				Title:       "Test Issue 1",
				Description: ptr("This is a test issue"),
				StatusID:    state.ID,
				Priority:    model.PriorityMedium,
				CreatedByID: user.ID,
			},
			wantErr: false,
		},
		{
			name: "创建不带描述的 Issue",
			issue: &model.Issue{
				TeamID:      team.ID,
				Title:       "Test Issue 2",
				StatusID:    state.ID,
				Priority:    model.PriorityHigh,
				CreatedByID: user.ID,
			},
			wantErr: false,
		},
		{
			name: "创建带 Assignee 的 Issue",
			issue: &model.Issue{
				TeamID:      team.ID,
				Title:       "Test Issue 3",
				StatusID:    state.ID,
				Priority:    model.PriorityUrgent,
				AssigneeID:  &user.ID,
				CreatedByID: user.ID,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Create(ctx, tt.issue)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.issue.ID, "ID 应该被生成")
				assert.Greater(t, tt.issue.Number, 0, "Number 应该被自动生成")
				assert.GreaterOrEqual(t, tt.issue.Position, float64(0), "Position 应该有默认值")

				// 验证可以查询到
				var found model.Issue
				err = tx.Where("id = ?", tt.issue.ID).First(&found).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.issue.Title, found.Title)
			}
		})
	}

	_ = workspace // 避免未使用警告
}

// TestIssueStore_Create_NumberGeneration 测试 Issue Number 自动生成
func TestIssueStore_Create_NumberGeneration(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	_, user, team, state := setupIssueTestFixtures(t, tx)

	// 创建第一个 Issue
	issue1 := &model.Issue{
		TeamID:      team.ID,
		Title:       "First Issue",
		StatusID:    state.ID,
		CreatedByID: user.ID,
	}
	err := store.Create(ctx, issue1)
	assert.NoError(t, err)
	assert.Equal(t, 1, issue1.Number, "第一个 Issue 的 Number 应该是 1")

	// 创建第二个 Issue
	issue2 := &model.Issue{
		TeamID:      team.ID,
		Title:       "Second Issue",
		StatusID:    state.ID,
		CreatedByID: user.ID,
	}
	err = store.Create(ctx, issue2)
	assert.NoError(t, err)
	assert.Equal(t, 2, issue2.Number, "第二个 Issue 的 Number 应该是 2")

	// 创建第三个 Issue
	issue3 := &model.Issue{
		TeamID:      team.ID,
		Title:       "Third Issue",
		StatusID:    state.ID,
		CreatedByID: user.ID,
	}
	err = store.Create(ctx, issue3)
	assert.NoError(t, err)
	assert.Equal(t, 3, issue3.Number, "第三个 Issue 的 Number 应该是 3")
}

// TestIssueStore_Create_NumberUniquenessAcrossTeams 测试不同团队的 Number 独立
func TestIssueStore_Create_NumberUniquenessAcrossTeams(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据 - Workspace
	workspace := &model.Workspace{
		Name: "Multi Team WS",
		Slug: "multi-team-ws-" + uuid.New().String()[:8],
	}
	assert.NoError(t, tx.Create(workspace).Error)

	// User
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        "multiteam-" + uuid.New().String()[:8] + "@example.com",
		Username:     "multiteam" + uuid.New().String()[:8],
		Name:         "Multi Team User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	assert.NoError(t, tx.Create(user).Error)

	// Team 1
	team1 := &model.Team{
		WorkspaceID: workspace.ID,
		Key:         "T1" + uuid.New().String()[:6],
		Name:        "Team 1",
	}
	assert.NoError(t, tx.Create(team1).Error)

	// Team 2
	team2 := &model.Team{
		WorkspaceID: workspace.ID,
		Key:         "T2" + uuid.New().String()[:6],
		Name:        "Team 2",
	}
	assert.NoError(t, tx.Create(team2).Error)

	// State for Team 1
	state1 := &model.WorkflowState{
		TeamID:    team1.ID,
		Name:      "Backlog",
		Type:      model.StateTypeBacklog,
		Position:  1000,
		IsDefault: true,
	}
	assert.NoError(t, tx.Create(state1).Error)

	// State for Team 2
	state2 := &model.WorkflowState{
		TeamID:    team2.ID,
		Name:      "Backlog",
		Type:      model.StateTypeBacklog,
		Position:  1000,
		IsDefault: true,
	}
	assert.NoError(t, tx.Create(state2).Error)

	// 在 Team 1 创建 Issue
	issue1 := &model.Issue{
		TeamID:      team1.ID,
		Title:       "Team 1 Issue",
		StatusID:    state1.ID,
		CreatedByID: user.ID,
	}
	err := store.Create(ctx, issue1)
	assert.NoError(t, err)
	assert.Equal(t, 1, issue1.Number)

	// 在 Team 2 创建 Issue（Number 应该从 1 开始）
	issue2 := &model.Issue{
		TeamID:      team2.ID,
		Title:       "Team 2 Issue",
		StatusID:    state2.ID,
		CreatedByID: user.ID,
	}
	err = store.Create(ctx, issue2)
	assert.NoError(t, err)
	assert.Equal(t, 1, issue2.Number, "不同团队的 Number 应该独立")

	// 在 Team 1 再创建一个 Issue
	issue3 := &model.Issue{
		TeamID:      team1.ID,
		Title:       "Team 1 Issue 2",
		StatusID:    state1.ID,
		CreatedByID: user.ID,
	}
	err = store.Create(ctx, issue3)
	assert.NoError(t, err)
	assert.Equal(t, 2, issue3.Number, "Team 1 的第二个 Issue Number 应该是 2")
}

// =============================================================================
// GetByID 测试 (Task 2.3)
// =============================================================================

func TestIssueStore_GetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	_, user, team, state := setupIssueTestFixtures(t, tx)

	// 创建测试 Issue
	issue := &model.Issue{
		TeamID:      team.ID,
		Title:       "GetByID Test Issue",
		Description: ptr("Test description"),
		StatusID:    state.ID,
		Priority:    model.PriorityHigh,
		AssigneeID:  &user.ID,
		CreatedByID: user.ID,
	}
	assert.NoError(t, store.Create(ctx, issue))

	tests := []struct {
		name        string
		id          uuid.UUID
		wantErr     bool
		errCheck    func(t *testing.T, err error)
		resultCheck func(t *testing.T, issue *model.Issue)
	}{
		{
			name:    "正常获取",
			id:      issue.ID,
			wantErr: false,
			resultCheck: func(t *testing.T, issue *model.Issue) {
				assert.Equal(t, "GetByID Test Issue", issue.Title)
				assert.NotNil(t, issue.Team, "应该预加载 Team")
				assert.NotNil(t, issue.Status, "应该预加载 Status")
				assert.NotNil(t, issue.Assignee, "应该预加载 Assignee")
				assert.NotNil(t, issue.CreatedBy, "应该预加载 CreatedBy")
			},
		},
		{
			name:     "不存在的 Issue",
			id:       uuid.New(),
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := store.GetByID(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.resultCheck != nil {
					tt.resultCheck(t, result)
				}
			}
		})
	}
}

// =============================================================================
// List 测试 (Task 2.5)
// =============================================================================

func TestIssueStore_List(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	_, user, team, state := setupIssueTestFixtures(t, tx)

	// 创建多个 Issue
	for i := 1; i <= 5; i++ {
		issue := &model.Issue{
			TeamID:      team.ID,
			Title:       "List Test Issue",
			StatusID:    state.ID,
			Priority:    i % 5,
			CreatedByID: user.ID,
		}
		assert.NoError(t, store.Create(ctx, issue))
	}

	// 测试基础列表查询
	issues, total, err := store.List(ctx, team.ID, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, issues, 5)

	// 测试分页
	issues, total, err = store.List(ctx, team.ID, nil, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, issues, 2)

	issues, total, err = store.List(ctx, team.ID, nil, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, issues, 2)

	issues, total, err = store.List(ctx, team.ID, nil, 3, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, issues, 1)
}

// =============================================================================
// Update 测试 (Task 2.7)
// =============================================================================

func TestIssueStore_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	_, user, team, state := setupIssueTestFixtures(t, tx)

	// 创建测试 Issue
	issue := &model.Issue{
		TeamID:      team.ID,
		Title:       "Original Title",
		Description: ptr("Original description"),
		StatusID:    state.ID,
		Priority:    model.PriorityMedium,
		CreatedByID: user.ID,
	}
	assert.NoError(t, store.Create(ctx, issue))

	tests := []struct {
		name        string
		updateFn    func(*model.Issue)
		wantErr     bool
		resultCheck func(t *testing.T, updated *model.Issue)
	}{
		{
			name: "更新标题",
			updateFn: func(i *model.Issue) {
				i.Title = "Updated Title"
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Issue) {
				assert.Equal(t, "Updated Title", updated.Title)
			},
		},
		{
			name: "更新描述",
			updateFn: func(i *model.Issue) {
				i.Description = ptr("Updated description")
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Issue) {
				assert.Equal(t, "Updated description", *updated.Description)
			},
		},
		{
			name: "更新优先级",
			updateFn: func(i *model.Issue) {
				i.Priority = model.PriorityUrgent
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Issue) {
				assert.Equal(t, model.PriorityUrgent, updated.Priority)
			},
		},
		{
			name: "更新负责人",
			updateFn: func(i *model.Issue) {
				i.AssigneeID = &user.ID
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Issue) {
				assert.NotNil(t, updated.AssigneeID)
				assert.Equal(t, user.ID, *updated.AssigneeID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 应用更新
			tt.updateFn(issue)

			// 执行更新
			err := store.Update(ctx, issue)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// 重新查询验证
			updated, err := store.GetByID(ctx, issue.ID)
			assert.NoError(t, err)
			if tt.resultCheck != nil {
				tt.resultCheck(t, updated)
			}
		})
	}
}

// =============================================================================
// SoftDelete & Restore 测试 (Task 2.9)
// =============================================================================

func TestIssueStore_SoftDelete(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	_, user, team, state := setupIssueTestFixtures(t, tx)

	// 创建测试 Issue
	issue := &model.Issue{
		TeamID:      team.ID,
		Title:       "Delete Test Issue",
		StatusID:    state.ID,
		CreatedByID: user.ID,
	}
	assert.NoError(t, store.Create(ctx, issue))

	// 删除 Issue
	err := store.SoftDelete(ctx, issue.ID)
	assert.NoError(t, err)

	// 验证无法通过正常查询获取
	_, err = store.GetByID(ctx, issue.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// 验证数据库中仍存在（软删除）
	var count int64
	tx.Unscoped().Model(&model.Issue{}).Where("id = ?", issue.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestIssueStore_Restore(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	_, user, team, state := setupIssueTestFixtures(t, tx)

	// 创建并删除 Issue
	issue := &model.Issue{
		TeamID:      team.ID,
		Title:       "Restore Test Issue",
		StatusID:    state.ID,
		CreatedByID: user.ID,
	}
	assert.NoError(t, store.Create(ctx, issue))
	assert.NoError(t, store.SoftDelete(ctx, issue.ID))

	// 恢复 Issue
	err := store.Restore(ctx, issue.ID)
	assert.NoError(t, err)

	// 验证可以重新获取
	restored, err := store.GetByID(ctx, issue.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Restore Test Issue", restored.Title)
}

// =============================================================================
// UpdatePosition 测试 (Task 2.11)
// =============================================================================

func TestIssueStore_UpdatePosition(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	_, user, team, state := setupIssueTestFixtures(t, tx)

	// 创建第二个状态（用于跨状态拖拽测试）
	state2 := &model.WorkflowState{
		TeamID:    team.ID,
		Name:      "In Progress",
		Type:      model.StateTypeStarted,
		Position:  2000,
		IsDefault: false,
	}
	assert.NoError(t, tx.Create(state2).Error)

	// 创建多个 Issue
	issues := make([]*model.Issue, 3)
	for i := 0; i < 3; i++ {
		issues[i] = &model.Issue{
			TeamID:      team.ID,
			Title:       "Position Test Issue",
			StatusID:    state.ID,
			CreatedByID: user.ID,
		}
		assert.NoError(t, store.Create(ctx, issues[i]))
	}

	tests := []struct {
		name        string
		issueID     uuid.UUID
		position    float64
		statusID    *uuid.UUID
		wantErr     bool
		resultCheck func(t *testing.T)
	}{
		{
			name:     "更新位置（同一状态内）",
			issueID:  issues[0].ID,
			position: 500,
			statusID: nil,
			wantErr:  false,
			resultCheck: func(t *testing.T) {
				issue, err := store.GetByID(ctx, issues[0].ID)
				assert.NoError(t, err)
				assert.Equal(t, float64(500), issue.Position)
			},
		},
		{
			name:     "跨状态拖拽（更新位置和状态）",
			issueID:  issues[1].ID,
			position: 1500,
			statusID: &state2.ID,
			wantErr:  false,
			resultCheck: func(t *testing.T) {
				issue, err := store.GetByID(ctx, issues[1].ID)
				assert.NoError(t, err)
				assert.Equal(t, float64(1500), issue.Position)
				assert.Equal(t, state2.ID, issue.StatusID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.UpdatePosition(ctx, tt.issueID, tt.position, tt.statusID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.resultCheck != nil {
				tt.resultCheck(t)
			}
		})
	}
}

// =============================================================================
// ListBySubscription 测试 (Task 2.13)
// =============================================================================

func TestIssueStore_ListBySubscription(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	issueStore := NewIssueStore(tx)
	subStore := NewIssueSubscriptionStore(tx)
	ctx := context.Background()

	// 准备测试数据
	_, user, team, state := setupIssueTestFixtures(t, tx)

	// 创建第二个用户
	user2 := &model.User{
		WorkspaceID:  team.WorkspaceID,
		Email:        "subscriber-" + uuid.New().String()[:8] + "@example.com",
		Username:     "subscriber" + uuid.New().String()[:8],
		Name:         "Subscriber User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	assert.NoError(t, tx.Create(user2).Error)

	// 创建多个 Issue
	for i := 0; i < 3; i++ {
		issue := &model.Issue{
			TeamID:      team.ID,
			Title:       "Subscription Test Issue",
			StatusID:    state.ID,
			CreatedByID: user.ID,
		}
		assert.NoError(t, issueStore.Create(ctx, issue))

		// 用户1 订阅所有 Issue
		assert.NoError(t, subStore.Subscribe(ctx, issue.ID, user.ID))

		// 用户2 只订阅第一个 Issue
		if i == 0 {
			assert.NoError(t, subStore.Subscribe(ctx, issue.ID, user2.ID))
		}
	}

	// 测试用户1订阅的 Issue
	issues, err := issueStore.ListBySubscription(ctx, user.ID)
	assert.NoError(t, err)
	assert.Len(t, issues, 3, "用户1应该订阅了3个Issue")

	// 测试用户2订阅的 Issue
	issues, err = issueStore.ListBySubscription(ctx, user2.ID)
	assert.NoError(t, err)
	assert.Len(t, issues, 1, "用户2应该只订阅了1个Issue")
}

// =============================================================================
// 辅助函数
// =============================================================================

// setupIssueTestFixtures 创建测试所需的基础数据
func setupIssueTestFixtures(t *testing.T, tx *gorm.DB) (*model.Workspace, *model.User, *model.Team, *model.WorkflowState) {
	workspace := &model.Workspace{
		Name: "Fixture WS",
		Slug: "fixture-ws-" + uuid.New().String()[:8],
	}
	assert.NoError(t, tx.Create(workspace).Error)

	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        "fixture-" + uuid.New().String()[:8] + "@example.com",
		Username:     "fixture" + uuid.New().String()[:8],
		Name:         "Fixture User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	assert.NoError(t, tx.Create(user).Error)

	team := &model.Team{
		WorkspaceID: workspace.ID,
		Key:         "FX" + uuid.New().String()[:6],
		Name:        "Fixture Team",
	}
	assert.NoError(t, tx.Create(team).Error)

	state := &model.WorkflowState{
		TeamID:    team.ID,
		Name:      "Backlog",
		Type:      model.StateTypeBacklog,
		Position:  1000,
		IsDefault: true,
	}
	assert.NoError(t, tx.Create(state).Error)

	return workspace, user, team, state
}

// ptr 返回字符串的指针
func ptr(s string) *string {
	return &s
}
