package handler

import (
	"bytes"
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
// CommentHandler 接口测试
// =============================================================================

func TestCommentHandler_Interface(t *testing.T) {
	var _ *CommentHandler = NewCommentHandler(nil)
}

// =============================================================================
// CreateComment 测试 (任务 8.1)
// =============================================================================

func TestCommentHandler_CreateComment(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupCommentHandlerFixtures(t, tx)
	issueID := fixtures.issue.ID.String()

	tests := []struct {
		name       string
		issueID    string
		body       interface{}
		setupAuth  bool
		wantStatus int
		checkFunc  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "成功创建评论",
			issueID: issueID,
			body: gin.H{
				"body": "这是一条测试评论",
			},
			setupAuth:  true,
			wantStatus: http.StatusCreated,
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				if resp["body"] != "这是一条测试评论" {
					t.Errorf("期望 body '这是一条测试评论', 得到 %v", resp["body"])
				}
				if resp["id"] == nil {
					t.Error("响应应包含 id")
				}
			},
		},
		{
			name:    "创建带父评论的回复",
			issueID: issueID,
			body: gin.H{
				"body":      "这是一条回复",
				"parent_id": fixtures.parentComment.ID.String(),
			},
			setupAuth:  true,
			wantStatus: http.StatusCreated,
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				if resp["parent_id"] == nil {
					t.Error("响应应包含 parent_id")
				}
			},
		},
		{
			name:    "未授权创建评论",
			issueID: issueID,
			body: gin.H{
				"body": "未授权评论",
			},
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "缺少评论内容",
			issueID: issueID,
			body: gin.H{
				// 缺少 body 字段
			},
			setupAuth:  true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "无效的 Issue ID",
			issueID: "invalid-uuid",
			body: gin.H{
				"body": "无效 Issue",
			},
			setupAuth:  true,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/issues/"+tt.issueID+"/comments", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.issueID}}

			if tt.setupAuth {
				setAuthContext(c, fixtures.userID, fixtures.userRole)
			}

			fixtures.handler.CreateComment(c)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateComment() status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, w)
			}
		})
	}
}

// =============================================================================
// ListIssueComments 测试 (任务 8.3)
// =============================================================================

func TestCommentHandler_ListIssueComments(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupCommentHandlerFixtures(t, tx)
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
			name:       "成功获取评论列表",
			issueID:    issueID,
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				comments, ok := resp["comments"].([]interface{})
				if !ok {
					t.Error("响应应包含 comments 数组")
					return
				}
				if len(comments) < 1 {
					t.Error("评论列表不应为空")
				}
			},
		},
		{
			name:       "空评论列表",
			issueID:    uuid.New().String(), // 不存在的 Issue
			setupAuth:  true,
			wantStatus: http.StatusOK,
			checkFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				comments, ok := resp["comments"].([]interface{})
				if !ok {
					t.Error("响应应包含 comments 数组")
					return
				}
				if len(comments) != 0 {
					t.Errorf("期望空列表，得到 %d 条评论", len(comments))
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/issues/" + tt.issueID + "/comments"
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

			fixtures.handler.ListIssueComments(c)

			if w.Code != tt.wantStatus {
				t.Errorf("ListIssueComments() status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, w)
			}
		})
	}
}

// =============================================================================
// UpdateComment 测试 (任务 8.5)
// =============================================================================

