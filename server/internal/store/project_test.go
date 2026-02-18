package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestProjectStore_Interface 测试 ProjectStore 接口定义存在
func TestProjectStore_Interface(t *testing.T) {
	var _ ProjectStore = (*projectStore)(nil)
}

// =============================================================================
// Create 测试 (Task 1.1)
// =============================================================================

func TestProjectStore_Create(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewProjectStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, user, _ := setupProjectTestFixtures(t, tx)

	longDescription := strings.Repeat("a", 10001) // 超长描述

	tests := []struct {
		name    string
		project *model.Project
		wantErr bool
		errMsg  string
	}{
		{
			name: "正常创建项目",
			project: &model.Project{
				WorkspaceID: workspace.ID,
				Name:        "Test Project 1",
				Description: strPtr("This is a test project"),
				Status:      model.ProjectStatusPlanned,
				LeadID:      &user.ID,
			},
			wantErr: false,
		},
		{
			name: "创建不带描述的项目",
			project: &model.Project{
				WorkspaceID: workspace.ID,
				Name:        "Test Project 2",
				Status:      model.ProjectStatusPlanned,
			},
			wantErr: false,
		},
		{
			name: "创建带日期的项目",
			project: &model.Project{
				WorkspaceID: workspace.ID,
				Name:        "Test Project with Dates",
				Status:      model.ProjectStatusPlanned,
				StartDate:   ptrTime(time.Now()),
				TargetDate:  ptrTime(time.Now().Add(30 * 24 * time.Hour)),
			},
			wantErr: false,
		},
		{
			name: "名称为空",
			project: &model.Project{
				WorkspaceID: workspace.ID,
				Name:        "",
				Status:      model.ProjectStatusPlanned,
			},
			wantErr: true,
			errMsg:  "名称不能为空",
		},
		{
			name: "描述超长",
			project: &model.Project{
				WorkspaceID: workspace.ID,
				Name:        "Test Project Long Desc",
				Description: &longDescription,
				Status:      model.ProjectStatusPlanned,
			},
			wantErr: true,
			errMsg:  "描述长度不能超过10000字符",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Create(ctx, tt.project)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.project.ID, "ID 应该被生成")
				assert.NotZero(t, tt.project.CreatedAt, "CreatedAt 应该被设置")

				// 验证可以查询到
				var found model.Project
				err = tx.Where("id = ?", tt.project.ID).First(&found).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.project.Name, found.Name)
			}
		})
	}

	_ = user // 避免未使用警告
}

// =============================================================================
// GetByID 测试 (Task 1.3)
// =============================================================================

