-- 000002_core.down.sql
-- 回滚核心业务表

-- 按依赖顺序删除表（先删除依赖表）
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS labels;

-- 删除 issues 表中后添加的外键约束
ALTER TABLE issues DROP CONSTRAINT IF EXISTS fk_issues_cycle_id;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS fk_issues_milestone_id;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS fk_issues_project_id;

DROP TABLE IF EXISTS cycles;
DROP TABLE IF EXISTS milestones;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS issue_relations;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS workflow_states;
