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

func TestTeamMemberHandler_ListMembers(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	tx := db.Begin()
	defer tx.Rollback()

	cfg := &config.Config{
		JWTSecret:        "teammember-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	prefix := uuid.New().String()[:8]

	workspace := &model.Workspace{Name: "TMList Test " + prefix, Slug: "tmlist-test-" + prefix}
	tx.Create(workspace)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Team " + prefix, Key: "TM" + "ABC"}
	tx.Create(team)

	member := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "@example.com", Username: prefix + "_member",
		Name: "Member", PasswordHash: "hash", Role: model.RoleMember,
	}
	tx.Create(member)
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: model.RoleMember, JoinedAt: time.Now()})

	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamStore := store.NewTeamStore(tx)
	teamMemberService := service.NewTeamMemberService(teamMemberStore, userStore, teamStore)
	handler := NewTeamMemberHandler(teamMemberService)

	token, _ := jwtService.GenerateAccessToken(member.ID, member.Email, member.Role)

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.GET("/teams/:teamId/members", handler.ListMembers)

	req := httptest.NewRequest("GET", "/teams/"+team.ID.String()+"/members", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTeamMemberHandler_AddMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	tx := db.Begin()
	defer tx.Rollback()

	cfg := &config.Config{
		JWTSecret:        "teammember-add-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	prefix := uuid.New().String()[:8]

	workspace := &model.Workspace{Name: "TMAdd Test " + prefix, Slug: "tmadd-test-" + prefix}
	tx.Create(workspace)

	team := &model.Team{WorkspaceID: workspace.ID, Name: "Team " + prefix, Key: "TA" + "ABC"}
	tx.Create(team)

	owner := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "_owner@example.com", Username: prefix + "_owner",
		Name: "Owner", PasswordHash: "hash", Role: model.RoleMember,
	}
	newUser := &model.User{
		WorkspaceID: workspace.ID, Email: prefix + "_new@example.com", Username: prefix + "_new",
		Name: "New", PasswordHash: "hash", Role: model.RoleMember,
	}
	tx.Create(owner)
	tx.Create(newUser)
	tx.Create(&model.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: model.RoleAdmin, JoinedAt: time.Now()})

	teamMemberStore := store.NewTeamMemberStore(tx)
	userStore := store.NewUserStore(tx)
	teamStore := store.NewTeamStore(tx)
	teamMemberService := service.NewTeamMemberService(teamMemberStore, userStore, teamStore)
	handler := NewTeamMemberHandler(teamMemberService)

	token, _ := jwtService.GenerateAccessToken(owner.ID, owner.Email, owner.Role)

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("db", tx) })
	router.Use(middleware.Auth(jwtService))
	router.POST("/teams/:teamId/members", handler.AddMember)

	bodyBytes, _ := json.Marshal(map[string]interface{}{
		"user_id": newUser.ID.String(),
		"role":    "member",
	})
	req := httptest.NewRequest("POST", "/teams/"+team.ID.String()+"/members", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
}
