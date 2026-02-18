-- 000006_add_issue_position_and_subscriptions.up.sql
-- C05: Issue 核心 CRUD
-- 1. 为 issues 表添加 position 列（用于拖拽排序）
-- 2. 为 issues 表添加 deleted_at 列（支持软删除）
-- 3. 创建 issue_subscriptions 表（用户订阅 Issue）

-- 1. 为 issues 表添加 position 列
ALTER TABLE issues ADD COLUMN position FLOAT NOT NULL DEFAULT 0;

-- 为 position 列创建索引
CREATE INDEX idx_issues_position ON issues(position);

-- 为现有 Issue 设置 position（按 created_at 排序，递增 1000）
-- 注意：这个更新在事务中执行，确保数据一致性
DO $$
DECLARE
    issue_rec RECORD;
    pos FLOAT := 1000;
BEGIN
    FOR issue_rec IN SELECT id FROM issues ORDER BY created_at ASC LOOP
        UPDATE issues SET position = pos WHERE id = issue_rec.id;
        pos := pos + 1000;
    END LOOP;
END $$;

-- 2. 为 issues 表添加 deleted_at 列（支持软删除）
ALTER TABLE issues ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;

-- 为 deleted_at 列创建索引
CREATE INDEX idx_issues_deleted_at ON issues(deleted_at);

-- 3. 创建 issue_subscriptions 表
CREATE TABLE issue_subscriptions (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (issue_id, user_id)
);

-- 创建索引以支持按用户查询订阅的 Issue
CREATE INDEX idx_issue_subscriptions_user_id ON issue_subscriptions(user_id);
