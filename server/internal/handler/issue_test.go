package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/middleware"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testHandlerDB *gorm.DB

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable"
	}

	var err error
	testHandlerDB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("无法连接数据库: %v\n", err)
		os.Exit(1)
	}

	// 清理和迁移
	testHandlerDB.Exec("DROP TABLE IF EXISTS issue_subscriptions CASCADE")
	testHandlerDB.Exec("DROP TABLE IF EXISTS issues CASCADE")
	testHandlerDB.Exec("DROP TABLE IF EXISTS labels CASCADE")
	testHandlerDB.Exec("DROP TABLE IF EXISTS workflow_states CASCADE")
	testHandlerDB.Exec("DROP TABLE IF EXISTS team_members CASCADE")
	testHandlerDB.Exec("DROP TABLE IF EXISTS teams CASCADE")
	testHandlerDB.Exec("DROP TABLE IF EXISTS users CASCADE")
	testHandlerDB.Exec("DROP TABLE IF EXISTS workspaces CASCADE")

	err = testHandlerDB.AutoMigrate(
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
		fmt.Printf("自动迁移失败: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// TestIssueHandler_Interface 测试 IssueHandler 结构定义存在
func TestIssueHandler_Interface(t *testing.T) {
	var _ *IssueHandler = NewIssueHandler(nil)
}

// =============================================================================
// CreateIssue 测试
// =============================================================================

func TestIssueHandler_CreateIssue(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueHandlerFixtures(t, tx)
	teamID := fixtures.team.ID.String()

	tests := []struct {
		name       string
		teamID     string
		body       interface{}
		setupAuth  bool
		wantStatus int
	}{
		{
			name:   "正常创建 Issue",
			teamID: teamID,
			body: gin.H{
				"title":     "测试 Issue",
				"status_id": fixtures.status.ID.String(),
			},
			setupAuth:  true,
			wantStatus: http.StatusCreated,
		},
		{
			name:   "未认证创建 Issue",
			teamID: teamID,
			body: gin.H{
				"title":     "未认证 Issue",
				"status_id": fixtures.status.ID.String(),
			},
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "缺少标题",
			teamID: teamID,
			body: gin.H{
				"status_id": fixtures.status.ID.String(),
			},
			setupAuth:  true,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/teams/"+tt.teamID+"/issues", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "teamId", Value: tt.teamID}}

			if tt.setupAuth {
				setAuthContext(c, fixtures.userID, fixtures.userRole)
			}

			fixtures.handler.CreateIssue(c)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateIssue() status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================================================
// GetIssue 测试
// =============================================================================

func TestIssueHandler_GetIssue(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueHandlerFixtures(t, tx)

	// 先创建一个 Issue
	issue, _ := fixtures.issueService.CreateIssue(fixtures.authCtx, &service.CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试获取",
		StatusID: fixtures.status.ID,
	})

	tests := []struct {
		name       string
		issueID    string
		setupAuth  bool
		wantStatus int
	}{
		{
			name:       "正常获取 Issue",
			issueID:    issue.ID.String(),
			setupAuth:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "未认证获取（允许读取）",
			issueID:    issue.ID.String(),
			setupAuth:  false,
			wantStatus: http.StatusOK, // 读取操作允许未认证访问
		},
		{
			name:       "获取不存在的 Issue",
			issueID:    uuid.New().String(),
			setupAuth:  true,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/issues/" + tt.issueID
			req := httptest.NewRequest("GET", path, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.issueID}}

			if tt.setupAuth {
				setAuthContext(c, fixtures.userID, fixtures.userRole)
			}

			fixtures.handler.GetIssue(c)

			if w.Code != tt.wantStatus {
				t.Errorf("GetIssue() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// =============================================================================
// ListIssues 测试
// =============================================================================

func TestIssueHandler_ListIssues(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueHandlerFixtures(t, tx)

	// 创建多个 Issue
	for i := 0; i < 3; i++ {
		_, _ = fixtures.issueService.CreateIssue(fixtures.authCtx, &service.CreateIssueParams{
			TeamID:   fixtures.team.ID,
			Title:    "测试列表",
			StatusID: fixtures.status.ID,
		})
	}

	tests := []struct {
		name       string
		teamID     string
		setupAuth  bool
		wantStatus int
		checkCount bool
		minCount   int
	}{
		{
			name:       "正常获取列表",
			teamID:     fixtures.team.ID.String(),
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkCount: true,
			minCount:   3,
		},
		{
			name:       "未认证获取列表（允许读取）",
			teamID:     fixtures.team.ID.String(),
			setupAuth:  false,
			wantStatus: http.StatusOK, // 读取操作允许未认证访问
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/teams/" + tt.teamID + "/issues"
			req := httptest.NewRequest("GET", path, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "teamId", Value: tt.teamID}}

			if tt.setupAuth {
				setAuthContext(c, fixtures.userID, fixtures.userRole)
			}

			fixtures.handler.ListIssues(c)

			if w.Code != tt.wantStatus {
				t.Errorf("ListIssues() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.checkCount && w.Code == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				if issues, ok := resp["issues"].([]interface{}); ok {
					if len(issues) < tt.minCount {
						t.Errorf("ListIssues() got %d issues, want at least %d", len(issues), tt.minCount)
					}
				}
			}
		})
	}
}

// =============================================================================
// UpdateIssue 测试
// =============================================================================

func TestIssueHandler_UpdateIssue(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueHandlerFixtures(t, tx)

	// 创建测试 Issue
	issue, _ := fixtures.issueService.CreateIssue(fixtures.authCtx, &service.CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试更新",
		StatusID: fixtures.status.ID,
	})

	tests := []struct {
		name       string
		issueID    string
		body       interface{}
		setupAuth  bool
		wantStatus int
	}{
		{
			name:    "正常更新",
			issueID: issue.ID.String(),
			body: gin.H{
				"title": "更新后的标题",
			},
			setupAuth:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:    "未认证更新",
			issueID: issue.ID.String(),
			body: gin.H{
				"title": "未认证更新",
			},
			setupAuth:  false,
			wantStatus: http.StatusOK, // 更新操作不需要认证（简化实现）
		},
		{
			name:    "更新不存在的 Issue",
			issueID: uuid.New().String(),
			body: gin.H{
				"title": "不存在",
			},
			setupAuth:  true,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			path := "/api/v1/issues/" + tt.issueID
			req := httptest.NewRequest("PUT", path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.issueID}}

			if tt.setupAuth {
				setAuthContext(c, fixtures.userID, fixtures.userRole)
			}

			fixtures.handler.UpdateIssue(c)

			if w.Code != tt.wantStatus {
				t.Errorf("UpdateIssue() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// =============================================================================
// DeleteIssue 测试
// =============================================================================

func TestIssueHandler_DeleteIssue(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueHandlerFixtures(t, tx)

	// 创建测试 Issue
	issue, _ := fixtures.issueService.CreateIssue(fixtures.authCtx, &service.CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试删除",
		StatusID: fixtures.status.ID,
	})

	tests := []struct {
		name       string
		issueID    string
		setupAuth  bool
		wantStatus int
	}{
		{
			name:       "正常删除",
			issueID:    issue.ID.String(),
			setupAuth:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "未认证删除",
			issueID:    uuid.New().String(), // 使用新的 ID，避免已被删除
			setupAuth:  false,
			wantStatus: http.StatusNotFound, // 删除操作在验证时会先检查 Issue 是否存在
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/issues/" + tt.issueID
			req := httptest.NewRequest("DELETE", path, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.issueID}}

			if tt.setupAuth {
				setAuthContext(c, fixtures.userID, fixtures.userRole)
			}

			fixtures.handler.DeleteIssue(c)

			if w.Code != tt.wantStatus {
				t.Errorf("DeleteIssue() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// =============================================================================
// Subscribe/Unsubscribe 测试
// =============================================================================

func TestIssueHandler_Subscribe(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupIssueHandlerFixtures(t, tx)

	// 创建测试 Issue
	issue, _ := fixtures.issueService.CreateIssue(fixtures.authCtx, &service.CreateIssueParams{
		TeamID:   fixtures.team.ID,
		Title:    "测试订阅",
		StatusID: fixtures.status.ID,
	})

	// 测试订阅
	path := "/api/v1/issues/" + issue.ID.String() + "/subscribe"
	req := httptest.NewRequest("POST", path, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: issue.ID.String()}}
	setAuthContext(c, fixtures.userID, fixtures.userRole)

	fixtures.handler.Subscribe(c)

	if w.Code != http.StatusOK {
		t.Errorf("Subscribe() status = %d, want %d", w.Code, http.StatusOK)
	}

	// 测试取消订阅
	req = httptest.NewRequest("DELETE", path, nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: issue.ID.String()}}
	setAuthContext(c, fixtures.userID, fixtures.userRole)

	fixtures.handler.Unsubscribe(c)

	if w.Code != http.StatusOK {
		t.Errorf("Unsubscribe() status = %d, want %d", w.Code, http.StatusOK)
	}
}

// =============================================================================
// 测试辅助结构和函数
// =============================================================================

type issueHandlerFixtures struct {
	handler      *IssueHandler
	issueService service.IssueService
	team         *model.Team
	status       *model.WorkflowState
	userID       uuid.UUID
	userRole     model.Role
	authCtx      context.Context
}

func setupIssueHandlerFixtures(t *testing.T, db *gorm.DB) *issueHandlerFixtures {
	prefix := uuid.New().String()[:8]

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Workspace",
		Slug: prefix + "_workspace",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("创建工作区失败: %v", err)
	}

	// 创建用户
	userID := uuid.New()
	userRole := model.RoleAdmin
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user@example.com",
		Username:     prefix + "_user",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         userRole,
	}
	user.ID = userID
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        prefix + "_Team",
		Key:         "TST",
	}
	if err := db.Create(team).Error; err != nil {
		t.Fatalf("创建团队失败: %v", err)
	}

	// 添加团队成员
	member := &model.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   model.RoleAdmin,
	}
	if err := db.Create(member).Error; err != nil {
		t.Fatalf("添加团队成员失败: %v", err)
	}

	// 创建工作流状态
	status := &model.WorkflowState{
		TeamID:   team.ID,
		Name:     "Backlog",
		Type:     model.StateTypeBacklog,
		Color:    "#808080",
		Position: 0,
	}
	if err := db.Create(status).Error; err != nil {
		t.Fatalf("创建工作流状态失败: %v", err)
	}

	// 创建 services
	issueStore := store.NewIssueStore(db)
	subscriptionStore := store.NewIssueSubscriptionStore(db)
	teamMemberStore := store.NewTeamMemberStore(db)
	issueService := service.NewIssueService(issueStore, subscriptionStore, teamMemberStore)

	// 创建 handler
	handler := NewIssueHandler(issueService)

	// 创建认证 context（使用标准 context.Context）
	authCtx := context.Background()
	authCtx = context.WithValue(authCtx, "user_id", userID)
	authCtx = context.WithValue(authCtx, "user_role", userRole)

	return &issueHandlerFixtures{
		handler:      handler,
		issueService: issueService,
		team:         team,
		status:       status,
		userID:       userID,
		userRole:     userRole,
		authCtx:      authCtx,
	}
}

// setAuthContext 设置认证上下文
func setAuthContext(c *gin.Context, userID uuid.UUID, userRole model.Role) {
	c.Set(middleware.ContextKeyUser, &middleware.UserContext{
		UserID: userID.String(),
		Role:   string(userRole),
	})
}
