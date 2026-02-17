// Package service 提供业务逻辑层
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mylinear/server/internal/model"
	"github.com/mylinear/server/internal/store"
)

// TeamMemberService 定义团队成员服务接口
type TeamMemberService interface {
	// ListMembers 获取团队成员列表
	ListMembers(ctx context.Context, teamID string) ([]model.TeamMember, error)
	// AddMember 添加团队成员
	AddMember(ctx context.Context, teamID, userID string, role model.Role) error
	// RemoveMember 移除团队成员
	RemoveMember(ctx context.Context, teamID, userID string) error
	// UpdateRole 更新成员角色
	UpdateRole(ctx context.Context, teamID, userID string, role model.Role) error
}

// teamMemberService 实现 TeamMemberService 接口
type teamMemberService struct {
	teamMemberStore store.TeamMemberStore
	userStore       store.UserStore
	teamStore       store.TeamStore
}

// NewTeamMemberService 创建团队成员服务实例
func NewTeamMemberService(teamMemberStore store.TeamMemberStore, userStore store.UserStore, teamStore store.TeamStore) TeamMemberService {
	return &teamMemberService{
		teamMemberStore: teamMemberStore,
		userStore:       userStore,
		teamStore:       teamStore,
	}
}

// ListMembers 获取团队成员列表
func (s *teamMemberService) ListMembers(ctx context.Context, teamID string) ([]model.TeamMember, error) {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("未认证")
	}

	userRole, _ := ctx.Value("user_role").(model.Role)

	// 权限检查：Admin 绕过，否则需要是团队成员
	if userRole != model.RoleAdmin && userRole != model.RoleGlobalAdmin {
		role, _ := s.teamMemberStore.GetRole(ctx, teamID, userID.String())
		if role == "" {
			return nil, fmt.Errorf("无权限访问此团队")
		}
	}

	return s.teamMemberStore.List(ctx, teamID)
}

// AddMember 添加团队成员
func (s *teamMemberService) AddMember(ctx context.Context, teamID, userID string, role model.Role) error {
	// 获取当前用户信息
	currentUserID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return fmt.Errorf("未认证")
	}

	currentUserRole, _ := ctx.Value("user_role").(model.Role)

	// 权限检查：Admin 绕过，否则需要是 Team Owner
	if currentUserRole != model.RoleAdmin && currentUserRole != model.RoleGlobalAdmin {
		role, _ := s.teamMemberStore.GetRole(ctx, teamID, currentUserID.String())
		if role != model.RoleAdmin {
			return fmt.Errorf("无权限添加成员")
		}
	}

	// 解析目标用户 ID
	targetUserID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID")
	}

	// 解析团队 ID
	teamUUID, err := uuid.Parse(teamID)
	if err != nil {
		return fmt.Errorf("无效的团队ID")
	}

	// 添加成员
	member := &model.TeamMember{
		TeamID:   teamUUID,
		UserID:   targetUserID,
		Role:     role,
		JoinedAt: time.Now(),
	}

	return s.teamMemberStore.Add(ctx, member)
}

// RemoveMember 移除团队成员
func (s *teamMemberService) RemoveMember(ctx context.Context, teamID, userID string) error {
	// 获取当前用户信息
	currentUserID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return fmt.Errorf("未认证")
	}

	currentUserRole, _ := ctx.Value("user_role").(model.Role)

	// 权限检查：Admin 绕过，否则需要是 Team Owner
	if currentUserRole != model.RoleAdmin && currentUserRole != model.RoleGlobalAdmin {
		role, _ := s.teamMemberStore.GetRole(ctx, teamID, currentUserID.String())
		if role != model.RoleAdmin {
			return fmt.Errorf("无权限移除成员")
		}
	}

	return s.teamMemberStore.Remove(ctx, teamID, userID)
}

// UpdateRole 更新成员角色
func (s *teamMemberService) UpdateRole(ctx context.Context, teamID, userID string, role model.Role) error {
	// 获取当前用户信息
	currentUserID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return fmt.Errorf("未认证")
	}

	currentUserRole, _ := ctx.Value("user_role").(model.Role)

	// 权限检查：Admin 绕过，否则需要是 Team Owner
	if currentUserRole != model.RoleAdmin && currentUserRole != model.RoleGlobalAdmin {
		role, _ := s.teamMemberStore.GetRole(ctx, teamID, currentUserID.String())
		if role != model.RoleAdmin {
			return fmt.Errorf("无权限更新角色")
		}
	}

	return s.teamMemberStore.UpdateRole(ctx, teamID, userID, role)
}
