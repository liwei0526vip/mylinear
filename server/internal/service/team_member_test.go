package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TeamMemberService 测试
// =============================================================================

func TestTeamMemberService_ListMembers(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamStore := store.NewTeamStore(tx)
	svc := NewTeamMemberService(teamMemberStore, userStore, teamStore)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "ListMembers Test " + prefix,
		Slug: "listmembers-test-" + prefix,
	}
	tx.Create(workspace)

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "LM" + "ABC",
	}
	tx.Create(team)

	// 创建用户
	member1 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_m1@example.com",
		Username:     prefix + "_m1",
		Name:         "Member 1",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	member2 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_m2@example.com",
		Username:     prefix + "_m2",
		Name:         "Member 2",
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
	tx.Create(member1)
	tx.Create(member2)
	tx.Create(nonMember)

	// 添加成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member1.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member2.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name      string
		userID    uuid.UUID
		userRole  model.Role
		teamID    string
		wantCount int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "团队成员获取列表",
			userID:    member1.ID,
			userRole:  model.RoleMember,
			teamID:    team.ID.String(),
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "非团队成员拒绝",
			userID:   nonMember.ID,
			userRole: model.RoleMember,
			teamID:   team.ID.String(),
			wantErr:  true,
			errMsg:   "无权限",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)

			members, err := svc.ListMembers(ctx, tt.teamID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				assert.Len(t, members, tt.wantCount)
			}
		})
	}
}

func TestTeamMemberService_AddMember(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamStore := store.NewTeamStore(tx)
	svc := NewTeamMemberService(teamMemberStore, userStore, teamStore)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "AddMember Test " + prefix,
		Slug: "addmember-test-" + prefix,
	}
	tx.Create(workspace)

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "AM" + "ABC",
	}
	tx.Create(team)

	// 创建用户
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
	newUser := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_new@example.com",
		Username:     prefix + "_new",
		Name:         "New User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	tx.Create(owner)
	tx.Create(member)
	tx.Create(newUser)

	// 添加 Owner
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name       string
		userID     uuid.UUID
		userRole   model.Role
		teamID     string
		targetUser uuid.UUID
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Owner 添加成员",
			userID:     owner.ID,
			userRole:   model.RoleMember,
			teamID:     team.ID.String(),
			targetUser: newUser.ID,
			wantErr:    false,
		},
		{
			name:       "重复添加",
			userID:     owner.ID,
			userRole:   model.RoleMember,
			teamID:     team.ID.String(),
			targetUser: member.ID,
			wantErr:    true,
			errMsg:     "已是团队成员",
		},
		{
			name:       "普通成员无权限",
			userID:     member.ID,
			userRole:   model.RoleMember,
			teamID:     team.ID.String(),
			targetUser: newUser.ID,
			wantErr:    true,
			errMsg:     "无权限",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)

			err := svc.AddMember(ctx, tt.teamID, tt.targetUser.String(), model.RoleMember)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestTeamMemberService_RemoveMember(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamStore := store.NewTeamStore(tx)
	svc := NewTeamMemberService(teamMemberStore, userStore, teamStore)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "RemoveMember Test " + prefix,
		Slug: "removemember-test-" + prefix,
	}
	tx.Create(workspace)

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "RM" + "ABC",
	}
	tx.Create(team)

	// 创建用户
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
	tx.Create(owner)
	tx.Create(member)

	// 添加成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name       string
		userID     uuid.UUID
		userRole   model.Role
		teamID     string
		targetUser uuid.UUID
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Owner 移除成员",
			userID:     owner.ID,
			userRole:   model.RoleMember,
			teamID:     team.ID.String(),
			targetUser: member.ID,
			wantErr:    false,
		},
		{
			name:       "移除最后一个 Owner 拒绝",
			userID:     owner.ID,
			userRole:   model.RoleMember,
			teamID:     team.ID.String(),
			targetUser: owner.ID,
			wantErr:    true,
			errMsg:     "最后一个 Owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)

			err := svc.RemoveMember(ctx, tt.teamID, tt.targetUser.String())

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestTeamMemberService_UpdateRole(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamStore := store.NewTeamStore(tx)
	svc := NewTeamMemberService(teamMemberStore, userStore, teamStore)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "UpdateRole Test " + prefix,
		Slug: "updaterole-test-" + prefix,
	}
	tx.Create(workspace)

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "UR" + "ABC",
	}
	tx.Create(team)

	// 创建用户
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
	tx.Create(owner)
	tx.Create(member)

	// 添加成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name        string
		userID      uuid.UUID
		userRole    model.Role
		teamID      string
		targetUser  uuid.UUID
		newRole     model.Role
		wantErr     bool
		errMsg      string
	}{
		{
			name:       "Owner 提升成员",
			userID:     owner.ID,
			userRole:   model.RoleMember,
			teamID:     team.ID.String(),
			targetUser: member.ID,
			newRole:    model.RoleAdmin,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)

			err := svc.UpdateRole(ctx, tt.teamID, tt.targetUser.String(), tt.newRole)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				// 验证角色已更新
				role, _ := teamMemberStore.GetRole(ctx, tt.teamID, tt.targetUser.String())
				assert.Equal(t, tt.newRole, role)
			}
		})
	}
}
