// Package service 提供业务逻辑层
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

// TeamService 定义团队服务接口
type TeamService interface {
	// CreateTeam 创建团队
	CreateTeam(ctx context.Context, name, key, description string) (*model.Team, error)
	// ListTeams 获取团队列表
	ListTeams(ctx context.Context, workspaceID string, page, pageSize int) ([]model.Team, int64, error)
	// GetTeam 获取团队信息
	GetTeam(ctx context.Context, teamID string) (*model.Team, error)
	// UpdateTeam 更新团队信息
	UpdateTeam(ctx context.Context, teamID string, updates map[string]interface{}) (*model.Team, error)
	// DeleteTeam 删除团队
	DeleteTeam(ctx context.Context, teamID string) error
}

// teamService 实现 TeamService 接口
type teamService struct {
	teamStore       store.TeamStore
	teamMemberStore store.TeamMemberStore
	userStore       store.UserStore
	workflowService WorkflowService
}

// NewTeamService 创建团队服务实例
func NewTeamService(teamStore store.TeamStore, teamMemberStore store.TeamMemberStore, userStore store.UserStore, workflowService WorkflowService) TeamService {
	return &teamService{
		teamStore:       teamStore,
		teamMemberStore: teamMemberStore,
		userStore:       userStore,
		workflowService: workflowService,
	}
}

// CreateTeam 创建团队
func (s *teamService) CreateTeam(ctx context.Context, name, key, description string) (*model.Team, error) {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("未认证")
	}

	userRole, _ := ctx.Value("user_role").(model.Role)
	workspaceID, ok := ctx.Value("workspace_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("缺少工作区信息")
	}

	// 只有 Admin 可以创建团队
	if userRole != model.RoleAdmin && userRole != model.RoleGlobalAdmin {
		return nil, fmt.Errorf("无权限创建团队")
	}

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspaceID,
		Name:        name,
		Key:         key,
		Description: description,
	}

	if err := s.teamStore.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("创建团队失败: %w", err)
	}

	// 创建者成为团队 Owner
	member := &model.TeamMember{
		TeamID:   team.ID,
		UserID:   userID,
		Role:     model.RoleAdmin,
		JoinedAt: time.Now(),
	}

	if err := s.teamMemberStore.Add(ctx, member); err != nil {
		return nil, fmt.Errorf("添加创建者为 Owner 失败: %w", err)
	}

	// 初始化默认工作流状态
	// 这里显式指定 Position 且不再重复查询以防竞争
	defaultStates := []struct {
		Name     string
		Type     model.StateType
		Color    string
		Position float64
	}{
		{"Backlog", model.StateTypeBacklog, "#bec2c8", 1000},
		{"Todo", model.StateTypeUnstarted, "#e2e2e2", 2000},
		{"进行中", model.StateTypeStarted, "#f2c94c", 3000},
		{"已完成", model.StateTypeCompleted, "#5e6ad2", 4000},
		{"已取消", model.StateTypeCanceled, "#9aa5b1", 5000},
	}

	for _, ds := range defaultStates {
		_, err := s.workflowService.CreateState(ctx, &CreateStateParams{
			TeamID:   team.ID,
			Name:     ds.Name,
			Type:     ds.Type,
			Color:    ds.Color,
			Position: ds.Position,
		})
		if err != nil {
			// 如果初始化失败，记录但不中断（或者根据策略返回错误）
			// 这里打印一个日志，如果在集成测试中能看到。
			// 我们返回错误以保证一致性。
			return nil, fmt.Errorf("为新团队创建状态 [%s] 失败: %w", ds.Name, err)
		}
	}

	return team, nil
}

// ListTeams 获取团队列表
func (s *teamService) ListTeams(ctx context.Context, workspaceID string, page, pageSize int) ([]model.Team, int64, error) {
	return s.teamStore.List(ctx, workspaceID, page, pageSize)
}

// GetTeam 获取团队信息
func (s *teamService) GetTeam(ctx context.Context, teamID string) (*model.Team, error) {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("未认证")
	}

	userRole, _ := ctx.Value("user_role").(model.Role)

	// 获取团队
	team, err := s.teamStore.GetByID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("团队不存在")
	}

	// Workspace Admin 可以访问所有团队
	if userRole == model.RoleAdmin || userRole == model.RoleGlobalAdmin {
		return team, nil
	}

	// 检查用户是否是团队成员
	role, _ := s.teamMemberStore.GetRole(ctx, teamID, userID.String())
	if role == "" {
		return nil, fmt.Errorf("无权限访问此团队")
	}

	return team, nil
}

// UpdateTeam 更新团队信息
func (s *teamService) UpdateTeam(ctx context.Context, teamID string, updates map[string]interface{}) (*model.Team, error) {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("未认证")
	}

	userRole, _ := ctx.Value("user_role").(model.Role)

	// 获取团队
	team, err := s.teamStore.GetByID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("团队不存在")
	}

	// 权限检查：Admin 绕过，否则需要是 Team Owner
	if userRole != model.RoleAdmin && userRole != model.RoleGlobalAdmin {
		isOwner, err := s.teamMemberStore.GetRole(ctx, teamID, userID.String())
		if err != nil || isOwner != model.RoleAdmin {
			return nil, fmt.Errorf("无权限更新此团队")
		}
	}

	// 应用更新
	if name, ok := updates["name"].(string); ok {
		team.Name = name
	}
	if key, ok := updates["key"].(string); ok {
		team.Key = key
	}
	if description, ok := updates["description"].(string); ok {
		team.Description = description
	}

	// 保存更新
	if err := s.teamStore.Update(ctx, team); err != nil {
		return nil, fmt.Errorf("更新团队失败: %w", err)
	}

	return team, nil
}

// DeleteTeam 删除团队
func (s *teamService) DeleteTeam(ctx context.Context, teamID string) error {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return fmt.Errorf("未认证")
	}

	userRole, _ := ctx.Value("user_role").(model.Role)

	// 获取团队（验证存在）
	_, err := s.teamStore.GetByID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("团队不存在")
	}

	// 权限检查：Admin 绕过，否则需要是 Team Owner
	if userRole != model.RoleAdmin && userRole != model.RoleGlobalAdmin {
		isOwner, err := s.teamMemberStore.GetRole(ctx, teamID, userID.String())
		if err != nil || isOwner != model.RoleAdmin {
			return fmt.Errorf("无权限删除此团队")
		}
	}

	// 检查是否存在 Issue
	count, err := s.teamStore.CountIssuesByTeam(ctx, teamID)
	if err != nil {
		return fmt.Errorf("检查 Issue 数量失败: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("团队存在 %d 个 Issue，无法删除", count)
	}

	// 删除团队
	return s.teamStore.SoftDelete(ctx, teamID)
}
