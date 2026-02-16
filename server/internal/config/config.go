// Package config 提供应用配置加载功能
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config 应用配置结构体
type Config struct {
	// 数据库配置
	DatabaseURL string

	// Redis 配置
	RedisURL string

	// 服务配置
	Port string
	GinMode string

	// MinIO 配置
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioUseSSL     bool
	MinioBucket     string
	AvatarBaseURL   string

	// JWT 配置
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
}

// 默认配置值
const (
	defaultDatabaseURL    = "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable"
	defaultRedisURL       = "redis://localhost:6379/0"
	defaultPort           = "8080"
	defaultMinioEndpoint  = "localhost:9000"
	defaultMinioAccessKey = "minioadmin"
	defaultMinioSecretKey = "minioadmin"
	defaultMinioUseSSL    = false
	defaultMinioBucket    = "mylinear"
	defaultAvatarBaseURL  = "http://localhost:9000" // 注意：URL 不应包含 bucket 名称，bucket 会在代码中追加

	// JWT 默认配置
	defaultJWTSecret        = ""
	defaultJWTAccessExpiry  = 15 * time.Minute
	defaultJWTRefreshExpiry = 7 * 24 * time.Hour
)

// Load 加载配置，优先从环境变量读取，缺失时使用默认值
func Load() (*Config, error) {
	// 加载 .env 文件（如果存在）
	// 在 Docker 容器环境中，.env 文件可能不存在，godotenv 会静默跳过
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", defaultDatabaseURL),
		RedisURL:       getEnv("REDIS_URL", defaultRedisURL),
		Port:           getEnv("PORT", defaultPort),
		GinMode:        getEnv("GIN_MODE", "debug"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", defaultMinioEndpoint),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", defaultMinioAccessKey),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", defaultMinioSecretKey),
		MinioUseSSL:    getEnvBool("MINIO_USE_SSL", defaultMinioUseSSL),
		MinioBucket:    getEnv("MINIO_BUCKET", defaultMinioBucket),
		AvatarBaseURL:  getEnv("AVATAR_BASE_URL", defaultAvatarBaseURL),
		JWTSecret:      getEnv("JWT_SECRET", defaultJWTSecret),
	}

	// 解析 JWT 过期时间配置
	cfg.JWTAccessExpiry = getEnvDuration("JWT_ACCESS_EXPIRY", defaultJWTAccessExpiry)
	cfg.JWTRefreshExpiry = getEnvDuration("JWT_REFRESH_EXPIRY", defaultJWTRefreshExpiry)

	return cfg, nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 生产环境必须设置 JWT_SECRET
	if c.GinMode == "release" && c.JWTSecret == "" {
		return fmt.Errorf("生产环境必须设置 JWT_SECRET 环境变量")
	}

	// JWT_REFRESH_EXPIRY 必须大于 JWT_ACCESS_EXPIRY
	if c.JWTRefreshExpiry <= c.JWTAccessExpiry {
		return fmt.Errorf("JWT_REFRESH_EXPIRY (%v) 必须大于 JWT_ACCESS_EXPIRY (%v)", c.JWTRefreshExpiry, c.JWTAccessExpiry)
	}

	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool 获取布尔类型环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolValue
	}
	return defaultValue
}

// getEnvDuration 获取时间持续时间类型环境变量
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}
		return duration
	}
	return defaultValue
}
