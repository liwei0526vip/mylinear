// Package service 提供业务逻辑服务
package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// AvatarConfig 头像服务配置
type AvatarConfig struct {
	Endpoint      string // MinIO 端点地址
	AccessKey     string // 访问密钥
	SecretKey     string // 秘密密钥
	BucketName    string // 存储桶名称
	UseSSL        bool   // 是否使用 SSL
	AvatarBaseURL string // 头像访问基础 URL
}

// AvatarService 头像服务接口
type AvatarService interface {
	// UploadAvatar 上传头像
	UploadAvatar(ctx context.Context, userID string, file io.Reader, filename string, contentType string) (string, error)
}

// avatarService 头像服务实现
type avatarService struct {
	client    *minio.Client
	bucket    string
	baseURL   string
}

// NewAvatarService 创建头像服务
func NewAvatarService(cfg *AvatarConfig) (AvatarService, error) {
	// 创建 MinIO 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 MinIO 客户端失败: %w", err)
	}

	// 确保存储桶存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("创建存储桶失败: %w", err)
		}
	}

	return &avatarService{
		client:  client,
		bucket:  cfg.BucketName,
		baseURL: cfg.AvatarBaseURL,
	}, nil
}

// UploadAvatar 上传头像
func (s *avatarService) UploadAvatar(ctx context.Context, userID string, file io.Reader, filename string, contentType string) (string, error) {
	// 获取文件扩展名
	ext := getFileExtension(filename)

	// 读取文件前 16 字节用于 magic number 验证（WebP 需要至少 12 字节）
	buf := make([]byte, 16)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	buf = buf[:n]

	// 验证 magic number
	if !validateMagicNumber(buf, ext) {
		return "", fmt.Errorf("文件内容与扩展名不匹配")
	}

	// 组合原始数据和剩余数据
	reader := io.MultiReader(bytes.NewReader(buf), file)

	// 生成存储路径
	objectName := generateAvatarPath(userID, filename)

	// 上传到 MinIO
	_, err = s.client.PutObject(ctx, s.bucket, objectName, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}

	// 返回头像 URL
	avatarURL := fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucket, objectName)
	return avatarURL, nil
}

// validateMagicNumber 验证文件的 magic number
func validateMagicNumber(data []byte, ext string) bool {
	if len(data) < 2 {
		return false
	}

	switch strings.ToLower(ext) {
	case ".png":
		// PNG: 89 50 4E 47 0D 0A 1A 0A
		if len(data) < 8 {
			return false
		}
		return data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47
	case ".jpg", ".jpeg":
		// JPEG: FF D8 FF
		return data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
	case ".gif":
		// GIF: GIF87a or GIF89a
		if len(data) < 6 {
			return false
		}
		return string(data[0:3]) == "GIF"
	case ".webp":
		// WebP: RIFF....WEBP
		if len(data) < 12 {
			return false
		}
		return string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
	default:
		return false
	}
}

// generateAvatarPath 生成头像存储路径
// 格式: avatars/{user_id}/{uuid}.{ext}
func generateAvatarPath(userID string, filename string) string {
	ext := getFileExtension(filename)
	objectUUID := uuid.New().String()
	return fmt.Sprintf("avatars/%s/%s%s", userID, objectUUID, ext)
}

// getFileExtension 获取文件扩展名（小写）
func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}
