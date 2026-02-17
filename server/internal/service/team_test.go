package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// var testTeamServiceDB *gorm.DB // Already defined in service.go or main_test.go? No, I will define it in main_test.go globally or allow it to be reused.
// Wait, testTeamServiceDB is used in team_test.go.
// I will keep the variable declaration but remove assignment in init().
var testTeamServiceDB *gorm.DB

// =============================================================================
// CreateTeam 测试
// =============================================================================

func TestTeamService_CreateTeam(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := NewWorkflowService(workflowStateStore, teamStore)
	svc := NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "CreateTeam Test " + prefix,
		Slug: "createteam-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试用户
	admin := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_admin@example.com",
		Username:     prefix + "_admin",
		Name:         "Admin",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	member := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_member@example.com",
		Username:     prefix + "_member",
		Name:         "Member",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	tx.Create(admin)
	tx.Create(member)

	tests := []struct {
		name     string
		userID   uuid.UUID
		userRole model.Role
		teamName string
		teamKey  string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Admin 创建团队",
			userID:   admin.ID,
			userRole: admin.Role,
			teamName: "New Team " + prefix,
			teamKey:  "NT" + "ABC",
			wantErr:  false,
		},
		{
			name:     "Key 格式错误",
			userID:   admin.ID,
			userRole: admin.Role,
			teamName: "Bad Key Team",
			teamKey:  "abc", // 小写
			wantErr:  true,
			errMsg:   "团队标识符",
		},
		{
			name:     "Member 无权限",
			userID:   member.ID,
			userRole: member.Role,
			teamName: "Member Team",
			teamKey:  "MT" + "ABC",
			wantErr:  true,
			errMsg:   "无权限",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)
			ctx = context.WithValue(ctx, "workspace_id", workspace.ID)

			team, err := svc.CreateTeam(ctx, tt.teamName, tt.teamKey, "")

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				assert.NotNil(t, team)
				assert.Equal(t, tt.teamName, team.Name)
				assert.Equal(t, tt.teamKey, team.Key)

				// 验证创建者成为 Owner
				role, _ := teamMemberStore.GetRole(ctx, team.ID.String(), tt.userID.String())
				assert.Equal(t, model.RoleAdmin, role, "创建者应该成为团队 Owner")
			}
		})
	}
}

// =============================================================================
// ListTeams 测试
// =============================================================================

func TestTeamService_ListTeams(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := NewWorkflowService(workflowStateStore, teamStore)
	svc := NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "ListTeams Test " + prefix,
		Slug: "listteams-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试用户
	admin := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_admin@example.com",
		Username:     prefix + "_admin",
		Name:         "Admin",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	tx.Create(admin)

	// 创建多个团队
	for i := 0; i < 5; i++ {
		team := &model.Team{
			WorkspaceID: workspace.ID,
			Name:        "Team " + string(rune('A'+i)) + " " + prefix,
			Key:         "TL" + string(rune('A'+i)) + "ABC",
		}
		tx.Create(team)
	}

	tests := []struct {
		name        string
		userID      uuid.UUID
		workspaceID uuid.UUID
		page        int
		pageSize    int
		wantCount   int
		wantTotal   int64
		wantErr     bool
	}{
		{
			name:        "按 workspace 过滤 - 获取所有",
			userID:      admin.ID,
			workspaceID: workspace.ID,
			page:        1,
			pageSize:    10,
			wantCount:   5,
			wantTotal:   5,
			wantErr:     false,
		},
		{
			name:        "分页第1页",
			userID:      admin.ID,
			workspaceID: workspace.ID,
			page:        1,
			pageSize:    3,
			wantCount:   3,
			wantTotal:   5,
			wantErr:     false,
		},
		{
			name:        "分页第2页",
			userID:      admin.ID,
			workspaceID: workspace.ID,
			page:        2,
			pageSize:    3,
			wantCount:   2,
			wantTotal:   5,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "workspace_id", tt.workspaceID)

			teams, total, err := svc.ListTeams(ctx, tt.workspaceID.String(), tt.page, tt.pageSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListTeams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantTotal, total)
			assert.Len(t, teams, tt.wantCount)
		})
	}
}

// =============================================================================
// GetTeam 测试
// =============================================================================

