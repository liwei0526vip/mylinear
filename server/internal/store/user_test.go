package store

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestUserStore_Interface 测试 UserStore 接口定义存在
func TestUserStore_Interface(t *testing.T) {
	var _ UserStore = (*userStore)(nil)
}

// TestMain 设置集成测试环境
func TestMain(m *testing.M) {
	// 使用主数据库进行集成测试（事务隔离）
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
		fmt.Println("集成测试将被跳过")
		os.Exit(0) // 不报错，只是跳过
	}

	// 运行迁移（自动迁移测试表）
	// 先清理旧表，避免 schema 不一致
	// 注意：CASCADE 会删除依赖表，顺序不重要，但全面清理更安全
	db.Exec("DROP TABLE IF EXISTS issue_subscriptions CASCADE")
	db.Exec("DROP TABLE IF EXISTS issues CASCADE")
	db.Exec("DROP TABLE IF EXISTS labels CASCADE")
	db.Exec("DROP TABLE IF EXISTS workflow_states CASCADE")
	db.Exec("DROP TABLE IF EXISTS team_members CASCADE")
	db.Exec("DROP TABLE IF EXISTS teams CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")
	db.Exec("DROP TABLE IF EXISTS workspaces CASCADE")

	// 注意：顺序很重要，先父后子
	err = db.AutoMigrate(
		&model.Workspace{},
		&model.User{},
		&model.Team{},
		&model.TeamMember{},
		&model.WorkflowState{},
		&model.Label{},
		&model.Issue{},
		&model.IssueSubscription{},
	)
	if err != nil {
		fmt.Printf("警告: 自动迁移失败: %v\n", err)
	}

	// 创建测试用的工作区（如果不存在）
	var workspace model.Workspace
	result := db.Where("name = ?", "Test Workspace").First(&workspace)
	if result.Error == gorm.ErrRecordNotFound {
		workspace = model.Workspace{
			Name: "Test Workspace",
			Slug: "test-workspace",
		}
		db.Create(&workspace)
	}
	testWorkspaceID = workspace.ID

	testDB = db

	code := m.Run()

	os.Exit(code)
}

var testDB *gorm.DB
var testWorkspaceID uuid.UUID

// =============================================================================
// CreateUser 测试
// =============================================================================

func TestUserStore_CreateUser_Success(t *testing.T) {
	// 使用事务进行测试隔离
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()

	// 使用唯一前缀避免冲突
	prefix := uuid.New().String()[:8]

	tests := []struct {
		name string
		user *model.User
	}{
		{
			name: "创建基本用户",
			user: &model.User{
				WorkspaceID:  testWorkspaceID,
				Email:        prefix + "_test1@example.com",
				Username:     prefix + "_testuser1",
				Name:         "Test User 1",
				PasswordHash: "hashed_password",
				Role:         model.RoleMember,
			},
		},
		{
			name: "创建管理员用户",
			user: &model.User{
				WorkspaceID:  testWorkspaceID,
				Email:        prefix + "_admin@example.com",
				Username:     prefix + "_adminuser",
				Name:         "Admin User",
				PasswordHash: "hashed_password",
				Role:         model.RoleAdmin,
			},
		},
		{
			name: "创建带头像的用户",
			user: &model.User{
				WorkspaceID:  testWorkspaceID,
				Email:        prefix + "_avatar@example.com",
				Username:     prefix + "_avataruser",
				Name:         "Avatar User",
				PasswordHash: "hashed_password",
				Role:         model.RoleMember,
				AvatarURL:    strPtr("https://example.com/avatar.png"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateUser(ctx, tt.user)
			if err != nil {
				t.Errorf("CreateUser() error = %v", err)
				return
			}

			// 验证用户已创建并有 ID
			if tt.user.ID == uuid.Nil {
				t.Error("CreateUser() 未设置用户 ID")
			}

			// 验证可以通过 ID 查询到
			found, err := store.GetUserByID(ctx, tt.user.ID.String())
			if err != nil {
				t.Errorf("GetUserByID() error = %v", err)
				return
			}

			if found.Email != tt.user.Email {
				t.Errorf("GetUserByID().Email = %v, want %v", found.Email, tt.user.Email)
			}
			if found.Username != tt.user.Username {
				t.Errorf("GetUserByID().Username = %v, want %v", found.Username, tt.user.Username)
			}
		})
	}
}

// strPtr 辅助函数：将字符串转为指针
func strPtr(s string) *string {
	return &s
}

// =============================================================================
// CreateUser 重复测试
// =============================================================================

func TestUserStore_CreateUser_DuplicateEmail(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建第一个用户
	user1 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_dup@example.com",
		Username:     prefix + "_user1",
		Name:         "User 1",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	err := store.CreateUser(ctx, user1)
	if err != nil {
		t.Fatalf("第一次创建用户失败: %v", err)
	}

	// 尝试用相同邮箱创建第二个用户
	user2 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_dup@example.com", // 相同邮箱
		Username:     prefix + "_user2",
		Name:         "User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	err = store.CreateUser(ctx, user2)
	if err == nil {
		t.Error("CreateUser() 应该返回错误，邮箱重复")
	}
}

func TestUserStore_CreateUser_DuplicateUsername(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建第一个用户
	user1 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_email1@example.com",
		Username:     prefix + "_dupuser",
		Name:         "User 1",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	err := store.CreateUser(ctx, user1)
	if err != nil {
		t.Fatalf("第一次创建用户失败: %v", err)
	}

	// 尝试用相同用户名创建第二个用户
	user2 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_email2@example.com",
		Username:     prefix + "_dupuser", // 相同用户名
		Name:         "User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	err = store.CreateUser(ctx, user2)
	if err == nil {
		t.Error("CreateUser() 应该返回错误，用户名重复")
	}
}

// =============================================================================
// GetUserByEmail 测试
// =============================================================================

func TestUserStore_GetUserByEmail(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试用户
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_find@example.com",
		Username:     prefix + "_finduser",
		Name:         "Find Me",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	err := store.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	tests := []struct {
		name      string
		email     string
		wantErr   bool
		wantFound bool
	}{
		{
			name:      "找到用户",
			email:     prefix + "_find@example.com",
			wantErr:   false,
			wantFound: true,
		},
		{
			name:      "用户不存在",
			email:     prefix + "_notexist@example.com",
			wantErr:   true,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := store.GetUserByEmail(ctx, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantFound && found == nil {
				t.Error("GetUserByEmail() 应该找到用户")
			}
			if !tt.wantFound && found != nil {
				t.Error("GetUserByEmail() 不应该找到用户")
			}
			if tt.wantFound && found != nil && found.Email != tt.email {
				t.Errorf("GetUserByEmail().Email = %v, want %v", found.Email, tt.email)
			}
		})
	}
}

// =============================================================================
// GetUserByID 测试
// =============================================================================

func TestUserStore_GetUserByID(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试用户
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_byid@example.com",
		Username:     prefix + "_byiduser",
		Name:         "By ID",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	err := store.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	tests := []struct {
		name      string
		id        string
		wantErr   bool
		wantFound bool
	}{
		{
			name:      "找到用户",
			id:        user.ID.String(),
			wantErr:   false,
			wantFound: true,
		},
		{
			name:      "用户不存在",
			id:        uuid.New().String(),
			wantErr:   true,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := store.GetUserByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantFound && found == nil {
				t.Error("GetUserByID() 应该找到用户")
			}
			if !tt.wantFound && found != nil {
				t.Error("GetUserByID() 不应该找到用户")
			}
		})
	}
}

