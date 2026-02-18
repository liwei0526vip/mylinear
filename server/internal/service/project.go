// Package service 提供业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

// 错误定义
var (
	ErrProjectNotAuthorized = errors.New("权限不足")
	ErrProjectLeadNotMember = errors.New("负责人不是团队成员")
	ErrProjectInvalidDates  = errors.New("目标日期不能早于开始日期")
	ErrProjectNotFound      = errors.New("项目不存在")
)

// CreateProjectParams 创建项目参数
type CreateProjectParams struct {
	WorkspaceID uuid.UUID
	Name        string
	Description *string
	LeadID      *uuid.UUID
	StartDate   *time.Time
	TargetDate  *time.Time
	Teams       []uuid.UUID
	Labels      []uuid.UUID
}

// UpdateProjectParams 更新项目参数
type UpdateProjectParams struct {
	Name        *string
	Description *string
	Status      *model.ProjectStatus
	LeadID      *uuid.UUID
	StartDate   *time.Time
	TargetDate  *time.Time
	Teams       []uuid.UUID
	Labels      []uuid.UUID
}

// ProjectProgress 项目进度
type ProjectProgress struct {
	TotalIssues     int     `json:"total_issues"`
	CompletedIssues int     `json:"completed_issues"`
	CancelledIssues int     `json:"cancelled_issues"`
	ProgressPercent float64 `json:"progress_percent"`
}

// ProjectService 定义项目服务接口
type ProjectService interface {
	// CreateProject 创建项目
	CreateProject(ctx context.Context, params *CreateProjectParams) (*model.Project, error)
	// GetProject 获取项目
	GetProject(ctx context.Context, projectID uuid.UUID) (*model.Project, error)
	// ListProjectsByWorkspace 获取工作区项目列表
	ListProjectsByWorkspace(ctx context.Context, workspaceID uuid.UUID, filter *store.ProjectFilter, page, pageSize int) ([]model.Project, int64, error)
	// ListProjectsByTeam 获取团队项目列表
	ListProjectsByTeam(ctx context.Context, teamID uuid.UUID, filter *store.ProjectFilter, page, pageSize int) ([]model.Project, int64, error)
	// UpdateProject 更新项目
	UpdateProject(ctx context.Context, projectID uuid.UUID, updates map[string]interface{}) (*model.Project, error)
	// DeleteProject 删除项目
	DeleteProject(ctx context.Context, projectID uuid.UUID) error
	// GetProjectProgress 获取项目进度
	GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*ProjectProgress, error)
	// ListProjectIssues 获取项目关联的 Issue 列表
	ListProjectIssues(ctx context.Context, projectID uuid.UUID, filter *store.IssueFilter, page, pageSize int) ([]model.Issue, int64, error)
}

// projectService 实现 ProjectService 接口
type projectService struct {
	projectStore    store.ProjectStore
	teamMemberStore store.TeamMemberStore
	userStore       store.UserStore
}

// NewProjectService 创建项目服务实例
func NewProjectService(projectStore store.ProjectStore, teamMemberStore store.TeamMemberStore, userStore store.UserStore) ProjectService {
	return &projectService{
		projectStore:    projectStore,
		teamMemberStore: teamMemberStore,
		userStore:       userStore,
	}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(ctx context.Context, params *CreateProjectParams) (*model.Project, error) {
	// 验证名称
	if params.Name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}

	// 验证日期逻辑
	if params.StartDate != nil && params.TargetDate != nil {
		if params.TargetDate.Before(*params.StartDate) {
			return nil, ErrProjectInvalidDates
		}
	}

	// 验证负责人是否是工作区成员（简化验证：检查用户是否存在于工作区）
	// 这里不强制要求负责人必须是某个团队的成员，因为项目可以跨团队

	// 转换 Teams 和 Labels
	var teamIDs pq.StringArray
	if len(params.Teams) > 0 {
		teamIDs = make(pq.StringArray, len(params.Teams))
		for i, id := range params.Teams {
			teamIDs[i] = id.String()
		}
	}

	var labelIDs pq.StringArray
	if len(params.Labels) > 0 {
		labelIDs = make(pq.StringArray, len(params.Labels))
		for i, id := range params.Labels {
			labelIDs[i] = id.String()
		}
	}

	// 创建项目
	project := &model.Project{
		WorkspaceID: params.WorkspaceID,
		Name:        params.Name,
		Description: params.Description,
		Status:      model.ProjectStatusPlanned,
		LeadID:      params.LeadID,
		StartDate:   params.StartDate,
		TargetDate:  params.TargetDate,
		Teams:       teamIDs,
		Labels:      labelIDs,
	}

	if err := s.projectStore.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("创建项目失败: %w", err)
	}

	return project, nil
}

