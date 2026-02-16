-- 000003_auxiliary.down.sql
-- 回滚辅助表

-- 按依赖顺序删除表
DROP TABLE IF EXISTS workflow_transitions;
DROP TABLE IF EXISTS issue_status_history;
DROP TABLE IF EXISTS issue_closure;
