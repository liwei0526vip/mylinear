-- 000002_core.up.sql
-- 核心业务表：issues、issue_relations、workflow_states、projects、milestones、cycles、labels、comments、attachments、documents、notifications

-- 工作流状态表
CREATE TABLE workflow_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'backlog',
    color VARCHAR(20) DEFAULT '#808080',
    position FLOAT NOT NULL DEFAULT 0,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_workflow_state_type CHECK (type IN ('backlog', 'unstarted', 'started', 'completed', 'cancelled'))
);

CREATE INDEX idx_workflow_states_team_id ON workflow_states(team_id);
CREATE INDEX idx_workflow_states_position ON workflow_states(position);

-- Issue 工单表
CREATE TABLE issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status_id UUID NOT NULL REFERENCES workflow_states(id) ON DELETE RESTRICT,
    priority INTEGER NOT NULL DEFAULT 0,
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    project_id UUID,  -- 外键在 projects 表创建后添加
    milestone_id UUID, -- 外键在 milestones 表创建后添加
    cycle_id UUID,     -- 外键在 cycles 表创建后添加
    parent_id UUID REFERENCES issues(id) ON DELETE SET NULL,
    estimate INTEGER,
    due_date DATE,
    sla_due_at TIMESTAMPTZ,
    labels UUID[] DEFAULT '{}',
    created_by_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    CONSTRAINT chk_issue_priority CHECK (priority >= 0 AND priority <= 4)
);

-- Issue 核心索引
CREATE UNIQUE INDEX idx_issue_team_number ON issues(team_id, number);
CREATE INDEX idx_issue_assignee_id ON issues(assignee_id);
CREATE INDEX idx_issue_status_id ON issues(status_id);
CREATE INDEX idx_issue_parent_id ON issues(parent_id);
CREATE INDEX idx_issue_created_by_id ON issues(created_by_id);
CREATE INDEX idx_issue_labels ON issues USING GIN (labels);
CREATE INDEX idx_issue_due_date ON issues(due_date);
CREATE INDEX idx_issue_created_at ON issues(created_at);

-- Issue 关系表
CREATE TABLE issue_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    related_issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL DEFAULT 'related',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (issue_id, related_issue_id, type),
    CONSTRAINT chk_issue_relation_type CHECK (type IN ('blocked_by', 'blocking', 'related', 'duplicate'))
);

CREATE INDEX idx_issue_relations_issue_id ON issue_relations(issue_id);
CREATE INDEX idx_issue_relations_related_issue_id ON issue_relations(related_issue_id);

-- 项目表
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'planned',
    priority INTEGER NOT NULL DEFAULT 0,
    lead_id UUID REFERENCES users(id) ON DELETE SET NULL,
    start_date DATE,
    target_date DATE,
    teams UUID[] DEFAULT '{}',
    labels UUID[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT chk_project_status CHECK (status IN ('planned', 'in_progress', 'paused', 'completed', 'cancelled')),
    CONSTRAINT chk_project_priority CHECK (priority >= 0 AND priority <= 4)
);

CREATE INDEX idx_projects_workspace_id ON projects(workspace_id);
CREATE INDEX idx_projects_lead_id ON projects(lead_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_teams ON projects USING GIN (teams);
CREATE INDEX idx_projects_labels ON projects USING GIN (labels);

-- 添加 issues 表到 projects 的外键
ALTER TABLE issues ADD CONSTRAINT fk_issues_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX idx_issue_project_id ON issues(project_id);

-- 里程碑表
CREATE TABLE milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    target_date DATE,
    position FLOAT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_milestones_project_id ON milestones(project_id);
CREATE INDEX idx_milestones_position ON milestones(position);

-- 添加 issues 表到 milestones 的外键
ALTER TABLE issues ADD CONSTRAINT fk_issues_milestone_id FOREIGN KEY (milestone_id) REFERENCES milestones(id) ON DELETE SET NULL;
CREATE INDEX idx_issue_milestone_id ON issues(milestone_id);

-- 迭代表
CREATE TABLE cycles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    name VARCHAR(255),
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    cooldown_end_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'upcoming',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (team_id, number),
    CONSTRAINT chk_cycle_status CHECK (status IN ('upcoming', 'active', 'completed'))
);

CREATE INDEX idx_cycles_team_id ON cycles(team_id);
CREATE INDEX idx_cycles_status ON cycles(status);
CREATE INDEX idx_cycles_dates ON cycles(start_date, end_date);

-- 添加 issues 表到 cycles 的外键
ALTER TABLE issues ADD CONSTRAINT fk_issues_cycle_id FOREIGN KEY (cycle_id) REFERENCES cycles(id) ON DELETE SET NULL;
CREATE INDEX idx_issue_cycle_id ON issues(cycle_id);

-- 标签表
CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(20) DEFAULT '#808080',
    parent_id UUID REFERENCES labels(id) ON DELETE SET NULL,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_labels_workspace_id ON labels(workspace_id);
CREATE INDEX idx_labels_team_id ON labels(team_id);
CREATE INDEX idx_labels_parent_id ON labels(parent_id);
CREATE INDEX idx_labels_is_archived ON labels(is_archived);

-- 评论表
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMPTZ
);

CREATE INDEX idx_comments_issue_id ON comments(issue_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);

-- 附件表
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attachments_issue_id ON attachments(issue_id);
CREATE INDEX idx_attachments_user_id ON attachments(user_id);

-- 文档表
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    issue_id UUID REFERENCES issues(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT,
    icon VARCHAR(50),
    created_by_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_workspace_id ON documents(workspace_id);
CREATE INDEX idx_documents_project_id ON documents(project_id);
CREATE INDEX idx_documents_issue_id ON documents(issue_id);
CREATE INDEX idx_documents_created_by_id ON documents(created_by_id);

-- 通知表
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    resource_type VARCHAR(50),
    resource_id UUID,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_notification_type CHECK (type IN ('issue_assigned', 'issue_mentioned', 'issue_commented', 'issue_status_changed', 'project_updated', 'cycle_started', 'cycle_ended'))
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_read_at ON notifications(read_at);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_resource ON notifications(resource_type, resource_id);