// =============================================================================
// GetUserByUsername 测试
// =============================================================================

func TestUserStore_GetUserByUsername(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试用户
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_byname@example.com",
		Username:     prefix + "_bynameuser",
		Name:         "By Username",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	err := store.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	tests := []struct {
		name      string
		username  string
		wantErr   bool
		wantFound bool
	}{
		{
			name:      "找到用户",
			username:  prefix + "_bynameuser",
			wantErr:   false,
			wantFound: true,
		},
		{
			name:      "用户不存在",
			username:  prefix + "_notexist",
			wantErr:   true,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := store.GetUserByUsername(ctx, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantFound && found == nil {
				t.Error("GetUserByUsername() 应该找到用户")
			}
			if !tt.wantFound && found != nil {
				t.Error("GetUserByUsername() 不应该找到用户")
			}
		})
	}
}

// =============================================================================
// UpdateUser 测试
// =============================================================================

func TestUserStore_UpdateUser(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建测试用户
	user := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_update@example.com",
		Username:     prefix + "_updateuser",
		Name:         "Original Name",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	err := store.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	tests := []struct {
		name       string
		updateFn   func(*model.User)
		wantErr    bool
		checkField func(*model.User) bool
	}{
		{
			name: "更新用户名",
			updateFn: func(u *model.User) {
				u.Name = "Updated Name"
			},
			wantErr: false,
			checkField: func(u *model.User) bool {
				return u.Name == "Updated Name"
			},
		},
		{
			name: "更新头像URL",
			updateFn: func(u *model.User) {
				avatarURL := "https://example.com/new-avatar.png"
				u.AvatarURL = &avatarURL
			},
			wantErr: false,
			checkField: func(u *model.User) bool {
				return u.AvatarURL != nil && *u.AvatarURL == "https://example.com/new-avatar.png"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 应用更新
			tt.updateFn(user)

			err := store.UpdateUser(ctx, user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证更新成功
				found, err := store.GetUserByID(ctx, user.ID.String())
				if err != nil {
					t.Errorf("GetUserByID() error = %v", err)
					return
				}
				if !tt.checkField(found) {
					t.Error("UpdateUser() 字段未正确更新")
				}
			}
		})
	}
}

func TestUserStore_UpdateUser_DuplicateEmail(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建两个用户
	user1 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_user1@example.com",
		Username:     prefix + "_user1",
		Name:         "User 1",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_user2@example.com",
		Username:     prefix + "_user2",
		Name:         "User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}

	store.CreateUser(ctx, user1)
	store.CreateUser(ctx, user2)

	// 尝试将 user2 的邮箱更新为 user1 的邮箱
	user2.Email = prefix + "_user1@example.com"
	err := store.UpdateUser(ctx, user2)
	if err == nil {
		t.Error("UpdateUser() 应该返回错误，邮箱重复")
	}
}

func TestUserStore_UpdateUser_DuplicateUsername(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewUserStore(tx)
	ctx := context.Background()
	prefix := uuid.New().String()[:8]

	// 创建两个用户
	user1 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_user1u@example.com",
		Username:     prefix + "_user1u",
		Name:         "User 1",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2 := &model.User{
		WorkspaceID:  testWorkspaceID,
		Email:        prefix + "_user2u@example.com",
		Username:     prefix + "_user2u",
		Name:         "User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}

	store.CreateUser(ctx, user1)
	store.CreateUser(ctx, user2)

	// 尝试将 user2 的用户名更新为 user1 的用户名
	user2.Username = prefix + "_user1u"
	err := store.UpdateUser(ctx, user2)
	if err == nil {
		t.Error("UpdateUser() 应该返回错误，用户名重复")
	}
}
