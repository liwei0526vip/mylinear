// Package store 提供数据访问层
package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// TeamStore 定义团队数据访问接口
type TeamStore interface {
	// List 获取团队列表
	List(ctx context.Context, workspaceID string, page, pageSize int) ([]model.Team, int64, error)
	// GetByID 通过 ID 获取团队
	GetByID(ctx context.Context, id string) (*model.Team, error)
	// Create 创建团队
	Create(ctx context.Context, team *model.Team) error
	// Update 更新团队
	Update(ctx context.Context, team *model.Team) error
	// SoftDelete 软删除团队
	SoftDelete(ctx context.Context, id string) error
	// CountIssuesByTeam 统计团队下的 Issue 数量
	CountIssuesByTeam(ctx context.Context, teamID string) (int64, error)
}

// teamStore 实现 TeamStore 接口
type teamStore struct {
	db *gorm.DB
}

// NewTeamStore 创建团队存储实例
func NewTeamStore(db *gorm.DB) TeamStore {
	return &teamStore{db: db}
}

// List 获取团队列表
func (s *teamStore) List(ctx context.Context, workspaceID string, page, pageSize int) ([]model.Team, int64, error) {
	var teams []model.Team
	var total int64

	// 构建查询
	query := s.db.WithContext(ctx).Model(&model.Team{}).Where("workspace_id = ?", workspaceID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计团队数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&teams).Error; err != nil {
		return nil, 0, fmt.Errorf("查询团队列表失败: %w", err)
	}

	return teams, total, nil
}

// GetByID 通过 ID 获取团队
func (s *teamStore) GetByID(ctx context.Context, id string) (*model.Team, error) {
	var team model.Team
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// Create 创建团队
func (s *teamStore) Create(ctx context.Context, team *model.Team) error {
	// 校验 Key 格式
	if err := ValidateTeamKey(team.Key); err != nil {
		return err
	}

	return s.db.WithContext(ctx).Create(team).Error
}

// ValidateTeamKey 校验团队标识符格式
func ValidateTeamKey(key string) error {
	if len(key) < 2 || len(key) > 10 {
		return fmt.Errorf("团队标识符长度必须为 2-10 位")
	}

	// 首字母必须为大写字母
	if key[0] < 'A' || key[0] > 'Z' {
		return fmt.Errorf("团队标识符首字母必须为大写字母")
	}

	// 其余字符必须为大写字母或数字
	for i := 1; i < len(key); i++ {
		c := key[i]
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return fmt.Errorf("团队标识符必须为大写字母和数字")
		}
	}

	return nil
}

// Update 更新团队
func (s *teamStore) Update(ctx context.Context, team *model.Team) error {
	return s.db.WithContext(ctx).Save(team).Error
}

// SoftDelete 软删除团队（使用 gorm.DeletedAt 需要模型支持，这里用硬删除）
func (s *teamStore) SoftDelete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.Team{}, "id = ?", id).Error
}

// CountIssuesByTeam 统计团队下的 Issue 数量
func (s *teamStore) CountIssuesByTeam(ctx context.Context, teamID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Issue{}).Where("team_id = ?", teamID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计 Issue 数量失败: %w", err)
	}
	return count, nil
}

// TeamMemberStore 定义团队成员数据访问接口
type TeamMemberStore interface {
	// List 获取团队成员列表
	List(ctx context.Context, teamID string) ([]model.TeamMember, error)
	// Add 添加团队成员
	Add(ctx context.Context, member *model.TeamMember) error
	// Remove 移除团队成员
	Remove(ctx context.Context, teamID, userID string) error
	// UpdateRole 更新成员角色
	UpdateRole(ctx context.Context, teamID, userID string, role model.Role) error
	// GetRole 获取成员角色
	GetRole(ctx context.Context, teamID, userID string) (model.Role, error)
}

// teamMemberStore 实现 TeamMemberStore 接口
type teamMemberStore struct {
	db *gorm.DB
}

// NewTeamMemberStore 创建团队成员存储实例
func NewTeamMemberStore(db *gorm.DB) TeamMemberStore {
	return &teamMemberStore{db: db}
}

