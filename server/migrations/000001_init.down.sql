-- 000001_init.down.sql
-- 回滚基础实体表

-- 按依赖顺序删除表（先删除依赖表）
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS workspaces;