func TestCommentHandler_UpdateComment(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupCommentHandlerFixtures(t, tx)

	// 创建一个新评论用于更新测试
	newComment, _ := fixtures.commentService.CreateComment(
		fixtures.authCtx,
		fixtures.issue.ID,
		fixtures.userID,
		"原始评论内容",
		nil,
	)

	tests := []struct {
		name       string
		commentID  string
		body       interface{}
		setupAuth  bool
		authUser   uuid.UUID // 使用不同的用户
		wantStatus int
	}{
		{
			name:      "成功更新评论",
			commentID: newComment.ID.String(),
			body: gin.H{
				"body": "更新后的评论内容",
			},
			setupAuth:  true,
			authUser:   fixtures.userID,
			wantStatus: http.StatusOK,
		},
		{
			name:      "非作者拒绝更新",
			commentID: fixtures.parentComment.ID.String(), // 使用 otherUser 的评论
			body: gin.H{
				"body": "尝试更新他人评论",
			},
			setupAuth:  true,
			authUser:   fixtures.thirdUserID, // 使用第三个用户（非作者）
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "未授权更新",
			commentID: newComment.ID.String(),
			body: gin.H{
				"body": "未授权更新",
			},
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "评论不存在",
			commentID: uuid.New().String(),
			body: gin.H{
				"body": "不存在的评论",
			},
			setupAuth:  true,
			authUser:   fixtures.userID,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PUT", "/api/v1/comments/"+tt.commentID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "commentId", Value: tt.commentID}}

			if tt.setupAuth {
				userID := tt.authUser
				if userID == uuid.Nil {
					userID = fixtures.userID
				}
				setAuthContext(c, userID, fixtures.userRole)
			}

			fixtures.handler.UpdateComment(c)

			if w.Code != tt.wantStatus {
				t.Errorf("UpdateComment() status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================================================
// DeleteComment 测试 (任务 8.7)
// =============================================================================

func TestCommentHandler_DeleteComment(t *testing.T) {
	tx := testHandlerDB.Begin()
	defer tx.Rollback()

	fixtures := setupCommentHandlerFixtures(t, tx)

	// 为删除测试创建评论
	commentToDelete, _ := fixtures.commentService.CreateComment(
		fixtures.authCtx,
		fixtures.issue.ID,
		fixtures.userID,
		"要删除的评论",
		nil,
	)

	// 为 Admin 删除测试创建评论
	commentForAdminDelete, _ := fixtures.commentService.CreateComment(
		context.WithValue(context.WithValue(context.Background(), "user_id", fixtures.otherUserID), "user_role", model.RoleMember),
		fixtures.issue.ID,
		fixtures.otherUserID,
		"其他用户的评论",
		nil,
	)

	tests := []struct {
		name       string
		commentID  string
		setupAuth  bool
		authUser   uuid.UUID
		authRole   model.Role
		wantStatus int
	}{
		{
			name:       "成功删除自己的评论",
			commentID:  commentToDelete.ID.String(),
			setupAuth:  true,
			authUser:   fixtures.userID,
			authRole:   fixtures.userRole,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Admin 可删除他人评论",
			commentID:  commentForAdminDelete.ID.String(),
			setupAuth:  true,
			authUser:   fixtures.userID, // Admin 用户
			authRole:   model.RoleAdmin,
			wantStatus: http.StatusOK,
		},
		{
			name:       "非作者拒绝删除",
			commentID:  fixtures.parentComment.ID.String(), // otherUser 的评论
			setupAuth:  true,
			authUser:   fixtures.thirdUserID, // 第三个用户（非作者，非Admin）
			authRole:   model.RoleMember,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "未授权删除",
			commentID:  fixtures.parentComment.ID.String(),
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "评论不存在",
			commentID:  uuid.New().String(),
			setupAuth:  true,
			authUser:   fixtures.userID,
			authRole:   model.RoleAdmin,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/v1/comments/"+tt.commentID, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "commentId", Value: tt.commentID}}

			if tt.setupAuth {
				role := tt.authRole
				if role == "" {
					role = fixtures.userRole
				}
				setAuthContext(c, tt.authUser, role)
			}

			fixtures.handler.DeleteComment(c)

			if w.Code != tt.wantStatus {
				t.Errorf("DeleteComment() status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================================================
// 测试辅助结构和函数
// =============================================================================

type commentHandlerFixtures struct {
	handler        *CommentHandler
	commentService service.CommentService
	issue          *model.Issue
	parentComment  *model.Comment
	userID         uuid.UUID    // Admin 用户
	otherUserID    uuid.UUID    // 普通用户（parentComment 的作者）
	thirdUserID    uuid.UUID    // 第三个用户（非作者）
	userRole       model.Role
	authCtx        context.Context
}

func setupCommentHandlerFixtures(t *testing.T, db *gorm.DB) *commentHandlerFixtures {
	prefix := uuid.New().String()[:8]

	// 确保迁移 comments 表
	db.AutoMigrate(&model.Comment{})

	// 创建工作区
	workspace := &model.Workspace{
		Name: prefix + "_Comment WS",
		Slug: prefix + "_cmt_ws",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("创建工作区失败: %v", err)
	}

	// 创建用户（评论作者）
	userID := uuid.New()
	userRole := model.RoleAdmin
	user := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_user@example.com",
		Username:     prefix + "_user",
		Name:         "Comment User",
		PasswordHash: "hash",
		Role:         userRole,
	}
	user.ID = userID
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 创建另一个用户
	otherUserID := uuid.New()
	otherUser := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_other@example.com",
		Username:     prefix + "_other",
		Name:         "Other User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	otherUser.ID = otherUserID
	if err := db.Create(otherUser).Error; err != nil {
		t.Fatalf("创建其他用户失败: %v", err)
	}

	// 创建团队
	team := &model.Team{
		WorkspaceID: workspace.ID,
		Name:        prefix + "_Comment Team",
		Key:         "CMT",
	}
	if err := db.Create(team).Error; err != nil {
		t.Fatalf("创建团队失败: %v", err)
	}

	// 添加团队成员
	db.Create(&model.TeamMember{TeamID: team.ID, UserID: userID, Role: model.RoleAdmin})
	db.Create(&model.TeamMember{TeamID: team.ID, UserID: otherUserID, Role: model.RoleMember})

	// 创建第三个用户（非作者，用于测试权限）
	thirdUserID := uuid.New()
	thirdUser := &model.User{
		WorkspaceID:  workspace.ID,
		Email:        prefix + "_third@example.com",
		Username:     prefix + "_third",
		Name:         "Third User",
		PasswordHash: "hash",
		Role:         model.RoleMember,
	}
	thirdUser.ID = thirdUserID
	if err := db.Create(thirdUser).Error; err != nil {
		t.Fatalf("创建第三个用户失败: %v", err)
	}
	db.Create(&model.TeamMember{TeamID: team.ID, UserID: thirdUserID, Role: model.RoleMember})

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
	commentStore := store.NewCommentStore(db)
	userStore := store.NewUserStore(db)

	// 创建 services
	issueService := service.NewIssueService(issueStore, subscriptionStore, teamMemberStore)
	commentService := service.NewCommentService(commentStore, issueStore, subscriptionStore, userStore)

	// 创建 handler
	handler := NewCommentHandler(commentService)

	// 创建认证 context
	authCtx := context.Background()
	authCtx = context.WithValue(authCtx, "user_id", userID)
	authCtx = context.WithValue(authCtx, "user_role", userRole)

	// 创建 Issue
	issue, err := issueService.CreateIssue(authCtx, &service.CreateIssueParams{
		TeamID:   team.ID,
		Title:    "Test Issue for Comments",
		StatusID: status.ID,
	})
	if err != nil {
		t.Fatalf("创建 Issue 失败: %v", err)
	}

	// 创建父评论（使用 otherUser）
	otherUserCtx := context.WithValue(context.WithValue(context.Background(), "user_id", otherUserID), "user_role", model.RoleMember)
	parentComment, err := commentService.CreateComment(
		otherUserCtx,
		issue.ID,
		otherUserID,
		"这是一条父评论",
		nil,
	)
	if err != nil {
		t.Fatalf("创建父评论失败: %v", err)
	}

	return &commentHandlerFixtures{
		handler:        handler,
		commentService: commentService,
		issue:          issue,
		parentComment:  parentComment,
		userID:         userID,
		otherUserID:    otherUserID,
		thirdUserID:    thirdUserID,
		userRole:       userRole,
		authCtx:        authCtx,
	}
}
