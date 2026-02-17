// Package service 提供业务逻辑层
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mylinear/server/internal/model"
	"github.com/mylinear/server/internal/store"
)

// WorkspaceService 定义工作区服务接口
type WorkspaceService interface {
	// GetWorkspace 获取工作区信息
	GetWorkspace(ctx context.Context, workspaceID string) (*model.Workspace, error)
	// UpdateWorkspace 更新工作区信息
	UpdateWorkspace(ctx context.Context, workspaceID string, updates map[string]interface{}) (*model.Workspace, error)
	// GetWorkspaceStats 获取工作区统计信息
	GetWorkspaceStats(ctx context.Context, workspaceID string) (*store.WorkspaceStats, error)
}

// workspaceService 实现 WorkspaceService 接口
type workspaceService struct {
	workspaceStore store.WorkspaceStore
	userStore      store.UserStore
}

// NewWorkspaceService 创建工作区服务实例
func NewWorkspaceService(workspaceStore store.WorkspaceStore, userStore store.UserStore) WorkspaceService {
	return &workspaceService{
		workspaceStore: workspaceStore,
		userStore:      userStore,
	}
}

// GetWorkspace 获取工作区信息
func (s *workspaceService) GetWorkspace(ctx context.Context, workspaceID string) (*model.Workspace, error) {
	// 获取当前用户 ID
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("未认证")
	}

	// 获取工作区
	workspace, err := s.workspaceStore.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("工作区不存在")
	}

	// 检查用户是否是工作区成员
	user, err := s.userStore.GetUserByID(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	if user.WorkspaceID.String() != workspaceID {
		return nil, fmt.Errorf("无权限访问此工作区")
	}

	return workspace, nil
}

// UpdateWorkspace 更新工作区信息
func (s *workspaceService) UpdateWorkspace(ctx context.Context, workspaceID string, updates map[string]interface{}) (*model.Workspace, error) {
	// 获取当前用户信息
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("未认证")
	}

	userRole, _ := ctx.Value("user_role").(model.Role)

	// 获取工作区
	workspace, err := s.workspaceStore.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("工作区不存在")
	}

	// 检查用户是否是工作区成员
	user, err := s.userStore.GetUserByID(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	if user.WorkspaceID.String() != workspaceID {
		return nil, fmt.Errorf("无权限访问此工作区")
	}

	// 只有 Admin 可以更新工作区
	if userRole != model.RoleAdmin && userRole != model.RoleGlobalAdmin {
		return nil, fmt.Errorf("无权限更新工作区")
	}

	// 应用更新
	if name, ok := updates["name"].(string); ok {
		workspace.Name = name
	}
	if logoURL, ok := updates["logo_url"].(string); ok {
		workspace.LogoURL = &logoURL
	}

	// 保存更新
	if err := s.workspaceStore.Update(ctx, workspace); err != nil {
		return nil, fmt.Errorf("更新工作区失败: %w", err)
	}

	return workspace, nil
}

// GetWorkspaceStats 获取工作区统计信息
func (s *workspaceService) GetWorkspaceStats(ctx context.Context, workspaceID string) (*store.WorkspaceStats, error) {
	// 获取当前用户 ID
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("未认证")
	}

	// 获取工作区
	_, err := s.workspaceStore.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("工作区不存在")
	}

	// 检查用户是否是工作区成员
	user, err := s.userStore.GetUserByID(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	if user.WorkspaceID.String() != workspaceID {
		return nil, fmt.Errorf("无权限访问此工作区")
	}

	// 获取统计信息
	return s.workspaceStore.GetStats(ctx, workspaceID)
}
