-- 创建 notification_preferences 表
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL DEFAULT 'in_app',
    type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, channel, type)
);

-- 创建索引
CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX idx_notification_preferences_user_channel ON notification_preferences(user_id, channel);

-- 添加注释
COMMENT ON TABLE notification_preferences IS '用户通知偏好配置表';
COMMENT ON COLUMN notification_preferences.channel IS '通知渠道：in_app, email, slack';
COMMENT ON COLUMN notification_preferences.type IS '通知类型：issue_assigned, issue_mentioned, issue_status_changed, issue_commented, issue_priority_changed';
COMMENT ON COLUMN notification_preferences.enabled IS '是否启用该类型通知';
