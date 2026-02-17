package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylinear/server/internal/config"
	"github.com/mylinear/server/internal/model"
	"github.com/mylinear/server/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testPermissionDB *gorm.DB

func init() {
	gin.SetMode(gin.TestMode)

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

	testPermissionDB = db
}

// TestRequireTeamOwner 测试团队 Owner 权限中间件
func TestRequireTeamOwner(t *testing.T) {
	if testPermissionDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	cfg := &config.Config{
		JWTSecret:        "require-team-owner-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	// 使用事务进行测试隔离
	tx := testPermissionDB.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "RequireTeamOwner Test " + prefix,
		Slug: "requireteamowner-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "RO" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	owner := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_owner@example.com",
		Username:     prefix + "_owner",
		Name:         "Owner",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	member := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_member@example.com",
		Username:     prefix + "_member",
		Name:         "Member",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	admin := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_admin@example.com",
		Username:     prefix + "_admin",
		Name:         "Admin",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	nonMember := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_non@example.com",
		Username:     prefix + "_non",
		Name:         "NonMember",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	tx.Create(owner)
	tx.Create(member)
	tx.Create(admin)
	tx.Create(nonMember)

	// 添加团队成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name           string
		userID         uuid.UUID
		userRole       model.Role
		teamID         string
		wantStatusCode int
	}{
		{
			name:           "Owner 通过",
			userID:         owner.ID,
			userRole:       model.RoleMember,
			teamID:         team.ID.String(),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Member 拒绝",
			userID:         member.ID,
			userRole:       model.RoleMember,
			teamID:         team.ID.String(),
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "Admin 绕过",
			userID:         admin.ID,
			userRole:       model.RoleAdmin,
			teamID:         team.ID.String(),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "非成员拒绝",
			userID:         nonMember.ID,
			userRole:       model.RoleMember,
			teamID:         team.ID.String(),
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := jwtService.GenerateAccessToken(tt.userID, "test@example.com", tt.userRole)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				// 注入数据库连接到上下文
				c.Set("db", tx)
			})
			router.Use(Auth(jwtService))
			router.Use(RequireTeamOwner())
			router.GET("/teams/:teamId/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/teams/"+tt.teamID+"/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}

// TestRequireTeamMember 测试团队成员权限中间件
func TestRequireTeamMember(t *testing.T) {
	if testPermissionDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	cfg := &config.Config{
		JWTSecret:        "require-team-member-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	// 使用事务进行测试隔离
	tx := testPermissionDB.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "RequireTeamMember Test " + prefix,
		Slug: "requireteammember-test-" + prefix,
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建测试工作区失败: %v", err)
	}

	// 创建测试团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        "Team " + prefix,
		Key:         "RM" + "ABC",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建测试团队失败: %v", err)
	}

	// 创建测试用户
	owner := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_owner@example.com",
		Username:     prefix + "_owner",
		Name:         "Owner",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	member := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_member@example.com",
		Username:     prefix + "_member",
		Name:         "Member",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	nonMember := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_non@example.com",
		Username:     prefix + "_non",
		Name:         "NonMember",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	tx.Create(owner)
	tx.Create(member)
	tx.Create(nonMember)

	// 添加团队成员
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	tests := []struct {
		name           string
		userID         uuid.UUID
		userRole       model.Role
		teamID         string
		wantStatusCode int
	}{
		{
			name:           "Owner 通过",
			userID:         owner.ID,
			userRole:       model.RoleMember,
			teamID:         team.ID.String(),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Member 通过",
			userID:         member.ID,
			userRole:       model.RoleMember,
			teamID:         team.ID.String(),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "非成员拒绝",
			userID:         nonMember.ID,
			userRole:       model.RoleMember,
			teamID:         team.ID.String(),
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := jwtService.GenerateAccessToken(tt.userID, "test@example.com", tt.userRole)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				// 注入数据库连接到上下文
				c.Set("db", tx)
			})
			router.Use(Auth(jwtService))
			router.Use(RequireTeamMember())
			router.GET("/teams/:teamId/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/teams/"+tt.teamID+"/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}