func TestProjectStore_GetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewProjectStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, user, _ := setupProjectTestFixtures(t, tx)

	// 创建测试项目
	project := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "GetByID Test Project",
		Description: strPtr("Test description"),
		Status:      model.ProjectStatusInProgress,
		LeadID:      &user.ID,
	}
	assert.NoError(t, store.Create(ctx, project))

	// 创建后软删除的项目
	deletedProject := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Deleted Project",
		Status:      model.ProjectStatusPlanned,
	}
	assert.NoError(t, store.Create(ctx, deletedProject))
	assert.NoError(t, store.SoftDelete(ctx, deletedProject.ID))

	tests := []struct {
		name        string
		id          uuid.UUID
		wantErr     bool
		errCheck    func(t *testing.T, err error)
		resultCheck func(t *testing.T, p *model.Project)
	}{
		{
			name:    "正常获取",
			id:      project.ID,
			wantErr: false,
			resultCheck: func(t *testing.T, p *model.Project) {
				assert.Equal(t, "GetByID Test Project", p.Name)
				assert.NotNil(t, p.Workspace, "应该预加载 Workspace")
				assert.NotNil(t, p.Lead, "应该预加载 Lead")
			},
		},
		{
			name:    "不存在的项目",
			id:      uuid.New(),
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			},
		},
		{
			name:    "已删除的项目",
			id:      deletedProject.ID,
			wantErr: true,
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
// Update 测试 (Task 1.5)
// =============================================================================

func TestProjectStore_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewProjectStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, user, _ := setupProjectTestFixtures(t, tx)

	// 创建测试项目
	project := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Original Name",
		Description: strPtr("Original description"),
		Status:      model.ProjectStatusPlanned,
	}
	assert.NoError(t, store.Create(ctx, project))

	now := time.Now()
	later := now.Add(30 * 24 * time.Hour)

	tests := []struct {
		name        string
		updateFn    func(*model.Project)
		wantErr     bool
		resultCheck func(t *testing.T, updated *model.Project)
	}{
		{
			name: "更新名称",
			updateFn: func(p *model.Project) {
				p.Name = "Updated Name"
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Project) {
				assert.Equal(t, "Updated Name", updated.Name)
			},
		},
		{
			name: "更新状态到 in_progress",
			updateFn: func(p *model.Project) {
				p.Status = model.ProjectStatusInProgress
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Project) {
				assert.Equal(t, model.ProjectStatusInProgress, updated.Status)
			},
		},
		{
			name: "更新状态到 completed 并设置 completed_at",
			updateFn: func(p *model.Project) {
				p.Status = model.ProjectStatusCompleted
				p.CompletedAt = &now
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Project) {
				assert.Equal(t, model.ProjectStatusCompleted, updated.Status)
				assert.NotNil(t, updated.CompletedAt)
			},
		},
		{
			name: "更新负责人",
			updateFn: func(p *model.Project) {
				p.LeadID = &user.ID
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Project) {
				assert.NotNil(t, updated.LeadID)
				assert.Equal(t, user.ID, *updated.LeadID)
			},
		},
		{
			name: "更新日期",
			updateFn: func(p *model.Project) {
				p.StartDate = &now
				p.TargetDate = &later
			},
			wantErr: false,
			resultCheck: func(t *testing.T, updated *model.Project) {
				assert.NotNil(t, updated.StartDate)
				assert.NotNil(t, updated.TargetDate)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 应用更新
			tt.updateFn(project)

			// 执行更新
			err := store.Update(ctx, project)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// 重新查询验证
			updated, err := store.GetByID(ctx, project.ID)
			assert.NoError(t, err)
			if tt.resultCheck != nil {
				tt.resultCheck(t, updated)
			}
		})
	}
}

// =============================================================================
// SoftDelete & Restore 测试 (Task 1.7)
// =============================================================================

