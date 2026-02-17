package handler

import (
	"bytes"
	"context"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupWorkflowHandlerTest(t *testing.T) (*gin.Engine, *gorm.DB, service.JWTService) {
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
		JWTSecret:        "workflow-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	return gin.New(), db, jwtService
}

func TestWorkflowHandler_ListStates(t *testing.T) {
	router, db, jwtService := setupWorkflowHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	// Setup Test Data
	workspace := &model.Workspace{Name: "Workflow Test " + prefix, Slug: "wf-test-" + prefix}
	tx.Create(workspace)

	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)

	// Team Service will auto-init states
	stateStore := store.NewWorkflowStateStore(tx)
	teamStore := store.NewTeamStore(tx)
	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)

	workflowSvc := service.NewWorkflowService(stateStore, teamStore)
	teamSvc := service.NewTeamService(teamStore, teamMemberStore, userStore, workflowSvc)

	// Create Team (this should trigger state initialization)
	ctx := context.WithValue(context.Background(), "user_id", admin.ID)
	ctx = context.WithValue(ctx, "user_role", admin.Role)
	ctx = context.WithValue(ctx, "workspace_id", workspace.ID)
	team, err := teamSvc.CreateTeam(ctx, "Test Team", "TT", "")
	require.NoError(t, err)

	workflowHandler := NewWorkflowHandler(workflowSvc)
	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.GET("/api/v1/teams/:teamId/workflow-states", workflowHandler.ListStates)

	req := httptest.NewRequest("GET", "/api/v1/teams/"+team.ID.String()+"/workflow-states", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []model.WorkflowState
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp, 5) // Default 5 states
	assert.Equal(t, "Backlog", resp[0].Name)
}

func TestWorkflowHandler_CreateState(t *testing.T) {
	router, db, jwtService := setupWorkflowHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]
	workspace := &model.Workspace{Name: "Workflow Post Test " + prefix, Slug: "wf-post-test-" + prefix}
	tx.Create(workspace)

	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)

	stateStore := store.NewWorkflowStateStore(tx)
	teamStore := store.NewTeamStore(tx)
	workflowSvc := service.NewWorkflowService(stateStore, teamStore)
	workflowHandler := NewWorkflowHandler(workflowSvc)
	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Test Team", Key: "TT" + prefix}
	tx.Create(team)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.POST("/api/v1/teams/:teamId/workflow-states", workflowHandler.CreateState)

	body := map[string]interface{}{
		"name": "New Handler State",
		"type": string(model.StateTypeStarted),
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/teams/"+team.ID.String()+"/workflow-states", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp model.WorkflowState
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "New Handler State", resp.Name)
	assert.Equal(t, model.StateTypeStarted, resp.Type)
}

func TestWorkflowHandler_UpdateState(t *testing.T) {
	router, db, jwtService := setupWorkflowHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]
	workspace := &model.Workspace{Name: "Workflow Put Test " + prefix, Slug: "wf-put-test-" + prefix}
	tx.Create(workspace)
	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)
	team := &model.Team{WorkspaceID: workspace.ID, Name: "Test Team", Key: "TT" + prefix}
	tx.Create(team)
	state := &model.WorkflowState{TeamID: team.ID, Name: "Original", Type: model.StateTypeUnstarted}
	tx.Create(state)

	stateStore := store.NewWorkflowStateStore(tx)
	teamStore := store.NewTeamStore(tx)
	workflowSvc := service.NewWorkflowService(stateStore, teamStore)
	workflowHandler := NewWorkflowHandler(workflowSvc)
	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.PUT("/api/v1/workflow-states/:id", workflowHandler.UpdateState)

	newName := "Updated Name"
	body := map[string]interface{}{"name": newName}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/v1/workflow-states/"+state.ID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.WorkflowState
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, newName, resp.Name)
}

func TestWorkflowHandler_DeleteState(t *testing.T) {
	router, db, jwtService := setupWorkflowHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]
	workspace := &model.Workspace{Name: "Workflow Delete Test " + prefix, Slug: "wf-del-test-" + prefix}
	tx.Create(workspace)
	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)
	team := &model.Team{WorkspaceID: workspace.ID, Name: "Test Team", Key: "TT" + prefix}
	tx.Create(team)

	// Create two states of same type to allow deletion
	state1 := &model.WorkflowState{TeamID: team.ID, Name: "S1", Type: model.StateTypeUnstarted}
	state2 := &model.WorkflowState{TeamID: team.ID, Name: "S2", Type: model.StateTypeUnstarted}
	tx.Create(state1)
	tx.Create(state2)

	stateStore := store.NewWorkflowStateStore(tx)
	teamStore := store.NewTeamStore(tx)
	workflowSvc := service.NewWorkflowService(stateStore, teamStore)
	workflowHandler := NewWorkflowHandler(workflowSvc)
	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.DELETE("/api/v1/workflow-states/:id", workflowHandler.DeleteState)

	req := httptest.NewRequest("DELETE", "/api/v1/workflow-states/"+state1.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
