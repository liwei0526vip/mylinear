package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Task 1.1: 编写 db/migrations 测试
// 该测试验证 000006_add_issue_position_and_subscriptions.up.sql 及其约束
func TestMigration_C05_Schema(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用，跳过集成测试")
	}

	// 1. 读取迁移文件
	// 注意：测试运行在 server/internal/store 目录，迁移文件在 server/migrations
	upFile := "../../migrations/000006_add_issue_position_and_subscriptions.up.sql"
	downFile := "../../migrations/000006_add_issue_position_and_subscriptions.down.sql"

	upContent, err := os.ReadFile(upFile)
	if err != nil {
		t.Fatalf("无法读取 Up 迁移文件: %v", err)
	}

	downContent, err := os.ReadFile(downFile)
	if err != nil {
		t.Fatalf("无法读取 Down 迁移文件: %v", err)
	}

	tx := testDB.Begin()
	defer tx.Rollback() // 测试结束后回滚

	// 0. 清理可能残留的表 (在事务内，使用 CASCADE)
	if err := tx.Exec("DROP TABLE IF EXISTS issue_subscriptions CASCADE").Error; err != nil {
		t.Fatalf("清理 issue_subscriptions 表失败: %v", err)
	}
	if err := tx.Exec("DROP TABLE IF EXISTS issues CASCADE").Error; err != nil {
		t.Fatalf("清理 issues 表失败: %v", err)
	}
	if err := tx.Exec("DROP TABLE IF EXISTS workflow_states CASCADE").Error; err != nil {
		t.Fatalf("清理 workflow_states 表失败: %v", err)
	}
	if err := tx.Exec("DROP TABLE IF EXISTS team_members CASCADE").Error; err != nil {
		t.Fatalf("清理 team_members 表失败: %v", err)
	}
	if err := tx.Exec("DROP TABLE IF EXISTS teams CASCADE").Error; err != nil {
		t.Fatalf("清理 teams 表失败: %v", err)
	}
	if err := tx.Exec("DROP TABLE IF EXISTS users CASCADE").Error; err != nil {
		t.Fatalf("清理 users 表失败: %v", err)
	}
	if err := tx.Exec("DROP TABLE IF EXISTS workspaces CASCADE").Error; err != nil {
		t.Fatalf("清理 workspaces 表失败: %v", err)
	}

	// 2. 创建基础表结构 (模拟 000002_core.up.sql 状态)
	// 需要 workspaces, users, teams, workflow_states, issues 表
	err = tx.Exec(`
		CREATE TABLE workspaces (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(50) UNIQUE NOT NULL,
			logo_url TEXT,
			settings JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`).Error
	if err != nil {
		t.Fatalf("创建 workspaces 表失败: %v", err)
	}

	err = tx.Exec(`
		CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			username VARCHAR(50) UNIQUE,
			avatar_url TEXT,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			settings JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`).Error
	if err != nil {
		t.Fatalf("创建 users 表失败: %v", err)
	}

	err = tx.Exec(`
		CREATE TABLE teams (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			parent_id UUID REFERENCES teams(id) ON DELETE SET NULL,
			name VARCHAR(255) NOT NULL,
			key VARCHAR(10) UNIQUE NOT NULL,
			description TEXT,
			icon_url TEXT,
			timezone VARCHAR(50) DEFAULT 'UTC',
			is_private BOOLEAN NOT NULL DEFAULT FALSE,
			cycle_settings JSONB,
			workflow_settings JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`).Error
	if err != nil {
		t.Fatalf("创建 teams 表失败: %v", err)
	}

	err = tx.Exec(`
		CREATE TABLE workflow_states (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20) NOT NULL DEFAULT 'backlog',
			color VARCHAR(20) DEFAULT '#808080',
			position FLOAT NOT NULL DEFAULT 0,
			is_default BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT chk_workflow_state_type CHECK (type IN ('backlog', 'unstarted', 'started', 'completed', 'cancelled'))
		);
	`).Error
	if err != nil {
		t.Fatalf("创建 workflow_states 表失败: %v", err)
	}

	// 创建 issues 表（不含 position 列）
	err = tx.Exec(`
		CREATE TABLE issues (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			number INTEGER NOT NULL,
			title VARCHAR(500) NOT NULL,
			description TEXT,
			status_id UUID NOT NULL REFERENCES workflow_states(id) ON DELETE RESTRICT,
			priority INTEGER NOT NULL DEFAULT 0,
			assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
			project_id UUID,
			milestone_id UUID,
			cycle_id UUID,
			parent_id UUID REFERENCES issues(id) ON DELETE SET NULL,
			estimate INTEGER,
			due_date DATE,
			sla_due_at TIMESTAMPTZ,
			labels UUID[] DEFAULT '{}',
			created_by_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMPTZ,
			cancelled_at TIMESTAMPTZ,
			CONSTRAINT chk_issue_priority CHECK (priority >= 0 AND priority <= 4)
		);
		CREATE UNIQUE INDEX idx_issue_team_number ON issues(team_id, number);
		CREATE INDEX idx_issue_status_id ON issues(status_id);
	`).Error
	if err != nil {
		t.Fatalf("创建 issues 表失败: %v", err)
	}

	// 3. 执行 Up 迁移
	if err := tx.Exec(string(upContent)).Error; err != nil {
		t.Fatalf("执行 Up 迁移失败: %v", err)
	}

	// 4. 验证表结构和约束

	// 4.1 验证 issue_subscriptions 表存在
	assertTableExists(t, tx, "issue_subscriptions")

	// 4.2 验证 issues 表添加了 position 列
	var positionColumnExists bool
	err = tx.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			AND table_name = 'issues'
			AND column_name = 'position'
		)
	`).Scan(&positionColumnExists).Error
	assert.NoError(t, err)
	assert.True(t, positionColumnExists, "issues 表应该有 position 列")

	// 4.3 验证 position 列类型为 FLOAT
	var positionColumnType string
	err = tx.Raw(`
		SELECT data_type FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = 'issues'
		AND column_name = 'position'
	`).Scan(&positionColumnType).Error
	assert.NoError(t, err)
	assert.Equal(t, "double precision", positionColumnType, "position 列类型应为 double precision (FLOAT)")

	// 4.4 验证 issue_subscriptions 表的复合主键
	var pkColumns []string
	err = tx.Raw(`
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = 'issue_subscriptions'::regclass AND i.indisprimary
		ORDER BY array_position(i.indkey, a.attnum)
	`).Scan(&pkColumns).Error
	assert.NoError(t, err)
	assert.Len(t, pkColumns, 2, "issue_subscriptions 表应有复合主键")
	assert.Contains(t, pkColumns, "issue_id", "复合主键应包含 issue_id")
	assert.Contains(t, pkColumns, "user_id", "复合主键应包含 user_id")

	// 4.5 验证 issue_subscriptions 表的外键约束
	var fkCount int64
	err = tx.Raw(`
		SELECT COUNT(*)
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_name = 'issue_subscriptions'
		AND tc.constraint_type = 'FOREIGN KEY'
	`).Scan(&fkCount).Error
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, fkCount, int64(2), "issue_subscriptions 表应有至少 2 个外键约束")

	// 4.6 验证 issue_subscriptions 表有 created_at 列
	var createdAtExists bool
	err = tx.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			AND table_name = 'issue_subscriptions'
			AND column_name = 'created_at'
		)
	`).Scan(&createdAtExists).Error
	assert.NoError(t, err)
	assert.True(t, createdAtExists, "issue_subscriptions 表应有 created_at 列")

	// 4.7 验证 issues 表的 position 索引
	assertIndexExists(t, tx, "idx_issues_position", false)

	// 5. 执行 Down 迁移
	if err := tx.Exec(string(downContent)).Error; err != nil {
		t.Fatalf("执行 Down 迁移失败: %v", err)
	}

	// 6. 验证 Down 迁移效果
	// 6.1 issue_subscriptions 表应该被删除
	assertTableNotExists(t, tx, "issue_subscriptions")

	// 6.2 position 列应该被删除
	err = tx.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			AND table_name = 'issues'
			AND column_name = 'position'
		)
	`).Scan(&positionColumnExists).Error
	assert.NoError(t, err)
	assert.False(t, positionColumnExists, "Down 迁移后 issues 表不应有 position 列")
}

func assertTableNotExists(t *testing.T, db *gorm.DB, tableName string) {
	var exists bool
	query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)`
	err := db.Raw(query, tableName).Scan(&exists).Error
	assert.NoError(t, err)
	assert.False(t, exists, "表 %s 不应该存在", tableName)
}
