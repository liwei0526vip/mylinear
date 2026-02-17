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

var workspaceHandlerDB *gorm.DB

func setupWorkspaceHandlerTest(t *testing.T) (*gin.Engine, *gorm.DB, uuid.UUID) {
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

	return gin.New(), db, uuid.New()
}

func TestWorkspaceHandler_GetWorkspace(t *testing.T) {
	router, db, _ := setupWorkspaceHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "GetWorkspace Handler Test " + prefix,
		Slug: "getworkspace-handler-test-" + prefix,
	}
	tx.Create(workspace)

	// 创建测试用户
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "@example.com",
		Username:     prefix + "_user",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
	}
	tx.Create(user)

	// 设置服务
	cfg := &config.Config{
		JWTSecret:        "workspace-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)
	workspaceStore := store.NewWorkspaceStore(tx)
	userStore := store.NewUserStore(tx)
	workspaceService := service.NewWorkspaceService(workspaceStore, userStore)

	// 创建 handler
	handler := NewWorkspaceHandler(workspaceService)

	// 生成 token
	token, _ := jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)

	router.Use(func(c *gin.Context) {
		c.Set("db", tx)
	})
	router.Use(middleware.Auth(jwtService))
	router.GET("/workspaces/:id", handler.GetWorkspace)

	tests := []struct {
		name           string
		workspaceID    string
		setupAuth      func(req *http.Request)
		wantStatusCode int
	}{
		{
			name:        "正常获取",
			workspaceID: workspace.ID.String(),
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "未认证",
			workspaceID: workspace.ID.String(),
			setupAuth: func(req *http.Request) {
				// 不设置认证
			},
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/workspaces/"+tt.workspaceID, nil)
			tt.setupAuth(req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}

func TestWorkspaceHandler_UpdateWorkspace(t *testing.T) {
	router, db, _ := setupWorkspaceHandlerTest(t)
	tx := db.Begin()
	defer tx.Rollback()

	prefix := uuid.New().String()[:8]

	// 创建测试工作区
	workspace := &model.Workspace{
		Name: "UpdateWorkspace Handler Test " + prefix,
		Slug: "updateworkspace-handler-test-" + prefix,
	}
	tx.Create(workspace)

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

	// 设置服务
	cfg := &config.Config{
		JWTSecret:        "workspace-update-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)
	workspaceStore := store.NewWorkspaceStore(tx)
	userStore := store.NewUserStore(tx)
	workspaceService := service.NewWorkspaceService(workspaceStore, userStore)

	// 创建 handler
	handler := NewWorkspaceHandler(workspaceService)

	// 生成 token
	adminToken, _ := jwtService.GenerateAccessToken(admin.ID, admin.Email, admin.Role)
	memberToken, _ := jwtService.GenerateAccessToken(member.ID, member.Email, member.Role)

	router.Use(func(c *gin.Context) {
		c.Set("db", tx)
	})
	router.Use(middleware.Auth(jwtService))
	router.PUT("/workspaces/:id", handler.UpdateWorkspace)

	tests := []struct {
		name           string
		workspaceID    string
		setupAuth      func(req *http.Request)
		body           map[string]interface{}
		wantStatusCode int
	}{
		{
			name:        "Admin 更新名称",
			workspaceID: workspace.ID.String(),
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+adminToken)
			},
			body:           map[string]interface{}{"name": "Updated Name"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "Member 无权限",
			workspaceID: workspace.ID.String(),
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+memberToken)
			},
			body:           map[string]interface{}{"name": "Should Not Update"},
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PUT", "/workspaces/"+tt.workspaceID, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			tt.setupAuth(req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}
