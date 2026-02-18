-- 000007_add_projects_deleted_at.up.sql
ALTER TABLE projects ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
CREATE INDEX idx_projects_deleted_at ON projects(deleted_at);
