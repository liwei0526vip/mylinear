// Package main 应用入口
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/liwei0526vip/mylinear/internal/config"
	"github.com/liwei0526vip/mylinear/internal/handler"
	"github.com/liwei0526vip/mylinear/internal/middleware"
	apiRouter "github.com/liwei0526vip/mylinear/internal/router"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Printf("警告: 数据库连接失败: %v", err)
		db = nil
	}

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: parseRedisAddr(cfg.RedisURL),
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("警告: Redis 连接失败: %v", err)
	}

	// 检查数据库健康状态
	dbHealthy := db != nil
	if dbHealthy {
		sqlDB, err := db.DB()
		if err != nil {
			dbHealthy = false
		} else {
			if err := sqlDB.Ping(); err != nil {
				dbHealthy = false
			}
		}

		// 执行数据库迁移
		if dbHealthy && os.Getenv("SKIP_MIGRATION") != "true" {
			if err := runMigrations(cfg.DatabaseURL); err != nil {
				log.Printf("警告: 数据库迁移失败: %v", err)
			}
		}
	}

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Recovery())

	// 注册健康检查端点
	healthHandler := handler.NewHealthHandler(dbHealthy)
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)
	}

	// 初始化服务和处理器（仅在数据库可用时）
	if db != nil && dbHealthy {
		// 初始化 Store
		userStore := store.NewUserStore(db)
		workspaceStore := store.NewWorkspaceStore(db)
		teamStore := store.NewTeamStore(db)
		teamMemberStore := store.NewTeamMemberStore(db)
		issueStore := store.NewIssueStore(db)
		issueSubscriptionStore := store.NewIssueSubscriptionStore(db)

		// 初始化服务
		jwtService := service.NewJWTService(cfg)
		authService := service.NewAuthService(userStore, workspaceStore, jwtService, rdb, cfg)
		userService := service.NewUserService(userStore)
		workspaceService := service.NewWorkspaceService(workspaceStore, userStore)

		// Workflow Service
		workflowStateStore := store.NewWorkflowStateStore(db)
		workflowService := service.NewWorkflowService(workflowStateStore, teamStore)

		teamService := service.NewTeamService(teamStore, teamMemberStore, userStore, workflowService)
		teamMemberService := service.NewTeamMemberService(teamMemberStore, userStore, teamStore)

		// Label Service
		labelStore := store.NewLabelStore(db)
		labelService := service.NewLabelService(labelStore)

		// Issue Service
		issueService := service.NewIssueService(issueStore, issueSubscriptionStore, teamMemberStore)

		// Project Service
		projectStore := store.NewProjectStore(db)
		projectService := service.NewProjectService(projectStore, teamMemberStore, userStore)

		// 初始化 AvatarService（可选，需要 MinIO）
		var avatarService service.AvatarService
		avatarCfg := &service.AvatarConfig{
			Endpoint:      cfg.MinioEndpoint,
			AccessKey:     cfg.MinioAccessKey,
			SecretKey:     cfg.MinioSecretKey,
			BucketName:    cfg.MinioBucket,
			UseSSL:        cfg.MinioUseSSL,
			AvatarBaseURL: cfg.AvatarBaseURL,
		}
		var err error
		avatarService, err = service.NewAvatarService(avatarCfg)
		if err != nil {
			log.Printf("警告: AvatarService 初始化失败: %v，头像上传功能不可用", err)
		}

		// 初始化处理器
		authHandler := handler.NewAuthHandler(authService)
		userHandler := handler.NewUserHandlerWithAvatar(userService, avatarService)

		// 认证中间件
		authMiddleware := middleware.Auth(jwtService)

		// 注册认证路由（公开）
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/logout", authHandler.Logout)
		}

		// 注册用户路由（需认证）
		usersGroup := v1.Group("/users")
		usersGroup.Use(authMiddleware)
		{
			usersGroup.GET("/me", userHandler.GetMe)
			usersGroup.PATCH("/me", userHandler.UpdateMe)
			usersGroup.POST("/me/avatar", userHandler.UploadAvatar)
		}

		// 注册 Workspace 路由
		apiRouter.RegisterWorkspaceRoutes(v1, db, jwtService, workspaceService)

		// 注册 Team 路由
		apiRouter.RegisterTeamRoutes(v1, db, jwtService, teamService)

		// 注册 TeamMember 路由
		apiRouter.RegisterTeamMemberRoutes(v1, db, jwtService, teamMemberService)

		// 注册 Workflow 路由
		apiRouter.RegisterWorkflowRoutes(v1, db, jwtService, workflowService)

		// 注册 Label 路由
		apiRouter.RegisterLabelRoutes(v1, db, jwtService, labelService, teamStore)

		// 注册 Project 路由
		apiRouter.RegisterProjectRoutes(v1, db, jwtService, projectService)

		// 注册 Issue 路由
		apiRouter.RegisterIssueRoutes(v1, db, jwtService, issueService)
	} else {
		log.Println("警告: 数据库不可用，认证和用户 API 不可用")
	}

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// 启动服务器（非阻塞）
	go func() {
		log.Printf("服务器启动在端口 %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号进行优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 给服务器 5 秒时间完成正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	}

	// 关闭 Redis 连接
	if err := rdb.Close(); err != nil {
		log.Printf("关闭 Redis 连接失败: %v", err)
	}

	// 关闭数据库连接
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("关闭数据库连接失败: %v", err)
			}
		}
	}

	log.Println("服务器已关闭")
}

// parseRedisAddr 从 Redis URL 解析地址
func parseRedisAddr(redisURL string) string {
	// 简单解析 redis://host:port/db 格式
	// 默认返回 localhost:6379
	if redisURL == "" {
		return "localhost:6379"
	}

	// 移除 redis:// 前缀
	var addr string
	if len(redisURL) > 8 && redisURL[:8] == "redis://" {
		addr = redisURL[8:]
	} else {
		addr = redisURL
	}

	// 移除末尾的数据库编号 (/0, /1 等)
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == '/' {
			addr = addr[:i]
			break
		}
		if addr[i] == ':' {
			break
		}
	}

	if addr == "" {
		return "localhost:6379"
	}

	return addr
}

// runMigrations 执行数据库迁移
func runMigrations(databaseURL string) error {
	// 获取迁移文件路径
	// 优先使用环境变量 MIGRATIONS_PATH，否则使用默认路径
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		// 默认迁移路径（相对于工作目录）
		migrationsPath = "migrations"
	}

	// 创建迁移实例
	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}
	defer m.Close()

	// 执行迁移
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("执行迁移失败: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		log.Printf("数据库迁移完成，无法获取版本: %v", err)
	} else {
		log.Printf("数据库迁移完成，当前版本: %d, dirty: %v", version, dirty)
	}

	return nil
}
