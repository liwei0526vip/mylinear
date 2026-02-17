package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMigrationFilesExist 测试迁移文件存在且格式正确
func TestMigrationFilesExist(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"init up migration", "000001_init.up.sql"},
		{"init down migration", "000001_init.down.sql"},
		{"core up migration", "000002_core.up.sql"},
		{"core down migration", "000002_core.down.sql"},
		{"auxiliary up migration", "000003_auxiliary.up.sql"},
		{"auxiliary down migration", "000003_auxiliary.down.sql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(".", tt.filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("迁移文件不存在: %s", tt.filename)
			}
		})
	}
}

// TestMigrationFileFormat 测试迁移文件格式
func TestMigrationFileFormat(t *testing.T) {
	files, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatalf("无法读取迁移文件: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("没有找到迁移文件")
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			// 检查文件名格式
			base := filepath.Base(file)
			if !strings.HasSuffix(base, ".up.sql") && !strings.HasSuffix(base, ".down.sql") {
				t.Errorf("迁移文件名格式不正确: %s (应为 .up.sql 或 .down.sql)", base)
			}

			// 读取文件内容
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("无法读取文件: %v", err)
			}

			contentStr := string(content)

			// 检查 SQL 注释头
			if !strings.HasPrefix(contentStr, "--") {
				t.Errorf("迁移文件应该以 SQL 注释开头: %s", base)
			}

			// 检查是否有内容
			if len(strings.TrimSpace(contentStr)) < 50 {
				t.Errorf("迁移文件内容过短: %s", base)
			}
		})
	}
}

// TestUpAndDownMigrationPairs 测试每个 up 迁移都有对应的 down 迁移
func TestUpAndDownMigrationPairs(t *testing.T) {
	files, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatalf("无法读取迁移文件: %v", err)
	}

	upMigrations := make(map[string]bool)
	downMigrations := make(map[string]bool)

	for _, file := range files {
		base := filepath.Base(file)
		if strings.HasSuffix(base, ".up.sql") {
			name := strings.TrimSuffix(base, ".up.sql")
			upMigrations[name] = true
		} else if strings.HasSuffix(base, ".down.sql") {
			name := strings.TrimSuffix(base, ".down.sql")
			downMigrations[name] = true
		}
	}

	// 检查每个 up 迁移都有对应的 down 迁移
	for name := range upMigrations {
		if !downMigrations[name] {
			t.Errorf("缺少 down 迁移: %s.down.sql", name)
		}
	}

	// 检查每个 down 迁移都有对应的 up 迁移
	for name := range downMigrations {
		if !upMigrations[name] {
			t.Errorf("缺少 up 迁移: %s.up.sql", name)
		}
	}
}

// TestMigrationVersionOrder 测试迁移版本号顺序
func TestMigrationVersionOrder(t *testing.T) {
	expectedVersions := []string{"000001", "000002", "000003", "000004"}

	for i, version := range expectedVersions {
		upFile := version + "_*.up.sql"
		matches, err := filepath.Glob(upFile)
		if err != nil {
			t.Fatalf("无法匹配文件: %v", err)
		}
		if len(matches) == 0 {
			t.Errorf("缺少版本 %s 的 up 迁移文件", version)
		}
		if len(matches) > 1 {
			t.Errorf("版本 %s 有多个 up 迁移文件: %v", version, matches)
		}

		// 检查版本号是否递增
		if i > 0 {
			if version <= expectedVersions[i-1] {
				t.Errorf("版本号顺序错误: %s 应该在 %s 之后", version, expectedVersions[i-1])
			}
		}
	}
}

// TestSQLSyntaxBasic 测试基本 SQL 语法
func TestSQLSyntaxBasic(t *testing.T) {
	files, err := filepath.Glob("*.up.sql")
	if err != nil {
		t.Fatalf("无法读取迁移文件: %v", err)
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("无法读取文件: %v", err)
			}

			contentStr := string(content)
			contentUpper := strings.ToUpper(contentStr)

			// 检查是否包含基本的 SQL 语句（CREATE TABLE 或 ALTER TABLE）
			hasCreate := strings.Contains(contentUpper, "CREATE TABLE")
			hasAlter := strings.Contains(contentUpper, "ALTER TABLE")
			if !hasCreate && !hasAlter {
				t.Errorf("up 迁移文件应该包含 CREATE TABLE 或 ALTER TABLE 语句: %s", filepath.Base(file))
			}

			// 检查是否有 UUID 主键（仅对 CREATE TABLE）
			if hasCreate && !strings.Contains(contentStr, "UUID PRIMARY KEY") {
				t.Errorf("表应该使用 UUID 主键: %s", filepath.Base(file))
			}
		})
	}
}

// TestDownMigrationSyntax 测试 down 迁移语法
func TestDownMigrationSyntax(t *testing.T) {
	files, err := filepath.Glob("*.down.sql")
	if err != nil {
		t.Fatalf("无法读取迁移文件: %v", err)
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("无法读取文件: %v", err)
			}

			contentStr := string(content)
			contentUpper := strings.ToUpper(contentStr)

			// 检查是否包含 DROP TABLE 或 ALTER TABLE（用于删除列）语句
			hasDropTable := strings.Contains(contentUpper, "DROP TABLE")
			hasAlterTable := strings.Contains(contentUpper, "ALTER TABLE")
			if !hasDropTable && !hasAlterTable {
				t.Errorf("down 迁移文件应该包含 DROP TABLE 或 ALTER TABLE 语句: %s", filepath.Base(file))
			}
		})
	}
}
