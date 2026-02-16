package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylinear/server/internal/config"
	"github.com/mylinear/server/internal/model"
	"github.com/mylinear/server/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestAuth_Middleware 测试认证中间件
func TestAuth_Middleware(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "middleware-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	tests := []struct {
		name           string
		setupAuth      func(req *http.Request)
		wantStatusCode int
		wantUserID     bool
	}{
		{
			name: "有效令牌通过",
			setupAuth: func(req *http.Request) {
				userID := uuid.New()
				token, _ := jwtService.GenerateAccessToken(userID, "test@example.com", model.RoleMember)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantStatusCode: http.StatusOK,
			wantUserID:     true,
		},
		{
			name: "缺少Authorization头",
			setupAuth: func(req *http.Request) {
				// 不设置 Authorization
			},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     false,
		},
		{
			name: "Bearer格式错误",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "InvalidFormat token")
			},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     false,
		},
		{
			name: "无效令牌",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid-token")
			},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Auth(jwtService))
			router.GET("/test", func(c *gin.Context) {
				userID := GetCurrentUserID(c)
				if tt.wantUserID && userID == uuid.Nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "user ID not found"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			tt.setupAuth(req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d, body: %s", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}

// TestGetCurrentUser 测试获取当前用户
func TestGetCurrentUser(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "getcurrent-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	userID := uuid.New()
	email := "getcurrent@example.com"
	role := model.RoleAdmin

	token, _ := jwtService.GenerateAccessToken(userID, email, role)

	router := gin.New()
	router.Use(Auth(jwtService))
	router.GET("/test", func(c *gin.Context) {
		// 测试 GetCurrentUser
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user is nil"})
			return
		}
		if user.UserID != userID.String() {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong user ID"})
			return
		}
		if user.Email != email {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong email"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

// TestGetCurrentUserID 测试未认证场景
func TestGetCurrentUserID_Unauthenticated(t *testing.T) {
	router := gin.New()
	// 不使用中间件
	router.GET("/test", func(c *gin.Context) {
		userID := GetCurrentUserID(c)
		if userID != uuid.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "should be nil UUID"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want %d", w.Code, http.StatusOK)
	}
}

// TestGetCurrentUserRole 测试获取当前用户角色
func TestGetCurrentUserRole(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "role-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	tests := []struct {
		name     string
		role     model.Role
		wantRole string
	}{
		{"成员角色", model.RoleMember, string(model.RoleMember)},
		{"管理员角色", model.RoleAdmin, string(model.RoleAdmin)},
		{"全局管理员角色", model.RoleGlobalAdmin, string(model.RoleGlobalAdmin)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			token, _ := jwtService.GenerateAccessToken(userID, "role@example.com", tt.role)

			router := gin.New()
			router.Use(Auth(jwtService))
			router.GET("/test", func(c *gin.Context) {
				role := GetCurrentUserRole(c)
				if role != tt.wantRole {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong role"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("状态码 = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

// TestIsAdmin 测试管理员检查
func TestIsAdmin(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "isadmin-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	tests := []struct {
		name      string
		role      model.Role
		wantAdmin bool
	}{
		{"成员不是管理员", model.RoleMember, false},
		{"管理员", model.RoleAdmin, true},
		{"全局管理员", model.RoleGlobalAdmin, true},
		{"访客不是管理员", model.RoleGuest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			token, _ := jwtService.GenerateAccessToken(userID, "isadmin@example.com", tt.role)

			router := gin.New()
			router.Use(Auth(jwtService))
			router.GET("/test", func(c *gin.Context) {
				isAdmin := IsAdmin(c)
				if isAdmin != tt.wantAdmin {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong admin status"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("状态码 = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

// TestRequireRole 测试角色要求中间件
func TestRequireRole(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "require-role-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	tests := []struct {
		name           string
		userRole       model.Role
		requiredRoles  []string
		wantStatusCode int
	}{
		{
			name:           "成员访问成员资源",
			userRole:       model.RoleMember,
			requiredRoles:  []string{string(model.RoleMember)},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "成员访问管理员资源被拒绝",
			userRole:       model.RoleMember,
			requiredRoles:  []string{string(model.RoleAdmin)},
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "管理员访问管理员资源",
			userRole:       model.RoleAdmin,
			requiredRoles:  []string{string(model.RoleAdmin)},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "全局管理员访问管理员资源",
			userRole:       model.RoleGlobalAdmin,
			requiredRoles:  []string{string(model.RoleAdmin)},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			token, _ := jwtService.GenerateAccessToken(userID, "reqrole@example.com", tt.userRole)

			router := gin.New()
			router.Use(Auth(jwtService))
			router.Use(RequireRole(tt.requiredRoles...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

// TestRequireAdmin 测试管理员中间件
func TestRequireAdmin(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "require-admin-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	tests := []struct {
		name           string
		userRole       model.Role
		wantStatusCode int
	}{
		{"成员被拒绝", model.RoleMember, http.StatusForbidden},
		{"管理员通过", model.RoleAdmin, http.StatusOK},
		{"全局管理员通过", model.RoleGlobalAdmin, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			token, _ := jwtService.GenerateAccessToken(userID, "reqadmin@example.com", tt.userRole)

			router := gin.New()
			router.Use(Auth(jwtService))
			router.Use(RequireAdmin())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

// TestRequireGlobalAdmin 测试全局管理员中间件
func TestRequireGlobalAdmin(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "require-global-admin-test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtService := service.NewJWTService(cfg)

	tests := []struct {
		name           string
		userRole       model.Role
		wantStatusCode int
	}{
		{"成员被拒绝", model.RoleMember, http.StatusForbidden},
		{"管理员被拒绝", model.RoleAdmin, http.StatusForbidden},
		{"全局管理员通过", model.RoleGlobalAdmin, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			token, _ := jwtService.GenerateAccessToken(userID, "reqglobal@example.com", tt.userRole)

			router := gin.New()
			router.Use(Auth(jwtService))
			router.Use(RequireGlobalAdmin())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("状态码 = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}
