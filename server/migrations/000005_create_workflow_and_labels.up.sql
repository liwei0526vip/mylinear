-- 000005_create_workflow_and_labels.up.sql
-- 补充和调整 workflow_states 和 labels 表结构 (基于 000002_core.up.sql)

-- 1. Workflow States
-- 修改 type check 约束，支持 'canceled' (单l)，并添加唯一约束

-- 移除旧约束 (假设名为 chk_workflow_state_type，来自 000002)
ALTER TABLE workflow_states DROP CONSTRAINT IF EXISTS chk_workflow_state_type;

-- 更新可能存在的旧数据 (防卫性)
UPDATE workflow_states SET type = 'canceled' WHERE type = 'cancelled';

-- 添加新约束
ALTER TABLE workflow_states ADD CONSTRAINT workflow_states_type_check 
    CHECK (type IN ('backlog', 'unstarted', 'started', 'completed', 'canceled'));

-- 添加 (team_id, name) 唯一约束
-- 先确保无重复数据 (可选，这里假设数据干净或为空)
ALTER TABLE workflow_states ADD CONSTRAINT idx_workflow_states_team_name UNIQUE (team_id, name);

-- 添加 description 字段 (Design 2.1)
ALTER TABLE workflow_states ADD COLUMN IF NOT EXISTS description TEXT;


-- 2. Labels
-- 000002 已创建 labels 表，含 workspace_id, team_id
-- 需要添加唯一性约束

-- 针对 team 级标签的唯一索引
CREATE UNIQUE INDEX idx_labels_team_name ON labels(team_id, name) WHERE team_id IS NOT NULL;

-- 针对 workspace 级标签的唯一索引 (team_id IS NULL)
-- 注意：000002 中 labels 必须属于 workspace_id。
-- 所以是：同一个 workspace 下，name 唯一（当 team_id 为空时）
CREATE UNIQUE INDEX idx_labels_workspace_name ON labels(workspace_id, name) WHERE team_id IS NULL;