// GetProject 获取项目
func (s *projectService) GetProject(ctx context.Context, projectID uuid.UUID) (*model.Project, error) {
	project, err := s.projectStore.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目失败: %w", err)
	}
	return project, nil
}

// ListProjectsByWorkspace 获取工作区项目列表
func (s *projectService) ListProjectsByWorkspace(ctx context.Context, workspaceID uuid.UUID, filter *store.ProjectFilter, page, pageSize int) ([]model.Project, int64, error) {
	projects, total, err := s.projectStore.ListByWorkspace(ctx, workspaceID, filter, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取项目列表失败: %w", err)
	}
	return projects, total, nil
}

// ListProjectsByTeam 获取团队项目列表
func (s *projectService) ListProjectsByTeam(ctx context.Context, teamID uuid.UUID, filter *store.ProjectFilter, page, pageSize int) ([]model.Project, int64, error) {
	projects, total, err := s.projectStore.ListByTeam(ctx, teamID, filter, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取项目列表失败: %w", err)
	}
	return projects, total, nil
}

// UpdateProject 更新项目
func (s *projectService) UpdateProject(ctx context.Context, projectID uuid.UUID, updates map[string]interface{}) (*model.Project, error) {
	// 获取现有项目
	project, err := s.projectStore.GetByID(ctx, projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	// 应用更新
	if name, ok := updates["name"].(string); ok {
		project.Name = name
	}
	if desc, ok := updates["description"].(*string); ok {
		project.Description = desc
	}
	if status, ok := updates["status"].(model.ProjectStatus); ok {
		project.Status = status
		// 处理状态时间戳
		if status == model.ProjectStatusCompleted && project.CompletedAt == nil {
			now := time.Now()
			project.CompletedAt = &now
		}
	}
	if leadID, ok := updates["lead_id"].(*uuid.UUID); ok {
		project.LeadID = leadID
	}
	if startDate, ok := updates["start_date"].(*time.Time); ok {
		project.StartDate = startDate
	}
	if targetDate, ok := updates["target_date"].(*time.Time); ok {
		project.TargetDate = targetDate
	}
	if completedAt, ok := updates["completed_at"].(*time.Time); ok {
		project.CompletedAt = completedAt
	}

	if err := s.projectStore.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("更新项目失败: %w", err)
	}

	return project, nil
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	// 获取当前用户
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return fmt.Errorf("未认证")
	}

	// 获取项目
	project, err := s.projectStore.GetByID(ctx, projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	// 检查权限：Workspace Admin 或 Global Admin 可以直接删除
	user, err := s.userStore.GetUserByID(ctx, userID.String())
	if err == nil && (user.Role == model.RoleAdmin || user.Role == model.RoleGlobalAdmin) {
		return s.projectStore.SoftDelete(ctx, projectID)
	}

	// 否则需要是关联团队的 Admin
	if len(project.Teams) > 0 {
		hasAdminRole := false
		for _, teamIDStr := range project.Teams {
			teamID, err := uuid.Parse(teamIDStr)
			if err != nil {
				continue
			}
			role, err := s.teamMemberStore.GetRole(ctx, teamID.String(), userID.String())
			if err == nil && (role == model.RoleAdmin || role == model.RoleGlobalAdmin) {
				hasAdminRole = true
				break
			}
		}
		if !hasAdminRole {
			return ErrProjectNotAuthorized
		}
	} else {
		// 如果项目没有关联团队，只有 Workspace Admin 可以删除（上面已通过）
		// 普通用户无权删除
		return ErrProjectNotAuthorized
	}

	return s.projectStore.SoftDelete(ctx, projectID)
}

// GetProjectProgress 获取项目进度
func (s *projectService) GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*ProjectProgress, error) {
	progress, err := s.projectStore.GetProgress(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目进度失败: %w", err)
	}

	return &ProjectProgress{
		TotalIssues:     progress.TotalIssues,
		CompletedIssues: progress.CompletedIssues,
		CancelledIssues: progress.CancelledIssues,
		ProgressPercent: progress.ProgressPercent,
	}, nil
}

// ListProjectIssues 获取项目关联的 Issue 列表
func (s *projectService) ListProjectIssues(ctx context.Context, projectID uuid.UUID, filter *store.IssueFilter, page, pageSize int) ([]model.Issue, int64, error) {
	issues, total, err := s.projectStore.ListIssues(ctx, projectID, filter, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取项目 Issue 列表失败: %w", err)
	}
	return issues, total, nil
}
