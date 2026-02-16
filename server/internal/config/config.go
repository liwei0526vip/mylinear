// Package config 提供应用配置加载功能
package config

import (
	"os"
	"strconv"

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

	// MinIO 配置
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
	MinioBucket    string
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
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", defaultMinioEndpoint),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", defaultMinioAccessKey),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", defaultMinioSecretKey),
		MinioUseSSL:    getEnvBool("MINIO_USE_SSL", defaultMinioUseSSL),
		MinioBucket:    getEnv("MINIO_BUCKET", defaultMinioBucket),
	}

	return cfg, nil
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
