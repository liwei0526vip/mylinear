package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/config"
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

func setupTeamHandlerTest(t *testing.T) (*gin.Engine, *gorm.DB, service.JWTService) {
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
		JWTSecret:        "team-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	return gin.New(), db, jwtService
}

func TestTeamHandler_ListTeams(t *testing.T) {
	router, db, jwtService := setupTeamHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	workspace := &model.Workspace{Name: "ListTeams Test " + prefix, Slug: "listteams-test-" + prefix}
	tx.Create(workspace)

	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)

	for i := 0; i < 3; i++ {
		team := &model.Team{
			WorkspaceID: workspace.ID, Name: "Team " + string(rune('A'+i)) + " " + prefix,
			Key: "LT" + string(rune('A'+i)) + "ABC",
		}
		tx.Create(team)
	}

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := service.NewWorkflowService(workflowStateStore, teamStore)
	teamService := service.NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)
	handler := NewTeamHandler(teamService)

	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.GET("/teams", handler.ListTeams)

	req := httptest.NewRequest("GET", "/teams?workspace_id="+workspace.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTeamHandler_CreateTeam(t *testing.T) {
	router, db, jwtService := setupTeamHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	workspace := &model.Workspace{Name: "CreateTeam Test " + prefix, Slug: "createteam-test-" + prefix}
	tx.Create(workspace)

	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	member := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "_member@example.com", Username: prefix + "_member",
		Name: "Member", PasswordHash: "hash", Role: model.RoleMember,
	}
	tx.Create(admin)
	tx.Create(member)

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := service.NewWorkflowService(workflowStateStore, teamStore)
	teamService := service.NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)
	handler := NewTeamHandler(teamService)

	adminToken, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)
	memberToken, _ := jwtService.GenerateAccessToken(member.ID, member.Email, member.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.POST("/teams", handler.CreateTeam)

	tests := []struct {
		name           string
		token          string
		body           map[string]interface{}
		wantStatusCode int
	}{
		{
			name:  "Admin 创建团队",
			token: adminToken,
			body: map[string]interface{}{
				"name":         "New Team",
				"key":          "NT" + "ABC",
				"workspace_id": workspace.ID.String(),
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:  "Member 无权限",
			token: memberToken,
			body: map[string]interface{}{
				"name":         "Member Team",
				"key":          "MT" + "ABC",
				"workspace_id": workspace.ID.String(),
			},
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/teams", bytes.NewReader(bodyBytes))
			req.Header.Set("Authorization", "Bearer "+tt.token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}

func TestTeamHandler_GetTeam(t *testing.T) {
	router, db, jwtService := setupTeamHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	workspace := &model.Workspace{Name: "GetTeam Test " + prefix, Slug: "getteam-test-" + prefix}
	tx.Create(workspace)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Team " + prefix, Key: "GT" + "ABC"}
	tx.Create(team)

	member := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_member",
		Name: "Member", PasswordHash: "hash", Role: model.RoleMember,
	}
	tx.Create(member)
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := service.NewWorkflowService(workflowStateStore, teamStore)
	teamService := service.NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)
	handler := NewTeamHandler(teamService)

	token, _ := jwtService.GenerateAccessToken(member.ID, member.Email, member.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.GET("/teams/:teamId", handler.GetTeam)

	req := httptest.NewRequest("GET", "/teams/"+team.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTeamHandler_UpdateTeam(t *testing.T) {
	router, db, jwtService := setupTeamHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	workspace := &model.Workspace{Name: "UpdateTeam Test " + prefix, Slug: "updateteam-test-" + prefix}
	tx.Create(workspace)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Team " + prefix, Key: "UT" + "ABC"}
	tx.Create(team)

	owner := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_owner",
		Name: "Owner", PasswordHash: "hash", Role: model.RoleMember,
	}
	tx.Create(owner)
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := service.NewWorkflowService(workflowStateStore, teamStore)
	teamService := service.NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)
	handler := NewTeamHandler(teamService)

	token, _ := jwtService.GenerateAccessToken(owner.ID, owner.Email, owner.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.PUT("/teams/:teamId", handler.UpdateTeam)

	bodyBytes, _ := json.Marshal(map[string]interface{}{"name": "Updated Team"})
	req := httptest.NewRequest("PUT", "/teams/"+team.ID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestTeamHandler_DeleteTeam(t *testing.T) {
	router, db, jwtService := setupTeamHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	workspace := &model.Workspace{Name: "DeleteTeam Test " + prefix, Slug: "deleteteam-test-" + prefix}
	tx.Create(workspace)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Team " + prefix, Key: "DT" + "ABC"}
	tx.Create(team)

	owner := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_owner",
		Name: "Owner", PasswordHash: "hash", Role: model.RoleMember,
	}
	tx.Create(owner)
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})

	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	workflowStateStore := store.NewWorkflowStateStore(tx)
	workflowSvc := service.NewWorkflowService(workflowStateStore, teamStore)
	teamService := service.NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)
	handler := NewTeamHandler(teamService)

	token, _ := jwtService.GenerateAccessToken(owner.ID, owner.Email, owner.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.DELETE("/teams/:teamId", handler.DeleteTeam)

	req := httptest.NewRequest("DELETE", "/teams/"+team.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}
