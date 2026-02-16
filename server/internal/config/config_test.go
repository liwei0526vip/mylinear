package config

import (
	"os"
	"testing"
	"time"
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

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		ginMode     string
		wantErr     bool
		errContains string
	}{
		{
			name: "开发环境不需要JWT_SECRET",
			envVars: map[string]string{
				"GIN_MODE": "debug",
			},
			wantErr: false,
		},
		{
			name: "生产环境必须设置JWT_SECRET",
			envVars: map[string]string{
				"GIN_MODE": "release",
			},
			ginMode:     "release",
			wantErr:     true,
			errContains: "JWT_SECRET",
		},
		{
			name: "生产环境设置了JWT_SECRET则通过",
			envVars: map[string]string{
				"GIN_MODE":   "release",
				"JWT_SECRET": "my-production-secret",
			},
			wantErr: false,
		},
		{
			name: "JWT_REFRESH_EXPIRY必须大于JWT_ACCESS_EXPIRY",
			envVars: map[string]string{
				"JWT_ACCESS_EXPIRY":  "24h",
				"JWT_REFRESH_EXPIRY": "1h",
			},
			wantErr:     true,
			errContains: "JWT_REFRESH_EXPIRY",
		},
		{
			name: "JWT_REFRESH_EXPIRY大于JWT_ACCESS_EXPIRY则通过",
			envVars: map[string]string{
				"JWT_ACCESS_EXPIRY":  "15m",
				"JWT_REFRESH_EXPIRY": "24h",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// 设置 GIN_MODE（用于 Validate 判断）
			if tt.ginMode != "" {
				// 将通过 cfg 中的某个字段或方法来判断
			}

			err = cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errContains != "" {
				if !containsString(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

// 辅助函数：检查字符串是否包含子串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConfig_JWTConfig(t *testing.T) {
	tests := []struct {
		name                 string
		envVars              map[string]string
		wantSecret           string
		wantAccessExpiry     time.Duration
		wantRefreshExpiry    time.Duration
	}{
		{
			name: "使用默认 JWT 配置",
			envVars: map[string]string{},
			wantSecret: "",
			wantAccessExpiry: 15 * time.Minute,
			wantRefreshExpiry: 7 * 24 * time.Hour,
		},
		{
			name: "从环境变量读取 JWT 配置",
			envVars: map[string]string{
				"JWT_SECRET":        "my-super-secret-key",
				"JWT_ACCESS_EXPIRY": "30m",
				"JWT_REFRESH_EXPIRY": "24h",
			},
			wantSecret: "my-super-secret-key",
			wantAccessExpiry: 30 * time.Minute,
			wantRefreshExpiry: 24 * time.Hour,
		},
		{
			name: "JWT 过期时间使用不同单位",
			envVars: map[string]string{
				"JWT_ACCESS_EXPIRY":  "1h30m",
				"JWT_REFRESH_EXPIRY": "168h",
			},
			wantSecret: "",
			wantAccessExpiry: 90 * time.Minute,
			wantRefreshExpiry: 168 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.JWTSecret != tt.wantSecret {
				t.Errorf("JWTSecret = %v, want %v", cfg.JWTSecret, tt.wantSecret)
			}
			if cfg.JWTAccessExpiry != tt.wantAccessExpiry {
				t.Errorf("JWTAccessExpiry = %v, want %v", cfg.JWTAccessExpiry, tt.wantAccessExpiry)
			}
			if cfg.JWTRefreshExpiry != tt.wantRefreshExpiry {
				t.Errorf("JWTRefreshExpiry = %v, want %v", cfg.JWTRefreshExpiry, tt.wantRefreshExpiry)
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
