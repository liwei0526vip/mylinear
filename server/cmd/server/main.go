// Package main 应用入口
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mylinear/server/internal/config"
	"github.com/mylinear/server/internal/handler"
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
