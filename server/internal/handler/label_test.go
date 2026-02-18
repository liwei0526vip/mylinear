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
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupLabelHandlerTest(t *testing.T) (*gin.Engine, *gorm.DB, service.JWTService) {
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
		JWTSecret:        "label-handler-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	return gin.New(), db, jwtService
}

func TestLabelHandler_ListLabels(t *testing.T) {
	router, db, jwtService := setupLabelHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]
	workspace := &model.Workspace{Name: "Label Test WS " + prefix, Slug: "label-ws-" + prefix}
	tx.Create(workspace)
	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)
	team := &model.Team{WorkspaceID: workspace.ID, Name: "Test Team", Key: "TT" + prefix}
	tx.Create(team)

	// Create a workspace label and a team label
	tx.Create(&model.Label{WorkspaceID: workspace.ID, Name: "Global", Color: "#fff"})
	tx.Create(&model.Label{WorkspaceID: workspace.ID, TeamID: &team.ID, Name: "Private", Color: "#000"})

	labelStore := store.NewLabelStore(tx)
	teamStore := store.NewTeamStore(tx)
	labelSvc := service.NewLabelService(labelStore)
	labelHandler := NewLabelHandler(labelSvc, teamStore)
	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.GET("/api/v1/teams/:teamId/labels", labelHandler.ListLabels)

	req := httptest.NewRequest("GET", "/api/v1/teams/"+team.ID.String()+"/labels", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data []model.Label `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Data, 2)
}

func TestLabelHandler_CreateLabel(t *testing.T) {
	router, db, jwtService := setupLabelHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]
	workspace := &model.Workspace{Name: "Label Test WS " + prefix, Slug: "label-post-ws-" + prefix}
	tx.Create(workspace)
	admin := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_admin",
		Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin,
	}
	tx.Create(admin)
	team := &model.Team{WorkspaceID: workspace.ID, Name: "Test Team", Key: "TT" + prefix}
	tx.Create(team)

	labelStore := store.NewLabelStore(tx)
	teamStore := store.NewTeamStore(tx)
	labelSvc := service.NewLabelService(labelStore)
	labelHandler := NewLabelHandler(labelSvc, teamStore)
	token, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)

	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.POST("/api/v1/teams/:teamId/labels", labelHandler.CreateLabel)

	body := map[string]interface{}{
		"name":  "New Team Label",
		"color": "#123456",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/teams/"+team.ID.String()+"/labels", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp struct {
		Data model.Label `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "New Team Label", resp.Data.Name)
	assert.Equal(t, &team.ID, resp.Data.TeamID)
}
