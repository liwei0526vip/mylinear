package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantErr     bool
		wantDBURL   string
		wantRedis   string
		wantPort    string
		wantMinioEp string
	}{
		{
			name: "使用环境变量值",
			envVars: map[string]string{
				"DATABASE_URL":     "postgres://user:pass@host:5432/db",
				"REDIS_URL":        "redis://host:6379/1",
				"PORT":             "3000",
				"MINIO_ENDPOINT":   "minio.example.com:9000",
				"MINIO_ACCESS_KEY": "access",
				"MINIO_SECRET_KEY": "secret",
			},
			wantErr:     false,
			wantDBURL:   "postgres://user:pass@host:5432/db",
			wantRedis:   "redis://host:6379/1",
			wantPort:    "3000",
			wantMinioEp: "minio.example.com:9000",
		},
		{
			name:        "使用默认值",
			envVars:     map[string]string{},
			wantErr:     false,
			wantDBURL:   "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable",
			wantRedis:   "redis://localhost:6379/0",
			wantPort:    "8080",
			wantMinioEp: "localhost:9000",
		},
		{
			name: "部分环境变量覆盖",
			envVars: map[string]string{
				"PORT": "9090",
			},
			wantErr:     false,
			wantDBURL:   "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable",
			wantRedis:   "redis://localhost:6379/0",
			wantPort:    "9090",
			wantMinioEp: "localhost:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理环境变量
			os.Clearenv()

			// 设置测试环境变量
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load()

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg.DatabaseURL != tt.wantDBURL {
					t.Errorf("Load().DatabaseURL = %v, want %v", cfg.DatabaseURL, tt.wantDBURL)
				}
				if cfg.RedisURL != tt.wantRedis {
					t.Errorf("Load().RedisURL = %v, want %v", cfg.RedisURL, tt.wantRedis)
				}
				if cfg.Port != tt.wantPort {
					t.Errorf("Load().Port = %v, want %v", cfg.Port, tt.wantPort)
				}
				if cfg.MinioEndpoint != tt.wantMinioEp {
					t.Errorf("Load().MinioEndpoint = %v, want %v", cfg.MinioEndpoint, tt.wantMinioEp)
				}
			}
		})
	}
}

func TestConfig_GetMinioCredentials(t *testing.T) {
	tests := []struct {
		name            string
		envVars         map[string]string
		wantAccessKey   string
		wantSecretKey   string
		wantUseSSL      bool
		wantBucket      string
	}{
		{
			name: "从环境变量读取 MinIO 配置",
			envVars: map[string]string{
				"MINIO_ACCESS_KEY": "myaccess",
				"MINIO_SECRET_KEY": "mysecret",
				"MINIO_USE_SSL":    "true",
				"MINIO_BUCKET":     "mybucket",
			},
			wantAccessKey: "myaccess",
			wantSecretKey: "mysecret",
			wantUseSSL:    true,
			wantBucket:    "mybucket",
		},
		{
			name:            "使用默认 MinIO 配置",
			envVars:         map[string]string{},
			wantAccessKey:   "minioadmin",
			wantSecretKey:   "minioadmin",
			wantUseSSL:      false,
			wantBucket:      "mylinear",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, _ := Load()

			if cfg.MinioAccessKey != tt.wantAccessKey {
				t.Errorf("MinioAccessKey = %v, want %v", cfg.MinioAccessKey, tt.wantAccessKey)
			}
			if cfg.MinioSecretKey != tt.wantSecretKey {
				t.Errorf("MinioSecretKey = %v, want %v", cfg.MinioSecretKey, tt.wantSecretKey)
			}
			if cfg.MinioUseSSL != tt.wantUseSSL {
				t.Errorf("MinioUseSSL = %v, want %v", cfg.MinioUseSSL, tt.wantUseSSL)
			}
			if cfg.MinioBucket != tt.wantBucket {
				t.Errorf("MinioBucket = %v, want %v", cfg.MinioBucket, tt.wantBucket)
			}
		})
	}
}
