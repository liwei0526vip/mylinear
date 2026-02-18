package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// notificationHandlerFixtures 测试 fixtures
type notificationHandlerFixtures struct {
	handler              *NotificationHandler
	preferenceHandler    *NotificationPreferenceHandler
	notificationService  service.NotificationService
	preferenceService    service.NotificationPreferenceService
	userID               uuid.UUID
	user2ID              uuid.UUID
	userRole             model.Role
	authCtx              context.Context
	workspaceID          uuid.UUID
}

// setupNotificationHandlerFixtures 初始化测试环境
func setupNotificationHandlerFixtures(t *testing.T, db *gorm.DB) *notificationHandlerFixtures {
	prefix := uuid.New().String()[:8]

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Workspace",
		Slug: prefix + "_workspace",
	}
	require.NoError(t, db.Create(workspace).Error)

	// 创建用户
	userID := uuid.New()
	userRole := model.RoleMember
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user@example.com",
		Username:     prefix + "_user",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         userRole,
	}
	user.ID = userID
	require.NoError(t, db.Create(user).Error)

	// 创建第二个用户
	user2ID := uuid.New()
	user2 := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user2@example.com",
		Username:     prefix + "_user2",
		Name:         "Test User 2",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	user2.ID = user2ID
	require.NoError(t, db.Create(user2).Error)

	// 创建 services
	notificationStore := store.NewNotificationStore(db)
	preferenceStore := store.NewNotificationPreferenceStore(db)
	userStore := store.NewUserStore(db)

	notificationService := service.NewNotificationService(notificationStore, preferenceStore, userStore)
	preferenceService := service.NewNotificationPreferenceService(preferenceStore)

	// 创建 handlers
	handler := NewNotificationHandler(notificationService)
	preferenceHandler := NewNotificationPreferenceHandler(preferenceService)

	// 创建认证上下文
	authCtx := context.Background()
	authCtx = context.WithValue(authCtx, "user_id", userID)
	authCtx = context.WithValue(authCtx, "user_role", userRole)

	return &notificationHandlerFixtures{
		handler:             handler,
		preferenceHandler:   preferenceHandler,
		notificationService: notificationService,
		preferenceService:   preferenceService,
		userID:              userID,
		user2ID:             user2ID,
		userRole:            userRole,
		authCtx:             authCtx,
		workspaceID:         workspace.ID,
	}
}

// createTestNotification 创建测试通知
func (f *notificationHandlerFixtures) createTestNotification(t *testing.T, userID uuid.UUID) *model.Notification {
	issueID := uuid.New()
	notification := &model.Notification{
		UserID:       userID,
		Type:         model.NotificationTypeIssueAssigned,
		Title:        "测试通知",
		ResourceType: "issue",
		ResourceID:   &issueID,
	}
	require.NoError(t, f.notificationService.CreateNotification(f.authCtx, notification))
	return notification
}

// =============================================================================
// 4.1 通知 API Handler 测试
// =============================================================================

// TestNotificationHandler_ListNotifications 测试获取通知列表
func TestNotificationHandler_ListNotifications(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	f := setupNotificationHandlerFixtures(t, tx)

	// 创建测试通知
	for i := 0; i < 5; i++ {
		f.createTestNotification(t, f.userID)
	}

	tests := []struct {
		name       string
		page       string
		pageSize   string
		read       string
		setupAuth  bool
		wantStatus int
		checkFunc  func(t *testing.T, body []byte)
	}{
		{
			name:       "正常获取通知列表",
			page:       "1",
			pageSize:   "10",
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkFunc: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				json.Unmarshal(body, &resp)
				notifications := resp["notifications"].([]interface{})
				assert.Equal(t, 5, len(notifications))
				assert.Equal(t, float64(5), resp["total"])
			},
		},
		{
			name:       "分页测试",
			page:       "1",
			pageSize:   "2",
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkFunc: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				json.Unmarshal(body, &resp)
				notifications := resp["notifications"].([]interface{})
				assert.Equal(t, 2, len(notifications))
				assert.Equal(t, float64(5), resp["total"])
			},
		},
		{
			name:       "未读过滤",
			page:       "1",
			pageSize:   "10",
			read:       "false",
			setupAuth:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "未授权",
			page:       "1",
			pageSize:   "10",
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/api/v1/notifications?page=%s&page_size=%s", tt.page, tt.pageSize)
			if tt.read != "" {
				path += "&read=" + tt.read
			}
			req := httptest.NewRequest("GET", path, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.setupAuth {
				setAuthContext(c, f.userID, f.userRole)
			}

			f.handler.ListNotifications(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.checkFunc != nil && w.Code == http.StatusOK {
				tt.checkFunc(t, w.Body.Bytes())
			}
		})
	}
}

