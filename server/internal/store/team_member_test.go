package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestTeamMemberStore_Interface 测试 TeamMemberStore 接口定义存在
func TestTeamMemberStore_Interface(t *testing.T) {
	var _ TeamMemberStore = (*teamMemberStore)(nil)
}

// =============================================================================
// List 测试
// =============================================================================

func TestTeamMemberStore_List(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 使用事务进行测试隔离
	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamMemberStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "TeamMember List Test " + prefix,
		Slug: "teammember-list-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "TM" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	users := make([]*model.User, 3)
	for i := 0; i < 3; i++ {
		users[i] = &model.User{
			WorkspaceID:  workspace.ID,
			Email:        prefix + "_" + string(rune('a'+i)) + "@example.com",
			Username:     prefix + "_user_" + string(rune('a'+i)),
			Name:         "Test User " + string(rune('A'+i)),
			PasswordHash: "hash",
			Role:         model.RoleMember,
		}
		if err := tx.Create(users[i]).Error; err != nil {
			t.Fatalf("创建测试用户失败: %v", err)
		}
	}

	// 添加团队成员
	for i, user := range users {
		member := &model.TeamMember{
			TeamID:   team.ID,
			UserID:   user.ID,
			Role:     model.RoleMember,
			JoinedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		if i == 0 {
			member.Role = model.RoleAdmin // 第一个是 Owner
		}
		if err := tx.Create(member).Error; err != nil {
			t.Fatalf("添加团队成员失败: %v", err)
		}
	}

	tests := []struct {
		name      string
		teamID    string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "正常列表",
			teamID:    team.ID.String(),
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "空团队",
			teamID:    uuid.New().String(),
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			members, err := store.List(ctx, tt.teamID)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Len(t, members, tt.wantCount, "List() members 数量不匹配")

			// 验证返回的成员都属于正确的团队
			for _, member := range members {
				assert.Equal(t, tt.teamID, member.TeamID.String(), "List() 返回了错误的成员")
			}
		})
	}
}

// =============================================================================
// Add 测试
// =============================================================================

func TestTeamMemberStore_Add(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamMemberStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "TeamMember Add Test " + prefix,
		Slug: "teammember-add-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "TA" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
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
		t.Fatalf("创建测试用户失败: %v", err)
	}

	tests := []struct {
		name    string
		member  *model.TeamMember
		wantErr bool
		errMsg  string
	}{
		{
			name: "正常添加成员",
			member: &model.TeamMember{
				TeamID:   team.ID,
				UserID:   user.ID,
				Role:     model.RoleMember,
				JoinedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "重复添加成员",
			member: &model.TeamMember{
				TeamID:   team.ID,
				UserID:   user.ID,
				Role:     model.RoleMember,
				JoinedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "已是团队成员",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Add(ctx, tt.member)

			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				// 验证成员已添加
				members, _ := store.List(ctx, tt.member.TeamID.String())
				found := false
				for _, m := range members {
					if m.UserID.String() == tt.member.UserID.String() {
						found = true
						break
					}
				}
				assert.True(t, found, "Add() 成员应该已添加到团队")
			}
		})
	}
}

// =============================================================================
// Remove 测试
// =============================================================================

func TestTeamMemberStore_Remove(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamMemberStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "TeamMember Remove Test " + prefix,
		Slug: "teammember-remove-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "TR" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	user1 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_1@example.com",
		Username:     prefix + "_user1",
		Name:         "Test User 1",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_2@example.com",
		Username:     prefix + "_user2",
		Name:         "Test User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := tx.Create(user1).Error; err != nil {
		t.Fatalf("创建测试用户1失败: %v", err)
	}
	if err := tx.Create(user2).Error; err != nil {
		t.Fatalf("创建测试用户2失败: %v", err)
	}

	// 添加成员 - user1 是 Owner，user2 是普通成员
	tx.Create(&model.TeamMember{
		TeamID: team.ID, UserID: user1.ID, Role: model.RoleAdmin, JoinedAt: time.Now(),
	})
	tx.Create(&model.TeamMember{
		TeamID: team.ID, UserID: user2.ID, Role: model.RoleMember, JoinedAt: time.Now(),
	})

	tests := []struct {
		name    string
		teamID  string
		userID  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "正常移除成员",
			teamID:  team.ID.String(),
			userID:  user2.ID.String(),
			wantErr: false,
		},
		{
			name:    "成员不存在",
			teamID:  team.ID.String(),
			userID:  uuid.New().String(),
			wantErr: false, // 移除不存在的成员不报错
		},
		{
			name:    "移除最后一个 Owner 拒绝",
			teamID:  team.ID.String(),
			userID:  user1.ID.String(),
			wantErr: true,
			errMsg:  "最后一个 Owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Remove(ctx, tt.teamID, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				// 验证成员已移除
				role, err := store.GetRole(ctx, tt.teamID, tt.userID)
				assert.NoError(t, err)
				assert.Empty(t, role, "Remove() 成员应该已从团队移除")
			}
		})
	}
}

