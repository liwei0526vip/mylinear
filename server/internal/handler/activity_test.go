package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
	"gorm.io/gorm"
)

// =============================================================================
// ActivityHandler 接口测试
// =============================================================================

func TestActivityHandler_Interface(t *testing.T) {
	var _ *ActivityHandler = NewActivityHandler(nil)
}

// =============================================================================
// ListIssueActivities 测试 (任务 9.1)
// =============================================================================

func TestActivityHandler_ListIssueActivities(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupActivityHandlerFixtures(t, tx)
	issueID := fixtures.issue.ID.String()

	tests := []struct {
		name       string
		issueID    string
		query      string
		setupAuth  bool
		wantStatus int
		checkFunc  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "成功获取活动列表",
			issueID:    issueID,
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				activities, ok := resp["activities"].([]interface{})
				if !ok {
					t.Error("响应应包含 activities 数组")
					return
				}
				// 应该至少有 issue_created 活动
				if len(activities) < 1 {
					t.Error("活动列表不应为空")
				}
			},
		},
		{
			name:       "空活动列表",
			issueID:    uuid.New().String(), // 不存在的 Issue
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				activities, ok := resp["activities"].([]interface{})
				if !ok {
					t.Error("响应应包含 activities 数组")
					return
				}
				if len(activities) != 0 {
					t.Errorf("期望空列表，得到 %d 条活动", len(activities))
				}
			},
		},
		{
			name:       "分页参数",
			issueID:    issueID,
			query:      "page=1&page_size=10",
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				if resp["page"] != nil {
					page := int(resp["page"].(float64))
					if page != 1 {
						t.Errorf("期望 page 1, 得到 %d", page)
					}
				}
				if resp["page_size"] != nil {
					pageSize := int(resp["page_size"].(float64))
					if pageSize != 10 {
						t.Errorf("期望 page_size 10, 得到 %d", pageSize)
					}
				}
			},
		},
		{
			name:       "按类型过滤",
			issueID:    issueID,
			query:      "types=issue_created,title_changed",
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				activities, ok := resp["activities"].([]interface{})
				if !ok {
					t.Error("响应应包含 activities 数组")
					return
				}
				// 验证返回的活动类型都是过滤的类型之一
				for _, a := range activities {
					activity := a.(map[string]interface{})
					actType := activity["type"].(string)
					if actType != "issue_created" && actType != "title_changed" {
						t.Errorf("活动类型 %s 不在过滤范围内", actType)
					}
				}
			},
		},
		{
			name:       "无效的 Issue ID",
			issueID:    "invalid-uuid",
			setupAuth:  true,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/issues/" + tt.issueID + "/activities"
			if tt.query != "" {
				path += "?" + tt.query
			}
			req := httptest.NewRequest("GET", path, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.issueID}}

			if tt.setupAuth {
				setAuthContext(c, fixtures.userID, fixtures.userRole)
			}

			fixtures.handler.ListIssueActivities(c)

			if w.Code != tt.wantStatus {
				t.Errorf("ListIssueActivities() status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, w)
			}
		})
	}
}

// =============================================================================
// 测试辅助结构和函数
// =============================================================================

type activityHandlerFixtures struct {
	handler         *ActivityHandler
	activityService service.ActivityService
	issueService    service.IssueService
	issue           *model.Issue
	userID          uuid.UUID
	userRole        model.Role
	authCtx         context.Context
}

func setupActivityHandlerFixtures(t *testing.T, db *gorm.DB) *activityHandlerFixtures {
	prefix := uuid.New().String()[:8]

	// 确保迁移 activities 表
	db.AutoMigrate(&model.Activity{})

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Activity WS",
		Slug: prefix + "_act_ws",
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
		Name:         "Activity User",
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
		Name:        prefix + "_Activity Team",
		Key:         "ACT",
	}
	if err := db.Create(team).Error; err != nil {
		t.Fatalf("创建团队失败: %v", err)
	}

	// 添加团队成员
	db.Create(&model.TeamMember{TeamID: team.ID, UserID: userID, Role: model.RoleAdmin})

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

	// 创建 stores
	issueStore := store.NewIssueStore(db)
	subscriptionStore := store.NewIssueSubscriptionStore(db)
	teamMemberStore := store.NewTeamMemberStore(db)
	activityStore := store.NewActivityStore(db)

	// 创建 services
	activityService := service.NewActivityService(activityStore)
	issueService := service.NewIssueServiceWithActivity(issueStore, subscriptionStore, teamMemberStore, activityService)

	// 创建 handler
	handler := NewActivityHandler(activityService)

	// 创建认证 context
	authCtx := context.Background()
	authCtx = context.WithValue(authCtx, "user_id", userID)
	authCtx = context.WithValue(authCtx, "user_role", userRole)

	// 创建 Issue（这会自动创建 issue_created 活动）
	issue, err := issueService.CreateIssue(authCtx, &service.CreateIssueParams{
		TeamID:   team.ID,
		Title:    "Test Issue for Activities",
		StatusID: status.ID,
	})
	if err != nil {
		t.Fatalf("创建 Issue 失败: %v", err)
	}

	// 创建一些额外的活动
	_, _ = issueService.UpdateIssue(authCtx, issue.ID.String(), map[string]interface{}{
		"title": "Updated Title",
	})
	_, _ = issueService.UpdateIssue(authCtx, issue.ID.String(), map[string]interface{}{
		"priority": 2,
	})

	return &activityHandlerFixtures{
		handler:         handler,
		activityService: activityService,
		issueService:    issueService,
		issue:           issue,
		userID:          userID,
		userRole:        userRole,
		authCtx:         authCtx,
	}
}
