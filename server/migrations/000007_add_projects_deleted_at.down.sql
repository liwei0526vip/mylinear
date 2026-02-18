-- 000007_add_projects_deleted_at.down.sql
DROP INDEX IF EXISTS idx_projects_deleted_at;
ALTER TABLE projects DROP COLUMN IF EXISTS deleted_at;