// TestNotificationHandler_GetUnreadCount 测试获取未读数量
func TestNotificationHandler_GetUnreadCount(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	f := setupNotificationHandlerFixtures(t, tx)

	// 创建 3 条未读通知
	for i := 0; i < 3; i++ {
		f.createTestNotification(t, f.userID)
	}

	tests := []struct {
		name       string
		setupAuth  bool
		wantStatus int
		wantCount  int64
	}{
		{
			name:       "正常获取未读数量",
			setupAuth:  true,
			wantStatus: http.StatusOK,
			wantCount:  3,
		},
		{
			name:       "未授权",
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/notifications/unread-count", nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.setupAuth {
				setAuthContext(c, f.userID, f.userRole)
			}

			f.handler.GetUnreadCount(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, float64(tt.wantCount), resp["count"])
			}
		})
	}
}

// TestNotificationHandler_MarkAsRead 测试标记单条已读
func TestNotificationHandler_MarkAsRead(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	f := setupNotificationHandlerFixtures(t, tx)

	// 创建测试通知
	notification := f.createTestNotification(t, f.userID)
	otherUserNotification := f.createTestNotification(t, f.user2ID)

	tests := []struct {
		name           string
		notificationID string
		setupAuth      bool
		wantStatus     int
	}{
		{
			name:           "正常标记已读",
			notificationID: notification.ID.String(),
			setupAuth:      true,
			wantStatus:     http.StatusOK,
		},
		{
			name:           "标记他人通知",
			notificationID: otherUserNotification.ID.String(),
			setupAuth:      true,
			wantStatus:     http.StatusNotFound,
		},
		{
			name:           "不存在通知",
			notificationID: uuid.New().String(),
			setupAuth:      true,
			wantStatus:     http.StatusNotFound,
		},
		{
			name:           "未授权",
			notificationID: notification.ID.String(),
			setupAuth:      false,
			wantStatus:     http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为"正常标记已读"测试创建新通知
			if tt.name == "正常标记已读" {
				newNotification := f.createTestNotification(t, f.userID)
				tt.notificationID = newNotification.ID.String()
			}

			path := "/api/v1/notifications/" + tt.notificationID + "/read"
			req := httptest.NewRequest("POST", path, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.notificationID}}

			if tt.setupAuth {
				setAuthContext(c, f.userID, f.userRole)
			}

			f.handler.MarkAsRead(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestNotificationHandler_MarkAllAsRead 测试标记全部已读
func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	f := setupNotificationHandlerFixtures(t, tx)

	tests := []struct {
		name        string
		setupAuth   bool
		wantStatus  int
		createCount int
		wantMarked  int64
	}{
		{
			name:        "正常标记全部已读",
			setupAuth:   true,
			wantStatus:  http.StatusOK,
			createCount: 5,
			wantMarked:  5,
		},
		{
			name:        "无未读通知",
			setupAuth:   true,
			wantStatus:  http.StatusOK,
			createCount: 0,
			wantMarked:  0,
		},
		{
			name:       "未授权",
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试通知
			for i := 0; i < tt.createCount; i++ {
				f.createTestNotification(t, f.userID)
			}

			req := httptest.NewRequest("POST", "/api/v1/notifications/read-all", nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.setupAuth {
				setAuthContext(c, f.userID, f.userRole)
			}

			f.handler.MarkAllAsRead(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, float64(tt.wantMarked), resp["marked"])
			}
		})
	}
}

// TestNotificationHandler_MarkBatchAsRead 测试批量标记已读
func TestNotificationHandler_MarkBatchAsRead(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	f := setupNotificationHandlerFixtures(t, tx)

	// 创建测试通知
	var notificationIDs []string
	for i := 0; i < 5; i++ {
		notification := f.createTestNotification(t, f.userID)
		notificationIDs = append(notificationIDs, notification.ID.String())
	}

	tests := []struct {
		name       string
		body       interface{}
		setupAuth  bool
		wantStatus int
		wantMarked int64
	}{
		{
			name: "正常批量标记",
			body: gin.H{
				"ids": notificationIDs[:3],
			},
			setupAuth:  true,
			wantStatus: http.StatusOK,
			wantMarked: 3,
		},
		{
			name: "部分不存在的ID",
			body: gin.H{
				"ids": append(notificationIDs[3:], uuid.New().String()),
			},
			setupAuth:  true,
			wantStatus: http.StatusOK,
			wantMarked: 2,
		},
		{
			name: "空ID列表",
			body: gin.H{
				"ids": []string{},
			},
			setupAuth:  true,
			wantStatus: http.StatusOK,
			wantMarked: 0,
		},
		{
			name: "未授权",
			body: gin.H{
				"ids": notificationIDs[:1],
			},
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/notifications/batch-read", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.setupAuth {
				setAuthContext(c, f.userID, f.userRole)
			}

			f.handler.MarkBatchAsRead(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, float64(tt.wantMarked), resp["marked"])
			}
		})
	}
}

