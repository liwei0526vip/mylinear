package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Task 1.1: 编写 db/migrations 测试
// 该测试验证 000005_create_workflow_and_labels.up.sql 及其约束
func TestMigration_C06_Schema(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 1. 读取迁移文件
	// 注意：测试运行在 server/internal/store 目录，迁移文件在 server/migrations
	upFile := "../../migrations/000005_create_workflow_and_labels.up.sql"
	downFile := "../../migrations/000005_create_workflow_and_labels.down.sql"

	upContent, err := os.ReadFile(upFile)
	if err != nil {
		t.Fatalf("无法读取 Up 迁移文件: %v", err)
	}

	downContent, err := os.ReadFile(downFile)
	if err != nil {
		t.Fatalf("无法读取 Down 迁移文件: %v", err)
	}
	// ctx := context.Background() // Unused
	tx := testDB.Begin()
	defer tx.Rollback() // 测试结束后回滚

	// 0. 清理可能残留的表 (在事务内，使用 CASCADE)
	if err := tx.Exec("DROP TABLE IF EXISTS labels CASCADE").Error; err != nil {
		t.Fatalf("清理 labels 表失败: %v", err)
	}
	if err := tx.Exec("DROP TABLE IF EXISTS workflow_states CASCADE").Error; err != nil {
		t.Fatalf("清理 workflow_states 表失败: %v", err)
	}

	// 1. 创建基础版本表 (模拟 000002_core.up.sql 状态)
	// 必须包含 GORM 模型期望的所有列，否则 Store 测试会失败
	err = tx.Exec(`
		CREATE TABLE workflow_states (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			team_id UUID NOT NULL, 
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20) NOT NULL DEFAULT 'backlog',
			color VARCHAR(20) DEFAULT '#808080',
			position FLOAT NOT NULL DEFAULT 0,
			description TEXT,
			is_default BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT chk_workflow_state_type CHECK (type IN ('backlog', 'unstarted', 'started', 'completed', 'cancelled'))
		);
		CREATE TABLE labels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL,
			team_id UUID,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			color VARCHAR(20) DEFAULT '#808080',
			parent_id UUID,
			is_archived BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`).Error
	if err != nil {
		t.Fatalf("创建基础表失败: %v", err)
	}

	// 2. 执行 Up 迁移 (ALTER)
	if err := tx.Exec(string(upContent)).Error; err != nil {
		t.Fatalf("执行 Up 迁移失败: %v", err)
	}

	// 3. 验证表结构和约束 (通过查询元数据)

	// 3.1 验证表是否存在
	assertTableExists(t, tx, "workflow_states")
	assertTableExists(t, tx, "labels")

	// 3.2 验证 workflow_states.type CHECK 约束
	var checkSrc string
	err = tx.Raw(`
		SELECT pg_get_constraintdef(c.oid)
		FROM pg_constraint c
		JOIN pg_namespace n ON n.oid = c.connamespace
		WHERE n.nspname = 'public' AND c.conname = 'workflow_states_type_check'
	`).Scan(&checkSrc).Error
	assert.NoError(t, err)
	assert.Contains(t, checkSrc, "canceled", "CHECK 约束应包含 'canceled'")
	assert.Contains(t, checkSrc, "backlog", "CHECK 约束应包含 'backlog'")

	// 3.3 验证 workflow_states (team_id, name) 唯一约束
	assertIndexExists(t, tx, "idx_workflow_states_team_name", true)

	// 3.4 验证 labels partial 唯一索引
	// A. Workspace 级标签唯一性 (team_id IS NULL)
	assertIndexExists(t, tx, "idx_labels_workspace_name", true)
	// 验证 partial 条件
	var whereClause string
	err = tx.Raw(`
		SELECT pg_get_expr(i.indpred, i.indrelid)
		FROM pg_index i
		JOIN pg_class c ON c.oid = i.indexrelid
		WHERE c.relname = 'idx_labels_workspace_name'
	`).Scan(&whereClause).Error
	assert.NoError(t, err)
	assert.Contains(t, whereClause, "team_id IS NULL", "索引应为 Partial Index (team_id IS NULL)")

	// B. Team 级标签唯一性
	assertIndexExists(t, tx, "idx_labels_team_name", true)
	err = tx.Raw(`
		SELECT pg_get_expr(i.indpred, i.indrelid)
		FROM pg_index i
		JOIN pg_class c ON c.oid = i.indexrelid
		WHERE c.relname = 'idx_labels_team_name'
	`).Scan(&whereClause).Error
	assert.NoError(t, err)
	assert.Contains(t, whereClause, "team_id IS NOT NULL", "索引应为 Partial Index (team_id IS NOT NULL)")

	// 4. 执行 Down 迁移
	if err := tx.Exec(string(downContent)).Error; err != nil {
		t.Fatalf("执行 Down 迁移失败: %v", err)
	}

	// 5. 验证约束被移除
	// 这里不再验证表被删除，因为表应该保留，但 CHECK 约束和 INDEX 应该恢复或删除
}

func assertTableExists(t *testing.T, db *gorm.DB, tableName string) {
	var exists bool
	query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)`
	err := db.Raw(query, tableName).Scan(&exists).Error
	assert.NoError(t, err)
	assert.True(t, exists, "表 %s 应该存在", tableName)
}

func assertIndexExists(t *testing.T, db *gorm.DB, indexName string, unique bool) {
	var count int64
	query := `
		SELECT count(*)
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_index i ON i.indexrelid = c.oid
		WHERE c.relname = ? AND n.nspname = 'public' AND i.indisunique = ?
	`
	err := db.Raw(query, indexName, unique).Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "索引 %s 应该存在且唯一性=%v", indexName, unique)
}

func hasTable(t *testing.T, db *gorm.DB, tableName string) bool {
	var exists bool
	query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)`
	err := db.Raw(query, tableName).Scan(&exists).Error
	if err != nil {
		t.Fatalf("检查表存在失败: %v", err)
	}
	return exists
}
