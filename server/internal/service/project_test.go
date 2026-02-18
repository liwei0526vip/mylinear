package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestProjectService_Interface 测试 ProjectService 接口定义存在
func TestProjectService_Interface(t *testing.T) {
	var _ ProjectService = (*projectService)(nil)
}

// =============================================================================
// CreateProject 测试 (Task 3.1)
// =============================================================================

func TestProjectService_CreateProject(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	projectStore := store.NewProjectStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	service := NewProjectService(projectStore, teamMemberStore, userStore)
	ctx := context.Background()

	// 准备测试数据
	workspace, user, _ := setupProjectServiceFixtures(t, tx)

	// 设置用户上下文
	ctx = context.WithValue(ctx, "user_id", user.ID)

	validLeadID := user.ID
	pastDate := time.Now().Add(-24 * time.Hour)
	futureDate := time.Now().Add(24 * time.Hour)
	reverseStartDate := futureDate
	reverseEndDate := pastDate

	tests := []struct {
		name    string
		params  *CreateProjectParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "正常创建项目",
			params: &CreateProjectParams{
				WorkspaceID: workspace.ID,
				Name:        "Test Project",
				Description: strPtr("Test description"),
				LeadID:      &validLeadID,
			},
			wantErr: false,
		},
		{
			name: "日期逻辑验证失败-targetDate早于startDate",
			params: &CreateProjectParams{
				WorkspaceID: workspace.ID,
				Name:        "Invalid Date Project",
				StartDate:   &reverseStartDate,
				TargetDate:  &reverseEndDate,
			},
			wantErr: true,
			errMsg:  "目标日期不能早于开始日期",
		},
		{
			name: "日期逻辑验证成功-正常日期",
			params: &CreateProjectParams{
				WorkspaceID: workspace.ID,
				Name:        "Valid Date Project",
				StartDate:   &pastDate,
				TargetDate:  &futureDate,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := service.CreateProject(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.NotEqual(t, uuid.Nil, project.ID)
				assert.Equal(t, tt.params.Name, project.Name)
			}
		})
	}
}

// =============================================================================
// UpdateProject 测试 (Task 3.3)
// =============================================================================

func TestProjectService_UpdateProject(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	projectStore := store.NewProjectStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	service := NewProjectService(projectStore, teamMemberStore, userStore)
	ctx := context.Background()

	// 准备测试数据
	workspace, _, _ := setupProjectServiceFixtures(t, tx)

	// 创建测试项目
	project := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Original Name",
		Status:      model.ProjectStatusPlanned,
	}
	assert.NoError(t, projectStore.Create(ctx, project))

	now := time.Now()

	tests := []struct {
		name        string
		projectID   uuid.UUID
		updates     map[string]interface{}
		wantErr     bool
		resultCheck func(t *testing.T, p *model.Project)
	}{
		{
			name:      "状态转换到completed设置时间戳",
			projectID: project.ID,
			updates: map[string]interface{}{
				"status": model.ProjectStatusCompleted,
			},
			wantErr: false,
			resultCheck: func(t *testing.T, p *model.Project) {
				assert.Equal(t, model.ProjectStatusCompleted, p.Status)
				assert.NotNil(t, p.CompletedAt)
			},
		},
		{
			name:      "状态转换到cancelled设置时间戳",
			projectID: project.ID,
			updates: map[string]interface{}{
				"status":       model.ProjectStatusCancelled,
				"cancelled_at": &now,
			},
			wantErr: false,
			resultCheck: func(t *testing.T, p *model.Project) {
				assert.Equal(t, model.ProjectStatusCancelled, p.Status)
			},
		},
		{
			name:      "状态从completed切换回in_progress清除completed_at",
			projectID: project.ID,
			updates: map[string]interface{}{
				"status":       model.ProjectStatusInProgress,
				"completed_at": nil,
			},
			wantErr: false,
			resultCheck: func(t *testing.T, p *model.Project) {
				assert.Equal(t, model.ProjectStatusInProgress, p.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := service.UpdateProject(ctx, tt.projectID, tt.updates)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updated)
				if tt.resultCheck != nil {
					// 重新获取以验证持久化
					p, err := projectStore.GetByID(ctx, tt.projectID)
					assert.NoError(t, err)
					tt.resultCheck(t, p)
				}
			}
		})
	}
}

// =============================================================================
// DeleteProject 测试 (Task 3.5)
// =============================================================================

func TestProjectService_DeleteProject(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	projectStore := store.NewProjectStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	service := NewProjectService(projectStore, teamMemberStore, userStore)
	ctx := context.Background()

	// 准备测试数据
	workspace, adminUser, team := setupProjectServiceFixtures(t, tx)

	// 创建普通成员
	memberUser := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        "member-" + uuid.New().String()[:8] + "@example.com",
		Username:     "member" + uuid.New().String()[:8],
		Name:         "Member User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	assert.NoError(t, tx.Create(memberUser).Error)

	// 添加普通成员到团队
	membership := &model.TeamMember{
		TeamID: team.ID,
		UserID: memberUser.ID,
		Role:   model.RoleMember,
	}
	assert.NoError(t, tx.Create(membership).Error)

	// 创建测试项目（关联团队）
	project := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Delete Test Project",
		Status:      model.ProjectStatusPlanned,
		Teams:       pq.StringArray{team.ID.String()},
	}
	assert.NoError(t, projectStore.Create(ctx, project))

	tests := []struct {
		name      string
		projectID uuid.UUID
		userID    uuid.UUID
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "权限校验-非Admin拒绝",
			projectID: project.ID,
			userID:    memberUser.ID,
			wantErr:   true,
			errMsg:    "权限不足",
		},
		{
			name:      "Admin可以删除",
			projectID: project.ID,
			userID:    adminUser.ID,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置用户上下文
			userCtx := context.WithValue(ctx, "user_id", tt.userID)

			err := service.DeleteProject(userCtx, tt.projectID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// 验证项目被软删除
				_, err := projectStore.GetByID(ctx, tt.projectID)
				assert.Error(t, err) // 应该找不到
			}
		})
	}
}

// =============================================================================
// 辅助函数
// =============================================================================

// setupProjectServiceFixtures 创建测试所需的基础数据
func setupProjectServiceFixtures(t *testing.T, tx *gorm.DB) (*model.Workspace, *model.User, *model.Team) {
	workspace := &model.Workspace{
		Name: "Project Service Fixture WS",
		Slug: "project-service-ws-" + uuid.New().String()[:8],
	}
	assert.NoError(t, tx.Create(workspace).Error)

	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        "project-service-" + uuid.New().String()[:8] + "@example.com",
		Username:     "projectservice" + uuid.New().String()[:8],
		Name:         "Project Service User",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	assert.NoError(t, tx.Create(user).Error)

	team := &model.Team{
		WorkspaceID: workspace.ID,
		Key:         "PS" + uuid.New().String()[:6],
		Name:        "Project Service Team",
	}
	assert.NoError(t, tx.Create(team).Error)

	// 添加用户为团队成员（Admin 角色）
	membership := &model.TeamMember{
		TeamID: team.ID,
		UserID: user.ID,
		Role:   model.RoleAdmin,
	}
	assert.NoError(t, tx.Create(membership).Error)

	return workspace, user, team
}
