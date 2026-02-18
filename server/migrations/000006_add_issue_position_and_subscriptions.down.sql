-- 000006_add_issue_position_and_subscriptions.down.sql
-- C05: Issue 核心 CRUD (回滚)

-- 1. 删除 issue_subscriptions 表
DROP TABLE IF EXISTS issue_subscriptions;

-- 2. 删除 issues 表的 deleted_at 列
DROP INDEX IF EXISTS idx_issues_deleted_at;
ALTER TABLE issues DROP COLUMN IF EXISTS deleted_at;

-- 3. 删除 issues 表的 position 列
DROP INDEX IF EXISTS idx_issues_position;
ALTER TABLE issues DROP COLUMN IF EXISTS position;