func TestProjectStore_SoftDelete(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewProjectStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, _, _ := setupProjectTestFixtures(t, tx)

	// 创建测试项目
	project := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Delete Test Project",
		Status:      model.ProjectStatusPlanned,
	}
	assert.NoError(t, store.Create(ctx, project))

	// 删除项目
	err := store.SoftDelete(ctx, project.ID)
	assert.NoError(t, err)

	// 验证无法通过正常查询获取
	_, err = store.GetByID(ctx, project.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// 验证数据库中仍存在（软删除）
	var count int64
	tx.Unscoped().Model(&model.Project{}).Where("id = ?", project.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestProjectStore_Restore(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewProjectStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, _, _ := setupProjectTestFixtures(t, tx)

	// 创建并删除项目
	project := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Restore Test Project",
		Status:      model.ProjectStatusPlanned,
	}
	assert.NoError(t, store.Create(ctx, project))
	assert.NoError(t, store.SoftDelete(ctx, project.ID))

	// 恢复项目
	err := store.Restore(ctx, project.ID)
	assert.NoError(t, err)

	// 验证可以重新获取
	restored, err := store.GetByID(ctx, project.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Restore Test Project", restored.Name)
}

// =============================================================================
// 辅助函数
// =============================================================================

// setupProjectTestFixtures 创建测试所需的基础数据
func setupProjectTestFixtures(t *testing.T, tx *gorm.DB) (*model.Workspace, *model.User, *model.Team) {
	workspace := &model.Workspace{
		Name: "Project Fixture WS",
		Slug: "project-fixture-ws-" + uuid.New().String()[:8],
	}
	assert.NoError(t, tx.Create(workspace).Error)

	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        "project-fixture-" + uuid.New().String()[:8] + "@example.com",
		Username:     "projectfixture" + uuid.New().String()[:8],
		Name:         "Project Fixture User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	assert.NoError(t, tx.Create(user).Error)

	team := &model.Team{
		WorkspaceID: workspace.ID,
		Key:         "PX" + uuid.New().String()[:6],
		Name:        "Project Fixture Team",
	}
	assert.NoError(t, tx.Create(team).Error)

	return workspace, user, team
}

// ptrTime 返回时间的指针
func ptrTime(t time.Time) *time.Time {
	return &t
}

// =============================================================================
// ListByWorkspace 测试 (Task 2.1)
// =============================================================================

func TestProjectStore_ListByWorkspace(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewProjectStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, _, _ := setupProjectTestFixtures(t, tx)

	// 创建多个项目
	for i := 1; i <= 5; i++ {
		project := &model.Project{
			WorkspaceID: workspace.ID,
			Name:        "List Test Project",
			Status:      model.ProjectStatusPlanned,
		}
		assert.NoError(t, store.Create(ctx, project))
	}

	tests := []struct {
		name      string
		page      int
		pageSize  int
		wantCount int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "空列表-无项目的工作区",
			page:      1,
			pageSize:  10,
			wantCount: 0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:      "多条记录",
			page:      1,
			pageSize:  10,
			wantCount: 5,
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:      "分页-第一页",
			page:      1,
			pageSize:  2,
			wantCount: 2,
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:      "分页-最后一页",
			page:      3,
			pageSize:  2,
			wantCount: 1,
			wantTotal: 5,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为"空列表"测试创建新的空工作区
			var wsID uuid.UUID
			if tt.name == "空列表-无项目的工作区" {
				emptyWS := &model.Workspace{
					Name: "Empty WS",
					Slug: "empty-ws-" + uuid.New().String()[:8],
				}
				assert.NoError(t, tx.Create(emptyWS).Error)
				wsID = emptyWS.ID
			} else {
				wsID = workspace.ID
			}

			projects, total, err := store.ListByWorkspace(ctx, wsID, nil, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Len(t, projects, tt.wantCount)
			}
		})
	}
}

// =============================================================================
// ListByTeam 测试 (Task 2.3)
// =============================================================================

func TestProjectStore_ListByTeam(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewProjectStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, _, team := setupProjectTestFixtures(t, tx)

	// 创建关联团队的项目
	plannedStatus := model.ProjectStatusPlanned
	inProgressStatus := model.ProjectStatusInProgress

	project1 := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Team Project 1",
		Status:      model.ProjectStatusPlanned,
		Teams:       pq.StringArray{team.ID.String()},
	}
	assert.NoError(t, store.Create(ctx, project1))

	project2 := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Team Project 2",
		Status:      model.ProjectStatusInProgress,
		Teams:       pq.StringArray{team.ID.String()},
	}
	assert.NoError(t, store.Create(ctx, project2))

	// 创建不关联团队的项目
	project3 := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Non-Team Project",
		Status:      model.ProjectStatusPlanned,
		Teams:       pq.StringArray{},
	}
	assert.NoError(t, store.Create(ctx, project3))

	tests := []struct {
		name      string
		teamID    uuid.UUID
		filter    *ProjectFilter
		page      int
		pageSize  int
		wantCount int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "按团队过滤-无过滤条件",
			teamID:    team.ID,
			filter:    nil,
			page:      1,
			pageSize:  10,
			wantCount: 2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "按团队和状态过滤",
			teamID:    team.ID,
			filter:    &ProjectFilter{Status: &plannedStatus},
			page:      1,
			pageSize:  10,
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "按团队和in_progress状态过滤",
			teamID:    team.ID,
			filter:    &ProjectFilter{Status: &inProgressStatus},
			page:      1,
			pageSize:  10,
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, total, err := store.ListByTeam(ctx, tt.teamID, tt.filter, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Len(t, projects, tt.wantCount)
			}
		})
	}

	_ = project3 // 避免未使用警告
}

