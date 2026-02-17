package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mylinear/server/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestTeamStore_Interface 测试 TeamStore 接口定义存在
func TestTeamStore_Interface(t *testing.T) {
	var _ TeamStore = (*teamStore)(nil)
}

// =============================================================================
// ValidateTeamKey 单元测试
// =============================================================================

func TestValidateTeamKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		// 有效 Key
		{"有效 - 2位大写字母", "AB", false},
		{"有效 - 3位大写字母", "ABC", false},
		{"有效 - 大写字母+数字", "AB1", false},
		{"有效 - 10位最大长度", "ABCDEFGHIJ", false},
		{"有效 - 混合", "TEAM123", false},

		// 无效 Key
		{"无效 - 过短（1位）", "A", true},
		{"无效 - 过长（11位）", "ABCDEFGHIJK", true},
		{"无效 - 小写字母", "abc", true},
		{"无效 - 数字开头", "1AB", true},
		{"无效 - 包含小写", "AbC", true},
		{"无效 - 包含特殊字符", "AB-C", true},
		{"无效 - 包含空格", "AB C", true},
		{"无效 - 空字符串", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeamKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTeamKey(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
		})
	}
}


// =============================================================================
// List 测试
// =============================================================================

func TestTeamStore_List(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 使用事务进行测试隔离
	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "Team List Test " + prefix,
		Slug: "team-list-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建多个测试团队
	for i := 1; i <= 5; i++ {
		team := &model.Team{
			WorkspaceID: workspace.ID,
			Name:        "Team " + string(rune('A'+i-1)) + " " + prefix,
			Key:         "TM" + string(rune('A'+i-1)) + prefix[:2],
		}
		if err := tx.Create(team).Error; err != nil {
			t.Fatalf("创建测试团队失败: %v", err)
		}
	}

	tests := []struct {
		name            string
		workspaceID     string
		page            int
		pageSize        int
		wantCount       int
		wantTotal       int64
		wantErr         bool
	}{
		{
			name:        "按 workspace 过滤 - 获取所有",
			workspaceID: workspace.ID.String(),
			page:        1,
			pageSize:    10,
			wantCount:   5,
			wantTotal:   5,
			wantErr:     false,
		},
		{
			name:        "按 workspace 过滤 - 分页第1页",
			workspaceID: workspace.ID.String(),
			page:        1,
			pageSize:    3,
			wantCount:   3,
			wantTotal:   5,
			wantErr:     false,
		},
		{
			name:        "按 workspace 过滤 - 分页第2页",
			workspaceID: workspace.ID.String(),
			page:        2,
			pageSize:    3,
			wantCount:   2,
			wantTotal:   5,
			wantErr:     false,
		},
		{
			name:        "空 workspace - 返回空列表",
			workspaceID: uuid.New().String(),
			page:        1,
			pageSize:    10,
			wantCount:   0,
			wantTotal:   0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teams, total, err := store.List(ctx, tt.workspaceID, tt.page, tt.pageSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantTotal, total, "List() total 不匹配")
			assert.Len(t, teams, tt.wantCount, "List() teams 数量不匹配")

			// 验证所有返回的团队都属于正确的工作区
			for _, team := range teams {
				assert.Equal(t, tt.workspaceID, team.WorkspaceID.String(), "List() 返回了错误的团队")
			}
		})
	}
}

// =============================================================================
// GetByID 测试
// =============================================================================

func TestTeamStore_GetByID(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 使用事务进行测试隔离
	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "Team Get Test " + prefix,
		Slug: "team-get-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Test Team " + prefix,
		Key:         "TT" + prefix[:3],
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	tests := []struct {
		name      string
		id        string
		wantErr   bool
		wantFound bool
		checkTeam func(*model.Team) bool
	}{
		{
			name:      "正常获取团队",
			id:        team.ID.String(),
			wantErr:   false,
			wantFound: true,
			checkTeam: func(t *model.Team) bool {
				return t.Name == team.Name && t.Key == team.Key
			},
		},
		{
			name:      "团队不存在",
			id:        uuid.New().String(),
			wantErr:   true,
			wantFound: false,
			checkTeam: nil,
		},
		{
			name:      "无效的ID格式",
			id:        "invalid-uuid",
			wantErr:   true,
			wantFound: false,
			checkTeam: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := store.GetByID(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantFound && found == nil {
				t.Error("GetByID() 应该找到团队")
				return
			}

			if !tt.wantFound && found != nil {
				t.Error("GetByID() 不应该找到团队")
				return
			}

			if tt.checkTeam != nil && found != nil {
				if !tt.checkTeam(found) {
					t.Error("GetByID() 返回的团队字段不符合预期")
				}
			}
		})
	}
}

// =============================================================================
// Create 测试
// =============================================================================