// =============================================================================
// UpdateRole 测试
// =============================================================================

func TestTeamMemberStore_UpdateRole(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamMemberStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "TeamMember UpdateRole Test " + prefix,
		Slug: "teammember-updaterole-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "TU" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	user1 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_1@example.com",
		Username:     prefix + "_user1",
		Name:         "Test User 1",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_2@example.com",
		Username:     prefix + "_user2",
		Name:         "Test User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := tx.Create(user1).Error; err != nil {
		t.Fatalf("创建测试用户1失败: %v", err)
	}
	if err := tx.Create(user2).Error; err != nil {
		t.Fatalf("创建测试用户2失败: %v", err)
	}

	// 添加成员 - user1 是 Owner，user2 是普通成员
	tx.Create(&model.TeamMember{
		TeamID: team.ID, UserID: user1.ID, Role: model.RoleAdmin, JoinedAt: time.Now(),
	})
	tx.Create(&model.TeamMember{
		TeamID: team.ID, UserID: user2.ID, Role: model.RoleMember, JoinedAt: time.Now(),
	})

	tests := []struct {
		name    string
		teamID  string
		userID  string
		newRole model.Role
		wantErr bool
		errMsg  string
	}{
		{
			name:    "最后一个 Owner 降级拒绝",
			teamID:  team.ID.String(),
			userID:  user1.ID.String(),
			newRole: model.RoleMember,
			wantErr: true,
			errMsg:  "最后一个 Owner",
		},
		{
			name:    "提升为 Owner",
			teamID:  team.ID.String(),
			userID:  user2.ID.String(),
			newRole: model.RoleAdmin,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.UpdateRole(ctx, tt.teamID, tt.userID, tt.newRole)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				// 验证角色已更新
				role, err := store.GetRole(ctx, tt.teamID, tt.userID)
				assert.NoError(t, err)
				assert.Equal(t, tt.newRole, role, "UpdateRole() 角色应该已更新")
			}
		})
	}
}

// =============================================================================
// GetRole 测试
// =============================================================================

func TestTeamMemberStore_GetRole(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewTeamMemberStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "TeamMember GetRole Test " + prefix,
		Slug: "teammember-getrole-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "TG" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	user1 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_1@example.com",
		Username:     prefix + "_user1",
		Name:         "Test User 1",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_2@example.com",
		Username:     prefix + "_user2",
		Name:         "Test User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := tx.Create(user1).Error; err != nil {
		t.Fatalf("创建测试用户1失败: %v", err)
	}
	if err := tx.Create(user2).Error; err != nil {
		t.Fatalf("创建测试用户2失败: %v", err)
	}

	// 添加成员 - user1 是 Owner，user2 是普通成员
	tx.Create(&model.TeamMember{
		TeamID: team.ID, UserID: user1.ID, Role: model.RoleAdmin, JoinedAt: time.Now(),
	})
	tx.Create(&model.TeamMember{
		TeamID: team.ID, UserID: user2.ID, Role: model.RoleMember, JoinedAt: time.Now(),
	})

	tests := []struct {
		name     string
		teamID   string
		userID   string
		wantRole model.Role
	}{
		{
			name:     "Owner 角色",
			teamID:   team.ID.String(),
			userID:   user1.ID.String(),
			wantRole: model.RoleAdmin,
		},
		{
			name:     "Member 角色",
			teamID:   team.ID.String(),
			userID:   user2.ID.String(),
			wantRole: model.RoleMember,
		},
		{
			name:     "非成员",
			teamID:   team.ID.String(),
			userID:   uuid.New().String(),
			wantRole: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := store.GetRole(ctx, tt.teamID, tt.userID)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRole, role, "GetRole() 角色不匹配")
		})
	}
}

