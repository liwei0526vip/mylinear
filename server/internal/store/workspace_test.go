package store

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mylinear/server/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestWorkspaceStore_Interface 测试 WorkspaceStore 接口定义存在
func TestWorkspaceStore_Interface(t *testing.T) {
	var _ WorkspaceStore = (*workspaceStore)(nil)
}

// testWorkspaceDB 用于工作区测试的数据库连接
var testWorkspaceDB *gorm.DB

// TestMain_Workspace 设置工作区集成测试环境
func init() {
	// 使用主数据库进行集成测试
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable"
	}

	// 尝试连接数据库
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("警告: 无法连接测试数据库: %v\n", err)
		return
	}

	testWorkspaceDB = db
}

// =============================================================================
// GetByID 测试
// =============================================================================

func TestWorkspaceStore_GetByID(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 使用事务进行测试隔离
	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewWorkspaceStore(tx)
	ctx := context.Background()

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "Test Workspace " + uuid.New().String()[:8],
		Slug: "test-ws-" + uuid.New().String()[:8],
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	tests := []struct {
		name      string
		id        string
		wantErr   bool
		wantFound bool
		checkWS   func(*model.Workspace) bool
	}{
		{
			name:      "正常获取工作区",
			id:        workspace.ID.String(),
			wantErr:   false,
			wantFound: true,
			checkWS: func(ws *model.Workspace) bool {
				return ws.Name == workspace.Name && ws.Slug == workspace.Slug
			},
		},
		{
			name:      "工作区不存在",
			id:        uuid.New().String(),
			wantErr:   true,
			wantFound: false,
			checkWS:   nil,
		},
		{
			name:      "无效的ID格式",
			id:        "invalid-uuid",
			wantErr:   true,
			wantFound: false,
			checkWS:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := store.GetByID(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantFound && found == nil {
				t.Error("GetByID() 应该找到工作区")
				return
			}

			if !tt.wantFound && found != nil {
				t.Error("GetByID() 不应该找到工作区")
				return
			}

			if tt.checkWS != nil && found != nil {
				if !tt.checkWS(found) {
					t.Error("GetByID() 返回的工作区字段不符合预期")
				}
			}
		})
	}
}

// =============================================================================
// Update 测试
// =============================================================================

func TestWorkspaceStore_Update(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 使用事务进行测试隔离
	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewWorkspaceStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "Update Test " + prefix,
		Slug: "update-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	tests := []struct {
		name     string
		setupFn  func(*model.Workspace)
		wantErr  bool
		checkFn  func(*model.Workspace) bool
	}{
		{
			name: "更新名称",
			setupFn: func(ws *model.Workspace) {
				ws.Name = "Updated Name " + prefix
			},
			wantErr: false,
			checkFn: func(ws *model.Workspace) bool {
				return ws.Name == "Updated Name "+prefix
			},
		},
		{
			name: "更新Logo URL",
			setupFn: func(ws *model.Workspace) {
				logoURL := "https://example.com/logo.png"
				ws.LogoURL = &logoURL
			},
			wantErr: false,
			checkFn: func(ws *model.Workspace) bool {
				return ws.LogoURL != nil && *ws.LogoURL == "https://example.com/logo.png"
			},
		},
		{
			name: "清空Logo URL",
			setupFn: func(ws *model.Workspace) {
				ws.LogoURL = nil
			},
			wantErr: false,
			checkFn: func(ws *model.Workspace) bool {
				return ws.LogoURL == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 应用修改
			tt.setupFn(workspace)

			err := store.Update(ctx, workspace)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证更新成功
				found, err := store.GetByID(ctx, workspace.ID.String())
				if err != nil {
					t.Errorf("GetByID() error = %v", err)
					return
				}
				if !tt.checkFn(found) {
					t.Error("Update() 字段未正确更新")
				}
			}
		})
	}
}

func TestWorkspaceStore_Update_DuplicateSlug(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 使用事务进行测试隔离
	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewWorkspaceStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建两个工作区
	ws1 := &model.Workspace{
		Name: "Workspace 1 " + prefix,
		Slug: "ws1-" + prefix,
	}
	ws2 := &model.Workspace{
		Name: "Workspace 2 " + prefix,
		Slug: "ws2-" + prefix,
	}
	if err := tx.Create(ws1).Error; err != nil {
		t.Fatalf("创建工作区1失败: %v", err)
	}
	if err := tx.Create(ws2).Error; err != nil {
		t.Fatalf("创建工作区2失败: %v", err)
	}

	// 尝试将 ws2 的 Slug 更新为 ws1 的 Slug
	ws2.Slug = "ws1-" + prefix
	err := store.Update(ctx, ws2)
	if err == nil {
		t.Error("Update() 应该返回错误，Slug 重复")
	}
}

// =============================================================================
// GetStats 测试
// =============================================================================

func TestWorkspaceStore_GetStats(t *testing.T) {
	if testWorkspaceDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 使用事务进行测试隔离
	tx := testWorkspaceDB.Begin()
	defer tx.Rollback()

	store := NewWorkspaceStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "Stats Test " + prefix,
		Slug: "stats-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "TM" + prefix[:3],
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "@example.com",
		Username:     prefix + "_user",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	tests := []struct {
		name         string
		workspaceID  string
		wantErr      bool
		checkStats   func(*WorkspaceStats) bool
	}{
		{
			name:        "正常获取统计",
			workspaceID: workspace.ID.String(),
			wantErr:     false,
			checkStats: func(stats *WorkspaceStats) bool {
				return stats.TeamsCount >= 1 && stats.MembersCount >= 1
			},
		},
		{
			name:        "工作区不存在",
			workspaceID: uuid.New().String(),
			wantErr:     false, // 返回零值统计
			checkStats: func(stats *WorkspaceStats) bool {
				return stats.TeamsCount == 0 && stats.MembersCount == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := store.GetStats(ctx, tt.workspaceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && stats == nil {
				t.Error("GetStats() 应该返回统计信息")
				return
			}

			if tt.checkStats != nil && stats != nil {
				if !tt.checkStats(stats) {
					t.Errorf("GetStats() 返回的统计信息不符合预期: %+v", stats)
				}
			}
		})
	}
}
