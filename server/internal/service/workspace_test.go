package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testWorkspaceServiceDB *gorm.DB

func init() {
	// 尝试连接数据库
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("警告: 无法连接测试数据库: %v\n", err)
		return
	}

	testWorkspaceServiceDB = db
}

// =============================================================================
// GetWorkspace 测试
// =============================================================================

func TestWorkspaceService_GetWorkspace(t *testing.T) {
	if testWorkspaceServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceServiceDB.Begin()
	defer tx.Rollback()

	workspaceStore := store.NewWorkspaceStore(tx)
	userStore := store.NewUserStore(tx)
	svc := NewWorkspaceService(workspaceStore, userStore)

	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "GetWorkspace Test " + prefix,
		Slug: "getworkspace-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试用户（工作区成员）
	member := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_member@example.com",
		Username:     prefix + "_member",
		Name:         "Member",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	// 创建另一个工作区的用户（非成员）
	otherWorkspace := &model.Workspace{
		Name: "Other Workspace " + prefix,
		Slug: "other-workspace-" + prefix,
	}
	if err := tx.Create(otherWorkspace).Error; err != nil {
		t.Fatalf("创建其他工作区失败: %v", err)
	}
	nonMember := &model.User{
		WorkspaceID:  otherWorkspace.ID,
		Email:        prefix + "_non@example.com",
		Username:     prefix + "_non",
		Name:         "NonMember",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	tx.Create(member)
	tx.Create(nonMember)

	tests := []struct {
		name          string
		userID        uuid.UUID
		workspaceID   string
		wantErr       bool
		errMsg        string
		checkWorkspace func(*model.Workspace) bool
	}{
		{
			name:        "正常获取",
			userID:      member.ID,
			workspaceID: workspace.ID.String(),
			wantErr:     false,
			checkWorkspace: func(ws *model.Workspace) bool {
				return ws.Name == workspace.Name
			},
		},
		{
			name:        "无权限访问（非工作区成员）",
			userID:      nonMember.ID,
			workspaceID: workspace.ID.String(),
			wantErr:     true,
			errMsg:      "无权限",
		},
		{
			name:        "工作区不存在",
			userID:      member.ID,
			workspaceID: uuid.New().String(),
			wantErr:     true,
			errMsg:      "工作区不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置上下文用户
			ctx := context.WithValue(ctx, "user_id", tt.userID)

			ws, err := svc.GetWorkspace(ctx, tt.workspaceID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorkspace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr && tt.checkWorkspace != nil {
				if !tt.checkWorkspace(ws) {
					t.Error("GetWorkspace() 返回的工作区不符合预期")
				}
			}
		})
	}
}

// =============================================================================
// UpdateWorkspace 测试
// =============================================================================

func TestWorkspaceService_UpdateWorkspace(t *testing.T) {
	if testWorkspaceServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceServiceDB.Begin()
	defer tx.Rollback()

	workspaceStore := store.NewWorkspaceStore(tx)
	userStore := store.NewUserStore(tx)
	svc := NewWorkspaceService(workspaceStore, userStore)

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "UpdateWorkspace Test " + prefix,
		Slug: "updateworkspace-test-" + prefix,
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
		name        string
		userID      uuid.UUID
		userRole    model.Role
		workspaceID string
		updates     map[string]interface{}
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Admin 更新名称",
			userID:      admin.ID,
			userRole:    admin.Role,
			workspaceID: workspace.ID.String(),
			updates: map[string]interface{}{
				"name": "Updated Name " + prefix,
			},
			wantErr: false,
		},
		{
			name:        "Admin 更新 Logo",
			userID:      admin.ID,
			userRole:    admin.Role,
			workspaceID: workspace.ID.String(),
			updates: map[string]interface{}{
				"logo_url": "https://example.com/logo.png",
			},
			wantErr: false,
		},
		{
			name:        "Member 无权限",
			userID:      member.ID,
			userRole:    member.Role,
			workspaceID: workspace.ID.String(),
			updates: map[string]interface{}{
				"name": "Should Not Update",
			},
			wantErr: true,
			errMsg:  "无权限",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置上下文
			ctx := context.WithValue(context.Background(), "user_id", tt.userID)
			ctx = context.WithValue(ctx, "user_role", tt.userRole)

			ws, err := svc.UpdateWorkspace(ctx, tt.workspaceID, tt.updates)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateWorkspace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				// 验证更新
				if name, ok := tt.updates["name"]; ok {
					assert.Equal(t, name, ws.Name)
				}
				if logoURL, ok := tt.updates["logo_url"]; ok {
					assert.NotNil(t, ws.LogoURL)
					assert.Equal(t, logoURL, *ws.LogoURL)
				}
			}
		})
	}
}

// =============================================================================
// GetWorkspaceStats 测试
// =============================================================================

func TestWorkspaceService_GetWorkspaceStats(t *testing.T) {
	if testWorkspaceServiceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	tx := testWorkspaceServiceDB.Begin()
	defer tx.Rollback()

	workspaceStore := store.NewWorkspaceStore(tx)
	userStore := store.NewUserStore(tx)
	svc := NewWorkspaceService(workspaceStore, userStore)

	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "GetWorkspaceStats Test " + prefix,
		Slug: "getworkspacestats-test-" + prefix,
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

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "GS" + "ABC",
	}
	tx.Create(team)

	tests := []struct {
		name          string
		userID        uuid.UUID
		workspaceID   string
		wantErr       bool
		checkStats    func(*store.WorkspaceStats) bool
	}{
		{
			name:        "正常获取统计",
			userID:      admin.ID,
			workspaceID: workspace.ID.String(),
			wantErr:     false,
			checkStats: func(stats *store.WorkspaceStats) bool {
				return stats.TeamsCount >= 1 && stats.MembersCount >= 1
			},
		},
		{
			name:        "工作区不存在",
			userID:      admin.ID,
			workspaceID: uuid.New().String(),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(ctx, "user_id", tt.userID)

			stats, err := svc.GetWorkspaceStats(ctx, tt.workspaceID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorkspaceStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkStats != nil {
				if !tt.checkStats(stats) {
					t.Errorf("GetWorkspaceStats() 返回的统计不符合预期: %+v", stats)
				}
			}
		})
	}
}