// List 获取团队成员列表
func (s *teamMemberStore) List(ctx context.Context, teamID string) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := s.db.WithContext(ctx).
		Preload("User").
		Where("team_id = ?", teamID).
		Order("joined_at ASC").
		Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("查询团队成员失败: %w", err)
	}
	return members, nil
}

// Add 添加团队成员
func (s *teamMemberStore) Add(ctx context.Context, member *model.TeamMember) error {
	// 检查是否已是成员
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&model.TeamMember{}).
		Where("team_id = ? AND user_id = ?", member.TeamID, member.UserID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("检查成员状态失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("用户已是团队成员")
	}

	return s.db.WithContext(ctx).Create(member).Error
}

// Remove 移除团队成员
func (s *teamMemberStore) Remove(ctx context.Context, teamID, userID string) error {
	// 检查该成员是否是 Owner
	var member model.TeamMember
	err := s.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if err != nil {
		// 成员不存在，直接返回（不报错）
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("查询成员失败: %w", err)
	}

	// 如果是 Owner，检查是否是最后一个
	if member.Role == model.RoleAdmin {
		var ownerCount int64
		if err := s.db.WithContext(ctx).
			Model(&model.TeamMember{}).
			Where("team_id = ? AND role = ?", teamID, model.RoleAdmin).
			Count(&ownerCount).Error; err != nil {
			return fmt.Errorf("统计 Owner 数量失败: %w", err)
		}
		if ownerCount <= 1 {
			return fmt.Errorf("无法移除最后一个 Owner")
		}
	}

	return s.db.WithContext(ctx).
		Delete(&model.TeamMember{}, "team_id = ? AND user_id = ?", teamID, userID).Error
}

// UpdateRole 更新成员角色
func (s *teamMemberStore) UpdateRole(ctx context.Context, teamID, userID string, role model.Role) error {
	// 检查当前角色
	var member model.TeamMember
	err := s.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if err != nil {
		return fmt.Errorf("查询成员失败: %w", err)
	}

	// 如果是从 Owner 降级为 Member，检查是否是最后一个 Owner
	if member.Role == model.RoleAdmin && role != model.RoleAdmin {
		var ownerCount int64
		if err := s.db.WithContext(ctx).
			Model(&model.TeamMember{}).
			Where("team_id = ? AND role = ?", teamID, model.RoleAdmin).
			Count(&ownerCount).Error; err != nil {
			return fmt.Errorf("统计 Owner 数量失败: %w", err)
		}
		if ownerCount <= 1 {
			return fmt.Errorf("无法降级最后一个 Owner")
		}
	}

	return s.db.WithContext(ctx).
		Model(&model.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("role", role).Error
}

// GetRole 获取成员角色
func (s *teamMemberStore) GetRole(ctx context.Context, teamID, userID string) (model.Role, error) {
	var member model.TeamMember
	err := s.db.WithContext(ctx).
		Select("role").
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // 非成员返回空角色
		}
		return "", fmt.Errorf("查询成员角色失败: %w", err)
	}
	return member.Role, nil
}

// GetTeamRole 获取用户在团队中的角色（辅助函数）
func GetTeamRole(ctx context.Context, db *gorm.DB, userID, teamID uuid.UUID) (model.Role, error) {
	var member model.TeamMember
	err := db.WithContext(ctx).
		Select("role").
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // 非成员返回空角色
		}
		return "", fmt.Errorf("查询团队成员角色失败: %w", err)
	}
	return member.Role, nil
}

// IsTeamOwner 检查用户是否是团队 Owner
func IsTeamOwner(ctx context.Context, db *gorm.DB, userID, teamID uuid.UUID) (bool, error) {
	role, err := GetTeamRole(ctx, db, userID, teamID)
	if err != nil {
		return false, err
	}
	return role == model.RoleAdmin, nil
}

// IsTeamMember 检查用户是否是团队成员
func IsTeamMember(ctx context.Context, db *gorm.DB, userID, teamID uuid.UUID) (bool, error) {
	role, err := GetTeamRole(ctx, db, userID, teamID)
	if err != nil {
		return false, err
	}
	return role != "", nil
}
