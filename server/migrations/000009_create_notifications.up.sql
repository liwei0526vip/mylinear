-- 创建 notifications 表
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    resource_type VARCHAR(50),
    resource_id UUID,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, read_at) WHERE read_at IS NULL;
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- 添加注释
COMMENT ON TABLE notifications IS '用户通知表';
COMMENT ON COLUMN notifications.type IS '通知类型：issue_assigned, issue_mentioned, issue_status_changed, issue_commented, issue_priority_changed';
COMMENT ON COLUMN notifications.resource_type IS '关联资源类型：issue, comment, project';
COMMENT ON COLUMN notifications.resource_id IS '关联资源ID';
COMMENT ON COLUMN notifications.read_at IS '已读时间，null表示未读';