// =============================================================================
// GetProgress 测试 (Task 2.5)
// =============================================================================

func TestProjectStore_GetProgress(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	projectStore := NewProjectStore(tx)
	issueStore := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, user, team := setupProjectTestFixtures(t, tx)

	// 创建工作流状态
	backlogState := &model.WorkflowState{
		TeamID:    team.ID,
		Name:      "Backlog",
		Type:      model.StateTypeBacklog,
		Position:  1000,
		IsDefault: true,
	}
	assert.NoError(t, tx.Create(backlogState).Error)

	completedState := &model.WorkflowState{
		TeamID:    team.ID,
		Name:      "Done",
		Type:      model.StateTypeCompleted,
		Position:  4000,
		IsDefault: false,
	}
	assert.NoError(t, tx.Create(completedState).Error)

	canceledState := &model.WorkflowState{
		TeamID:    team.ID,
		Name:      "Canceled",
		Type:      model.StateTypeCanceled,
		Position:  5000,
		IsDefault: false,
	}
	assert.NoError(t, tx.Create(canceledState).Error)

	tests := []struct {
		name         string
		setupFn      func(projectID uuid.UUID)
		wantTotal    int
		wantComplete int
		wantCancel   int
		wantPercent  float64
	}{
		{
			name:         "无Issue",
			setupFn:      func(projectID uuid.UUID) {},
			wantTotal:    0,
			wantComplete: 0,
			wantCancel:   0,
			wantPercent:  0,
		},
		{
			name: "部分完成",
			setupFn: func(projectID uuid.UUID) {
				for i := 0; i < 3; i++ {
					issue := &model.Issue{
						TeamID:      team.ID,
						ProjectID:   &projectID,
						Title:       "Partial Issue",
						StatusID:    backlogState.ID,
						CreatedByID: user.ID,
					}
					assert.NoError(t, issueStore.Create(ctx, issue))
				}
				for i := 0; i < 2; i++ {
					issue := &model.Issue{
						TeamID:      team.ID,
						ProjectID:   &projectID,
						Title:       "Completed Issue",
						StatusID:    completedState.ID,
						CreatedByID: user.ID,
					}
					assert.NoError(t, issueStore.Create(ctx, issue))
				}
			},
			wantTotal:    5,
			wantComplete: 2,
			wantCancel:   0,
			wantPercent:  40.0, // 2/5 * 100
		},
		{
			name: "全部完成",
			setupFn: func(projectID uuid.UUID) {
				for i := 0; i < 4; i++ {
					issue := &model.Issue{
						TeamID:      team.ID,
						ProjectID:   &projectID,
						Title:       "All Completed",
						StatusID:    completedState.ID,
						CreatedByID: user.ID,
					}
					assert.NoError(t, issueStore.Create(ctx, issue))
				}
			},
			wantTotal:    4,
			wantComplete: 4,
			wantCancel:   0,
			wantPercent:  100.0,
		},
		{
			name: "有取消Issue",
			setupFn: func(projectID uuid.UUID) {
				// 2 completed, 1 canceled, 2 backlog
				for i := 0; i < 2; i++ {
					issue := &model.Issue{
						TeamID:      team.ID,
						ProjectID:   &projectID,
						Title:       "Completed",
						StatusID:    completedState.ID,
						CreatedByID: user.ID,
					}
					assert.NoError(t, issueStore.Create(ctx, issue))
				}
				issue := &model.Issue{
					TeamID:      team.ID,
					ProjectID:   &projectID,
					Title:       "Canceled",
					StatusID:    canceledState.ID,
					CreatedByID: user.ID,
				}
				assert.NoError(t, issueStore.Create(ctx, issue))
				for i := 0; i < 2; i++ {
					issue := &model.Issue{
						TeamID:      team.ID,
						ProjectID:   &projectID,
						Title:       "Backlog",
						StatusID:    backlogState.ID,
						CreatedByID: user.ID,
					}
					assert.NoError(t, issueStore.Create(ctx, issue))
				}
			},
			wantTotal:    5,
			wantComplete: 2,
			wantCancel:   1,
			wantPercent:  50.0, // 2/(5-1) * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建新项目
			project := &model.Project{
				WorkspaceID: workspace.ID,
				Name:        "Progress Test Project",
				Status:      model.ProjectStatusInProgress,
			}
			assert.NoError(t, projectStore.Create(ctx, project))

			// 设置测试数据
			tt.setupFn(project.ID)

			// 获取进度
			progress, err := projectStore.GetProgress(ctx, project.ID)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantTotal, progress.TotalIssues)
			assert.Equal(t, tt.wantComplete, progress.CompletedIssues)
			assert.Equal(t, tt.wantCancel, progress.CancelledIssues)
			assert.InDelta(t, tt.wantPercent, progress.ProgressPercent, 0.01)
		})
	}
}

