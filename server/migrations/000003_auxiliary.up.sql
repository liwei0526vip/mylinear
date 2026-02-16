-- 000003_auxiliary.up.sql
-- 辅助表：issue_closure、issue_status_history、workflow_transitions

-- Issue 层级闭包表（支持任意深度子 Issue 查询）
CREATE TABLE issue_closure (
    ancestor_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    depth INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);

-- 闭包表索引
CREATE INDEX idx_issue_closure_descendant ON issue_closure(descendant_id);
CREATE INDEX idx_issue_closure_depth ON issue_closure(depth);

-- Issue 状态变更历史表
CREATE TABLE issue_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    from_status_id UUID REFERENCES workflow_states(id) ON DELETE SET NULL,
    to_status_id UUID NOT NULL REFERENCES workflow_states(id) ON DELETE RESTRICT,
    changed_by_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 状态历史索引
CREATE INDEX idx_issue_status_history_issue_id ON issue_status_history(issue_id);
CREATE INDEX idx_issue_status_history_changed_at ON issue_status_history(changed_at);
CREATE INDEX idx_issue_status_history_from_status ON issue_status_history(from_status_id);
CREATE INDEX idx_issue_status_history_to_status ON issue_status_history(to_status_id);

-- 工作流转换规则表
CREATE TABLE workflow_transitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    from_state_id UUID REFERENCES workflow_states(id) ON DELETE CASCADE,
    to_state_id UUID NOT NULL REFERENCES workflow_states(id) ON DELETE CASCADE,
    is_allowed BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (team_id, from_state_id, to_state_id)
);

-- 工作流转换索引
CREATE INDEX idx_workflow_transitions_team_id ON workflow_transitions(team_id);
CREATE INDEX idx_workflow_transitions_from_state ON workflow_transitions(from_state_id);
CREATE INDEX idx_workflow_transitions_to_state ON workflow_transitions(to_state_id);