func TestTeamStore_Create(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 使用事务进行测试隔离
	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "Team Create Test " + prefix,
		Slug: "team-create-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	tests := []struct {
		name    string
		team    *model.Team
		wantErr bool
	}{
		{
			name: "正常创建团队",
			team: &model.Team{
				WorkspaceID: workspace.ID,
				Name:        "New Team " + prefix,
				Key:         "NT" + "ABC", // 使用固定大写字母
			},
			wantErr: false,
		},
		{
			name: "Key 重复",
			team: &model.Team{
				WorkspaceID: workspace.ID,
				Name:        "Duplicate Key Team",
				Key:         "NTABC", // 与上面相同
			},
			wantErr: true,
		},
		{
			name: "Key 格式错误 - 小写",
			team: &model.Team{
				WorkspaceID: workspace.ID,
				Name:        "Lowercase Key Team",
				Key:         "abc", // 小写
			},
			wantErr: true,
		},
		{
			name: "Key 格式错误 - 数字开头",
			team: &model.Team{
				WorkspaceID: workspace.ID,
				Name:        "Number Key Team",
				Key:         "1AB", // 数字开头
			},
			wantErr: true,
		},
		{
			name: "Key 格式错误 - 过短",
			team: &model.Team{
				WorkspaceID: workspace.ID,
				Name:        "Short Key Team",
				Key:         "A", // 过短
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Create(ctx, tt.team)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证团队已创建并有 ID
				assert.NotEqual(t, uuid.Nil, tt.team.ID, "Create() 应该设置团队 ID")

				// 验证可以通过 ID 查询到
				found, err := store.GetByID(ctx, tt.team.ID.String())
				assert.NoError(t, err)
				assert.Equal(t, tt.team.Name, found.Name)
				assert.Equal(t, tt.team.Key, found.Key)
			}
		})
	}
}

// =============================================================================
// Update 测试
// =============================================================================

func TestTeamStore_Update(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{Name: "Update Test " + prefix, Slug: "update-test-" + prefix}
	tx.Create(workspace)

	t.Run("更新名称", func(t *testing.T) {
		team := &model.Team{
			WorkspaceID: workspace.ID,
			Name:        "Original Name " + prefix,
			Key:         "UA" + prefix[:3],
		}
		if err := tx.Create(team).Error; err != nil {
			t.Fatalf("创建测试团队失败: %v", err)
		}

		team.Name = "Updated Name " + prefix
		err := store.Update(ctx, team)
		assert.NoError(t, err)

		found, _ := store.GetByID(ctx, team.ID.String())
		assert.Equal(t, "Updated Name "+prefix, found.Name)
	})

	t.Run("更新 Key", func(t *testing.T) {
		team := &model.Team{
			WorkspaceID: workspace.ID,
			Name:        "Update Key Test " + prefix,
			Key:         "UB" + prefix[:3],
		}
		if err := tx.Create(team).Error; err != nil {
			t.Fatalf("创建测试团队失败: %v", err)
		}

		team.Key = "UC" + prefix[:3]
		err := store.Update(ctx, team)
		assert.NoError(t, err)

		found, _ := store.GetByID(ctx, team.ID.String())
		assert.Equal(t, "UC"+prefix[:3], found.Key)
	})
}

// =============================================================================
// SoftDelete 测试
// =============================================================================

func TestTeamStore_SoftDelete(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{Name: "Delete Test " + prefix, Slug: "delete-test-" + prefix}
	tx.Create(workspace)

	t.Run("正常删除", func(t *testing.T) {
		team := &model.Team{
			WorkspaceID: workspace.ID,
			Name:        "To Delete " + prefix,
			Key:         "DA" + prefix[:3],
		}
		if err := tx.Create(team).Error; err != nil {
			t.Fatalf("创建测试团队失败: %v", err)
		}

		err := store.SoftDelete(ctx, team.ID.String())
		assert.NoError(t, err)

		// 验证团队已被删除
		_, err = store.GetByID(ctx, team.ID.String())
		assert.Error(t, err, "团队应该已被删除")
	})

	t.Run("团队不存在", func(t *testing.T) {
		err := store.SoftDelete(ctx, uuid.New().String())
		assert.NoError(t, err, "删除不存在的团队不应报错")
	})
}

// =============================================================================
// CountIssuesByTeam 测试
// =============================================================================

func TestTeamStore_CountIssuesByTeam(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{Name: "Count Test " + prefix, Slug: "count-test-" + prefix}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建工作区失败: %v", err)
	}

	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Count Team " + prefix,
		Key:         "CT" + prefix[:3],
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建团队失败: %v", err)
	}

	// 创建测试用户
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "@example.com",
		Username:     prefix + "_user",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 创建测试 WorkflowState
	status := &model.WorkflowState{
		TeamID: team.ID,
		Name:   "Todo",
		Type:   model.StateTypeUnstarted,
		Color:  "#gray",
	}
	if err := tx.Create(status).Error; err != nil {
		t.Fatalf("创建状态失败: %v", err)
	}

	// 创建测试 Issue
	for i := 1; i <= 3; i++ {
		issue := &model.Issue{
			TeamID:      team.ID,
			Number:      i,
			Title:       "Test Issue " + string(rune('A'+i)),
			Priority:    model.PriorityMedium,
			StatusID:    status.ID,
			CreatedByID: user.ID,
		}
		if err := tx.Create(issue).Error; err != nil {
			t.Fatalf("创建 Issue 失败: %v", err)
		}
	}

	tests := []struct {
		name      string
		teamID    string
		wantCount int64
		wantErr   bool
	}{
		{
			name:      "有 Issue 的团队",
			teamID:    team.ID.String(),
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "无 Issue 的团队",
			teamID:    uuid.New().String(),
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := store.CountIssuesByTeam(ctx, tt.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountIssuesByTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}
