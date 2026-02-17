package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/config"
	"github.com/liwei0526vip/mylinear/internal/handler"
	"github.com/liwei0526vip/mylinear/internal/middleware"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
}

var testRouterDB *gorm.DB

func setupRouterTest(t *testing.T) (*gin.Engine, *gorm.DB, service.JWTService) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skip("无法连接数据库")
	}

	cfg := &config.Config{
		JWTSecret:        "router-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	return gin.New(), db, jwtService
}

func TestWorkspaceRoutes(t *testing.T) {
	router, db, jwtService := setupRouterTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	// 创建测试数据
	workspace := &model.Workspace{Name: "Router Test " + prefix, Slug: "router-test-" + prefix}
	tx.Create(workspace)

	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)

	// 设置服务和处理器
	workspaceStore := store.NewWorkspaceStore(tx)
	userStore := store.NewUserStore(tx)
	workspaceService := service.NewWorkspaceService(workspaceStore, userStore)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)

	// 生成 token
	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	// 设置中间件
	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))

	// 注册 Workspace 路由
	workspaceGroup := router.Group("/workspaces")
	{
		workspaceGroup.GET("/:id", workspaceHandler.GetWorkspace)
		workspaceGroup.PUT("/:id", workspaceHandler.UpdateWorkspace)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{"GET workspace", "GET", "/workspaces/" + workspace.ID.String(), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}

func TestTeamRoutes(t *testing.T) {
	router, db, jwtService := setupRouterTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	// 创建测试数据
	workspace := &model.Workspace{Name: "Team Router Test " + prefix, Slug: "team-router-test-" + prefix}
	tx.Create(workspace)

	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Team " + prefix, Key: "RT" + "ABC"}
	tx.Create(team)

	// 设置服务和处理器
	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamService := service.NewTeamService(teamStore, teamMemberStore, userStore)
	teamHandler := handler.NewTeamHandler(teamService)

	// 生成 token
	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	// 设置中间件
	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))

	// 注册 Team 路由
	teamsGroup := router.Group("/teams")
	{
		teamsGroup.GET("", teamHandler.ListTeams)
		teamsGroup.POST("", teamHandler.CreateTeam)
		teamsGroup.GET("/:teamId", teamHandler.GetTeam)
		teamsGroup.PUT("/:teamId", teamHandler.UpdateTeam)
		teamsGroup.DELETE("/:teamId", teamHandler.DeleteTeam)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{"GET teams list", "GET", "/teams?workspace_id=" + workspace.ID.String(), http.StatusOK},
		{"GET team", "GET", "/teams/" + team.ID.String(), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}

func TestTeamMemberRoutes(t *testing.T) {
	router, db, jwtService := setupRouterTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	// 创建测试数据
	workspace := &model.Workspace{Name: "TM Router Test " + prefix, Slug: "tm-router-test-" + prefix}
	tx.Create(workspace)

	owner := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_owner",
		Name: "Owner", PasswordHash: "hash", Role: model.RoleMember,
	}
	tx.Create(owner)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Team " + prefix, Key: "TM" + "ABC"}
	tx.Create(team)
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})

	// 设置服务和处理器
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamStore := store.NewTeamStore(tx)
	teamMemberService := service.NewTeamMemberService(teamMemberStore, userStore, teamStore)
	teamMemberHandler := handler.NewTeamMemberHandler(teamMemberService)

	// 生成 token
	token, _ := jwtService.GenerateAccessToken(owner.ID, owner.Email, owner.Role)

	// 设置中间件
	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))

	// 注册 TeamMember 路由
	membersGroup := router.Group("/teams/:teamId/members")
	{
		membersGroup.GET("", teamMemberHandler.ListMembers)
		membersGroup.POST("", teamMemberHandler.AddMember)
		membersGroup.DELETE("/:userId", teamMemberHandler.RemoveMember)
		membersGroup.PUT("/:userId", teamMemberHandler.UpdateMemberRole)
	}

	req := httptest.NewRequest("GET", "/teams/"+team.ID.String()+"/members", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestMiddlewareChain(t *testing.T) {
	router, db, jwtService := setupRouterTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	workspace := &model.Workspace{Name: "MW Test " + prefix, Slug: "mw-test-" + prefix}
	tx.Create(workspace)

	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Team " + prefix, Key: "MW" + "ABC"}
	tx.Create(team)

	// 设置服务和处理器
	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamService := service.NewTeamService(teamStore, teamMemberStore, userStore)
	teamHandler := handler.NewTeamHandler(teamService)

	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	// 设置中间件链
	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.Use(middleware.RequireTeamMember())

	// 注册需要团队成员权限的路由
	router.GET("/teams/:teamId/protected", teamHandler.GetTeam)

	tests := []struct {
		name           string
		token          string
		path           string
		wantStatusCode int
	}{
		{"有认证", token, "/teams/" + team.ID.String() + "/protected", http.StatusOK},
		{"无认证", "", "/teams/" + team.ID.String() + "/protected", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}