func TestTeamService_GetTeam(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := NewWorkflowService(workflowStateStore, teamStore)
	svc := NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "GetTeam Test " + prefix,
		Slug: "getteam-test-" + prefix,
	}
	tx.Create(workspace)

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Test Team " + prefix,
		Key:         "GT" + "ABC",
	}
	tx.Create(team)

	// 创建用户
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
	tx.Create(member)
	tx.Create(nonMember)

	// 添加成员
	tx.Create(&model.TeamMember{
		TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now(),
	})

	tests := []struct {
		name     string
		userID   uuid.UUID
		userRole model.Role
		teamID   string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "团队成员访问",
			userID:   member.ID,
			userRole: model.RoleMember,
			teamID:   team.ID.String(),
			wantErr:  false,
		},
		{
			name:     "非团队成员拒绝",
			userID:   nonMember.ID,
			userRole: model.RoleMember,
			teamID:   team.ID.String(),
			wantErr:  true,
			errMsg:   "无权限",
		},
		{
			name:     "团队不存在",
			userID:   member.ID,
			userRole: model.RoleMember,
			teamID:   uuid.New().String(),
			wantErr:  true,
			errMsg:   "团队不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)
			ctx = context.WithValue(ctx, "workspace_id", workspace.ID)

			result, err := svc.GetTeam(ctx, tt.teamID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				assert.NotNil(t, result)
				assert.Equal(t, tt.teamID, result.ID.String())
			}
		})
	}
}

// =============================================================================
// UpdateTeam 测试
// =============================================================================

func TestTeamService_UpdateTeam(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := NewWorkflowService(workflowStateStore, teamStore)
	svc := NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "UpdateTeam Test " + prefix,
		Slug: "updateteam-test-" + prefix,
	}
	tx.Create(workspace)

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Original Team " + prefix,
		Key:         "UT" + "ABC",
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

	// 添加成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name     string
		userID   uuid.UUID
		userRole model.Role
		teamID   string
		updates  map[string]interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Team Owner 更新",
			userID:   owner.ID,
			userRole: model.RoleMember,
			teamID:   team.ID.String(),
			updates:  map[string]interface{}{"name": "Updated by Owner"},
			wantErr:  false,
		},
		{
			name:     "普通成员无权限",
			userID:   member.ID,
			userRole: model.RoleMember,
			teamID:   team.ID.String(),
			updates:  map[string]interface{}{"name": "Should Not Update"},
			wantErr:  true,
			errMsg:   "无权限",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)
			ctx = context.WithValue(ctx, "workspace_id", workspace.ID)

			result, err := svc.UpdateTeam(ctx, tt.teamID, tt.updates)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				assert.NotNil(t, result)
				if name, ok := tt.updates["name"].(string); ok {
					assert.Equal(t, name, result.Name)
				}
			}
		})
	}
}

// =============================================================================
// DeleteTeam 测试
// =============================================================================

func TestTeamService_DeleteTeam(t *testing.T) {
	if testTeamServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testTeamServiceDB.Begin()
	defer tx.Rollback()

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := NewWorkflowService(workflowStateStore, teamStore)
	svc := NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "DeleteTeam Test " + prefix,
		Slug: "deleteteam-test-" + prefix,
	}
	tx.Create(workspace)

	// 创建有空 Issue 的团队
	teamWithIssue := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team With Issue " + prefix,
		Key:         "DI" + "ABC",
	}
	tx.Create(teamWithIssue)

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

	// 创建 Issue
	status := &model.WorkflowState{
		TeamID: teamWithIssue.ID,
		Name:   "Todo",
		Type:   model.StateTypeUnstarted,
		Color:  "#gray",
	}
	tx.Create(status)

	issue := &model.Issue{
		TeamID:      teamWithIssue.ID,
		Number:      1,
		Title:       "Test Issue",
		Priority:    model.PriorityMedium,
		StatusID:    status.ID,
		CreatedByID: owner.ID,
	}
	tx.Create(issue)

	// 添加成员
	tx.Create(&model.TeamMember{TeamID: teamWithIssue.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: teamWithIssue.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	// 创建空团队用于删除测试
	emptyTeam := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Empty Team " + prefix,
		Key:         "ET" + "ABC",
	}
	tx.Create(emptyTeam)
	tx.Create(&model.TeamMember{TeamID: emptyTeam.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})

	tests := []struct {
		name     string
		userID   uuid.UUID
		userRole model.Role
		teamID   string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "存在 Issue 时拒绝删除",
			userID:   owner.ID,
			userRole: model.RoleMember,
			teamID:   teamWithIssue.ID.String(),
			wantErr:  true,
			errMsg:   "Issue",
		},
		{
			name:     "普通成员无权限",
			userID:   member.ID,
			userRole: model.RoleMember,
			teamID:   emptyTeam.ID.String(),
			wantErr:  true,
			errMsg:   "无权限",
		},
		{
			name:     "Team Owner 删除空团队",
			userID:   owner.ID,
			userRole: model.RoleMember,
			teamID:   emptyTeam.ID.String(),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)
			ctx = context.WithValue(ctx, "workspace_id", workspace.ID)

			err := svc.DeleteTeam(ctx, tt.teamID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				// 验证团队已删除
				_, err := teamStore.GetByID(ctx, tt.teamID)
				assert.Error(t, err, "团队应该已被删除")
			}
		})
	}
}