// =============================================================================
// ListIssues 测试 (Task 2.7)
// =============================================================================

func TestProjectStore_ListIssues(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	projectStore := NewProjectStore(tx)
	issueStore := NewIssueStore(tx)
	ctx := context.Background()

	// 准备测试数据
	workspace, user, team := setupProjectTestFixtures(t, tx)

	// 创建工作流状态
	state := &model.WorkflowState{
		TeamID:    team.ID,
		Name:      "Backlog",
		Type:      model.StateTypeBacklog,
		Position:  1000,
		IsDefault: true,
	}
	assert.NoError(t, tx.Create(state).Error)

	state2 := &model.WorkflowState{
		TeamID:    team.ID,
		Name:      "Done",
		Type:      model.StateTypeCompleted,
		Position:  4000,
		IsDefault: false,
	}
	assert.NoError(t, tx.Create(state2).Error)

	// 创建项目
	project := &model.Project{
		WorkspaceID: workspace.ID,
		Name:        "Issue List Test Project",
		Status:      model.ProjectStatusInProgress,
	}
	assert.NoError(t, projectStore.Create(ctx, project))

	// 创建项目的 Issue
	for i := 0; i < 5; i++ {
		issue := &model.Issue{
			TeamID:      team.ID,
			ProjectID:   &project.ID,
			Title:       "Project Issue",
			StatusID:    state.ID,
			Priority:    model.PriorityMedium,
			CreatedByID: user.ID,
		}
		assert.NoError(t, issueStore.Create(ctx, issue))
	}

	// 创建不属于该项目的 Issue
	otherIssue := &model.Issue{
		TeamID:      team.ID,
		ProjectID:   nil,
		Title:       "Other Issue",
		StatusID:    state.ID,
		CreatedByID: user.ID,
	}
	assert.NoError(t, issueStore.Create(ctx, otherIssue))

	tests := []struct {
		name      string
		projectID uuid.UUID
		filter    *IssueFilter
		page      int
		pageSize  int
		wantCount int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "空列表-项目无Issue",
			projectID: uuid.New(),
			filter:    nil,
			page:      1,
			pageSize:  10,
			wantCount: 0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:      "多条记录",
			projectID: project.ID,
			filter:    nil,
			page:      1,
			pageSize:  10,
			wantCount: 5,
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:      "按状态过滤",
			projectID: project.ID,
			filter:    &IssueFilter{StatusID: &state.ID},
			page:      1,
			pageSize:  10,
			wantCount: 5,
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:      "分页",
			projectID: project.ID,
			filter:    nil,
			page:      1,
			pageSize:  2,
			wantCount: 2,
			wantTotal: 5,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues, total, err := projectStore.ListIssues(ctx, tt.projectID, tt.filter, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Len(t, issues, tt.wantCount)
			}
		})
	}
}
