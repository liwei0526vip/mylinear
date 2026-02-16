package service

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// 检查 MinIO 是否可用
func minioAvailable() bool {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	return endpoint != ""
}

func TestAvatarService_UploadAvatar_Success(t *testing.T) {
	if !minioAvailable() {
		t.Skip("MinIO 服务不可用，跳过测试")
	}

	// 创建 AvatarService
	cfg := &AvatarConfig{
		Endpoint:        os.Getenv("MINIO_ENDPOINT"),
		AccessKey:       os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey:       os.Getenv("MINIO_SECRET_KEY"),
		BucketName:      "test-avatars",
		UseSSL:          os.Getenv("MINIO_USE_SSL") == "true",
		AvatarBaseURL:   os.Getenv("MINIO_AVATAR_BASE_URL"),
	}

	avatarService, err := NewAvatarService(cfg)
	if err != nil {
		t.Fatalf("创建 AvatarService 失败: %v", err)
	}

	// 创建测试图片数据（1x1 PNG）
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}

	userID := uuid.New().String()
	reader := bytes.NewReader(pngData)

	// 上传头像
	avatarURL, err := avatarService.UploadAvatar(context.Background(), userID, reader, "test.png", "image/png")
	if err != nil {
		t.Errorf("上传失败: %v", err)
		return
	}

	// 验证返回的 URL
	if avatarURL == "" {
		t.Error("avatarURL 不应为空")
	}

	// 验证路径格式：avatars/{user_id}/{uuid}.png
	if !strings.Contains(avatarURL, "avatars/"+userID) {
		t.Errorf("avatarURL 路径格式不正确: %s", avatarURL)
	}
}

func TestAvatarService_ValidateMagicNumber(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		ext       string
		wantValid bool
	}{
		{
			name:      "有效的 PNG",
			data:      []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			ext:       ".png",
			wantValid: true,
		},
		{
			name:      "有效的 JPEG (FFD8FF)",
			data:      []byte{0xFF, 0xD8, 0xFF, 0xE0},
			ext:       ".jpg",
			wantValid: true,
		},
		{
			name:      "有效的 GIF",
			data:      []byte("GIF89a"),
			ext:       ".gif",
			wantValid: true,
		},
		{
			name:      "无效的文件头",
			data:      []byte{0x00, 0x00, 0x00, 0x00},
			ext:       ".png",
			wantValid: false,
		},
		{
			name:      "空文件",
			data:      []byte{},
			ext:       ".png",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateMagicNumber(tt.data, tt.ext)
			if valid != tt.wantValid {
				t.Errorf("validateMagicNumber() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestAvatarService_GenerateStoragePath(t *testing.T) {
	userID := uuid.New().String()

	path := generateAvatarPath(userID, "test.png")

	// 验证路径格式
	if !strings.HasPrefix(path, "avatars/"+userID+"/") {
		t.Errorf("路径前缀不正确: %s", path)
	}

	// 验证扩展名
	if !strings.HasSuffix(path, ".png") {
		t.Errorf("路径扩展名不正确: %s", path)
	}

	// 验证路径中间包含 UUID
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		t.Errorf("路径格式不正确: %s", path)
		return
	}

	// 中间部分应该是有效的 UUID
	middlePart := parts[2]
	uuidPart := strings.TrimSuffix(middlePart, ".png")
	if _, err := uuid.Parse(uuidPart); err != nil {
		t.Errorf("路径中的 UUID 无效: %s", uuidPart)
	}
}

func TestAvatarService_GetExtension(t *testing.T) {
	tests := []struct {
		filename  string
		wantExt   string
	}{
		{"test.png", ".png"},
		{"avatar.JPG", ".jpg"},
		{"image.JPEG", ".jpeg"},
		{"photo.gif", ".gif"},
		{"picture.WebP", ".webp"},
		{"noextension", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			ext := getFileExtension(tt.filename)
			if ext != tt.wantExt {
				t.Errorf("getFileExtension(%s) = %s, want %s", tt.filename, ext, tt.wantExt)
			}
		})
	}
}