// =============================================================================
// GetTeamRole 辅助函数测试
// =============================================================================

func TestGetTeamRole(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "GetTeamRole Test " + prefix,
		Slug: "getteamrole-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "GT" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	owner := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_owner@example.com",
		Username:     prefix + "_owner",
		Name:         "Owner",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	member := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_member@example.com",
		Username:     prefix + "_member",
		Name:         "Member",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	admin := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_admin@example.com",
		Username:     prefix + "_admin",
		Name:         "Admin",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	tx.Create(owner)
	tx.Create(member)
	tx.Create(admin)

	// 添加团队成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name     string
		userID   uuid.UUID
		teamID   uuid.UUID
		wantRole model.Role
		wantErr  bool
	}{
		{
			name:     "Owner 角色",
			userID:   owner.ID,
			teamID:   team.ID,
			wantRole: model.RoleAdmin,
			wantErr:  false,
		},
		{
			name:     "Member 角色",
			userID:   member.ID,
			teamID:   team.ID,
			wantRole: model.RoleMember,
			wantErr:  false,
		},
		{
			name:     "非成员",
			userID:   admin.ID,
			teamID:   team.ID,
			wantRole: "",
			wantErr:  false,
		},
		{
			name:     "Admin 用户但不绕过（需要业务层处理）",
			userID:   admin.ID,
			teamID:   team.ID,
			wantRole: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := GetTeamRole(ctx, tx, tt.userID, tt.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTeamRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantRole, role)
		})
	}
}

// =============================================================================
// IsTeamOwner 辅助函数测试
// =============================================================================

func TestIsTeamOwner(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "IsTeamOwner Test " + prefix,
		Slug: "isteamowner-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "IO" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	owner := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_owner@example.com",
		Username:     prefix + "_owner",
		Name:         "Owner",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	member := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_member@example.com",
		Username:     prefix + "_member",
		Name:         "Member",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	admin := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_admin@example.com",
		Username:     prefix + "_admin",
		Name:         "Admin",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	tx.Create(owner)
	tx.Create(member)
	tx.Create(admin)

	// 添加团队成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name      string
		userID    uuid.UUID
		teamID    uuid.UUID
		wantOwner bool
		wantErr   bool
	}{
		{
			name:      "是 Owner",
			userID:    owner.ID,
			teamID:    team.ID,
			wantOwner: true,
			wantErr:   false,
		},
		{
			name:      "不是 Owner（普通成员）",
			userID:    member.ID,
			teamID:    team.ID,
			wantOwner: false,
			wantErr:   false,
		},
		{
			name:      "不是 Owner（非成员 Admin）",
			userID:    admin.ID,
			teamID:    team.ID,
			wantOwner: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isOwner, err := IsTeamOwner(ctx, tx, tt.userID, tt.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsTeamOwner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantOwner, isOwner)
		})
	}
}

// =============================================================================
// IsTeamMember 辅助函数测试
// =============================================================================

func TestIsTeamMember(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "IsTeamMember Test " + prefix,
		Slug: "isteammember-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "IM" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	owner := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_owner@example.com",
		Username:     prefix + "_owner",
		Name:         "Owner",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	member := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_member@example.com",
		Username:     prefix + "_member",
		Name:         "Member",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	nonMember := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_non@example.com",
		Username:     prefix + "_non",
		Name:         "NonMember",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	tx.Create(owner)
	tx.Create(member)
	tx.Create(nonMember)

	// 添加团队成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name       string
		userID     uuid.UUID
		teamID     uuid.UUID
		wantMember bool
		wantErr    bool
	}{
		{
			name:       "是成员（Owner）",
			userID:     owner.ID,
			teamID:     team.ID,
			wantMember: true,
			wantErr:    false,
		},
		{
			name:       "是成员（普通成员）",
			userID:     member.ID,
			teamID:     team.ID,
			wantMember: true,
			wantErr:    false,
		},
		{
			name:       "不是成员",
			userID:     nonMember.ID,
			teamID:     team.ID,
			wantMember: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isMember, err := IsTeamMember(ctx, tx, tt.userID, tt.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsTeamMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantMember, isMember)
		})
	}
}