// =============================================================================
// 4.2 通知配置 API Handler 测试
// =============================================================================

// TestNotificationPreferenceHandler_GetPreferences 测试获取通知配置
func TestNotificationPreferenceHandler_GetPreferences(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	f := setupNotificationHandlerFixtures(t, tx)

	tests := []struct {
		name       string
		channel    string
		setupAuth  bool
		wantStatus int
	}{
		{
			name:       "正常获取配置",
			setupAuth:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "按渠道过滤",
			channel:    "in_app",
			setupAuth:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "未授权",
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/notification-preferences"
			if tt.channel != "" {
				path += "?channel=" + tt.channel
			}
			req := httptest.NewRequest("GET", path, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.setupAuth {
				setAuthContext(c, f.userID, f.userRole)
			}

			f.preferenceHandler.GetPreferences(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestNotificationPreferenceHandler_UpdatePreferences 测试更新通知配置
func TestNotificationPreferenceHandler_UpdatePreferences(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	f := setupNotificationHandlerFixtures(t, tx)

	tests := []struct {
		name       string
		body       interface{}
		setupAuth  bool
		wantStatus int
	}{
		{
			name: "单个更新",
			body: gin.H{
				"preferences": []gin.H{
					{
						"channel": "in_app",
						"type":    "issue_assigned",
						"enabled": false,
					},
				},
			},
			setupAuth:  true,
			wantStatus: http.StatusOK,
		},
		{
			name: "批量更新",
			body: gin.H{
				"preferences": []gin.H{
					{
						"channel": "in_app",
						"type":    "issue_assigned",
						"enabled": true,
					},
					{
						"channel": "in_app",
						"type":    "issue_mentioned",
						"enabled": false,
					},
				},
			},
			setupAuth:  true,
			wantStatus: http.StatusOK,
		},
		{
			name: "MVP仅支持in_app渠道",
			body: gin.H{
				"preferences": []gin.H{
					{
						"channel": "email",
						"type":    "issue_assigned",
						"enabled": true,
					},
				},
			},
			setupAuth:  true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "未授权",
			body: gin.H{
				"preferences": []gin.H{},
			},
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PUT", "/api/v1/notification-preferences", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.setupAuth {
				setAuthContext(c, f.userID, f.userRole)
			}

			f.preferenceHandler.UpdatePreferences(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// =============================================================================
// 4.3 路由注册测试
// =============================================================================

// TestNotificationRoutes_Registration 测试路由注册
func TestNotificationRoutes_Registration(t *testing.T) {
	router := gin.New()

	// 创建 handler
	notificationService := service.NewNotificationService(nil, nil, nil)
	preferenceService := service.NewNotificationPreferenceService(nil)
	notificationHandler := NewNotificationHandler(notificationService)
	preferenceHandler := NewNotificationPreferenceHandler(preferenceService)

	// 注册路由
	api := router.Group("/api/v1")
	notificationHandler.RegisterRoutes(api)
	preferenceHandler.RegisterRoutes(api)

	// 验证路由注册
	routes := router.Routes()

	expectedRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/notifications"},
		{"GET", "/api/v1/notifications/unread-count"},
		{"POST", "/api/v1/notifications/:id/read"},
		{"POST", "/api/v1/notifications/read-all"},
		{"POST", "/api/v1/notifications/batch-read"},
		{"GET", "/api/v1/notification-preferences"},
		{"PUT", "/api/v1/notification-preferences"},
	}

	for _, expected := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Method == expected.method && route.Path == expected.path {
				found = true
				break
			}
		}
		assert.True(t, found, "Route %s %s should be registered", expected.method, expected.path)
	}
}
