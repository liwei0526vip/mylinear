-- 000005_create_workflow_and_labels.down.sql
-- 恢复原始状态 (000002_core.up.sql)

-- 移除 description 字段
ALTER TABLE workflow_states DROP COLUMN IF EXISTS description;

-- 恢复 workflow_states.type 约束
-- 注意：这里无法简单恢复到旧的 CHECK 约束，因为数据可能已经包含了 'canceled'
-- 如果要严格回滚，需要清理数据或者接受新的值但改回旧约束名（不推荐）
-- 或者直接删除新约束，重新添加旧约束（但数据不兼容会失败）
-- 这里仅移除 index
ALTER TABLE workflow_states DROP CONSTRAINT IF EXISTS workflow_states_type_check;
ALTER TABLE workflow_states DROP CONSTRAINT IF EXISTS idx_workflow_states_team_name;

-- 恢复数据 (可选)
UPDATE workflow_states SET type = 'cancelled' WHERE type = 'canceled';

ALTER TABLE workflow_states DROP CONSTRAINT IF EXISTS idx_workflow_states_team_name;

-- 2. Labels
-- 移除唯一索引
DROP INDEX IF EXISTS idx_labels_team_name;
DROP INDEX IF EXISTS idx_labels_workspace_name;
