# C06 Design: Workflow States & Labels

## 1. 架构决策 (Architecture Decisions)

### 1.1 状态类型 (Workflow State Type)
*   **决策**: 使用 **String** 类型存储状态类别（`backlog`, `unstarted`, `started`, `completed`, `canceled`）。
*   **理由**:
    *   相比 Integer 枚举，String 在数据库和 API 响应中具有自解释性，便于调试。
    *   前端逻辑直接依赖这些类型来渲染图标和分组（例如 "In Progress" 分组），String 映射更直观。
*   **替代方案**: 使用 `SmallInt` (1-5)。虽然节省少量空间，但牺牲了可读性，且在前后端交互中需要额外的映射层。

### 1.2 排序机制 (Ordering)
*   **决策**: 使用 **Float64** (`position` 字段) 实现排序。
*   **理由**:
    *   支持 O(1) 的插入操作（取前后两个元素的平均值）。
    *   避免了 Integer 方案在插入时需要级联更新后续所有记录的问题。
    *   默认间隔设为 1000.0，足以支持大量中间插入而不发生精度冲突。如果发生冲突，可触发一次全量重整（Rebalance）。
*   **替代方案**: Linked List (prev_id/next_id) 查询复杂；Lexorank 算法实现复杂。Float64 是最平衡的选择。

### 1.3 标签作用域 (Label Scope)
*   **决策**: 在同一张 `labels` 表中，通过 `team_id` 字段是否为 `NULL` 来区分 Workspace 级和 Team 级标签。
*   **理由**:
    *   简化表结构，避免创建 `workspace_labels` 和 `team_labels` 两张表。
    *   查询时只需一个 `WHERE team_id = ? OR team_id IS NULL` 即可获取所有可用标签。
*   **替代方案**: 多态关联表。对于仅有两种作用域的场景过于复杂。

---

## 2. 数据库设计 (Database Design)

### 2.1 `workflow_states` 表
```sql
CREATE TABLE workflow_states (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id     UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name        VARCHAR(50) NOT NULL,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('backlog', 'unstarted', 'started', 'completed', 'canceled')),
    color       VARCHAR(7) NOT NULL DEFAULT '#E5E7EB', -- Hex color
    position    DOUBLE PRECISION NOT NULL DEFAULT 65535.0,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- 约束：同个团队内状态名唯一
    CONSTRAINT idx_workflow_states_team_name UNIQUE (team_id, name)
);

-- 索引：用于按顺序查询
CREATE INDEX idx_workflow_states_team_position ON workflow_states(team_id, type, position);
```

### 2.2 `labels` 表
```sql
CREATE TABLE labels (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id     UUID REFERENCES teams(id) ON DELETE CASCADE, -- NULL 表示 Workspace 级
    name        VARCHAR(50) NOT NULL,
    color       VARCHAR(7) NOT NULL DEFAULT '#E5E7EB',
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 约束：同个作用域内标签名唯一（PostgreSQL 对 NULL 的处理需注意，这里使用部分索引或应用层校验）
    -- 为简单起见，使用应用层校验唯一性，复合索引用于查询优化
);

CREATE INDEX idx_labels_team ON labels(team_id);
-- 针对 workspace 级标签的唯一索引
CREATE UNIQUE INDEX idx_labels_workspace_name ON labels(name) WHERE team_id IS NULL;
-- 针对 team 级标签的唯一索引
CREATE UNIQUE INDEX idx_labels_team_name ON labels(team_id, name) WHERE team_id IS NOT NULL;
```

---

## 3. API 设计 (API Design)

### 3.1 Workflow States

#### `GET /api/v1/teams/:team_id/workflow-states`
*   **功能**: 获取团队工作流状态列表。
*   **Response**:
    ```json
    [
      {
        "id": "uuid",
        "name": "Backlog",
        "type": "backlog",
        "color": "#bec2c8",
        "position": 1000.0
      },
      // ... 按 type 分组排序，组内按 position 排序
    ]
    ```

#### `POST /api/v1/teams/:team_id/workflow-states`
*   **Body**: `{ "name": "Review", "type": "started", "color": "#f2c94c", "position": 2500.0 }`
*   **Logic**: 若未提供 position，计算为该 type 组 max_position + 1000。

#### `PUT /api/v1/workflow-states/:id`
*   **Body**: `{ "name": "Code Review", "position": 2600.0 }` (Partial Update)
*   **Logic**: 若 type 变更，需校验是否违反 "每种类型至少一个状态" 规则。

#### `DELETE /api/v1/workflow-states/:id`
*   **Query**: `transfer_to_id` (Optional, 暂不实现迁移逻辑，仅做删除校验)
*   **Logic**: 
    1. 校验是否为该 Type 最后一个状态 -> 400.
    2. 校验是否有关联 Issue -> 400.

### 3.2 Labels

#### `GET /api/v1/teams/:team_id/labels`
*   **功能**: 获取团队可用标签（包含 Workspace 级）。
*   **Logic**: `SELECT * FROM labels WHERE team_id = ? OR team_id IS NULL`.

#### `POST /api/v1/teams/:team_id/labels`
*   **功能**: 创建团队标签。

#### `POST /api/v1/labels`
*   **功能**: 创建 Workspace 标签（需 Admin 权限）。

---

## 4. 前端设计 (Frontend Design)

### 4.1 UI 组件
1.  **WorkflowIcon**: 根据 `type` 渲染不同 SVG。
    *   `backlog`: 虚线圆
    *   `unstarted`: 空心圆
    *   `started`: 半圆
    *   `completed`: 勾选圆 (紫色/蓝色)
    *   `canceled`: 叉号圆 (灰色)
2.  **StateBadge**: 组合 Icon + Text，支持自定义颜色。
3.  **LabelBadge**: 圆角矩形，细边框，背景色为 color 的 10% 透明度。

### 4.2 State Management (Zustand)
*   **useWorkflowStore**:
    *   `states`: Map<TeamId, State[]>
    *   `fetchStates(teamId)`
    *   `addState(state)`
    *   `updateState(id, patch)`
*   **useLabelStore**:
    *   `labels`: Map<TeamId | 'workspace', Label[]>
    *   `fetchLabels(teamId)`

---

## 5. 风险与缓解 (Risks)

*   **风险**: 浮点数精度问题可能导致 position 冲突。
    *   **缓解**: 在应用层检测 `position` 差值是否小于 `0.0001`，如果是，触发后台 Rebalance 任务（重置所有 position 为整千数）。
*   **风险**: 下一阶段 Issue 数据量大时，关联状态的查询性能。
    *   **缓解**: `workflow_states`表通常很小（每团队 ~10 条），常驻 Redis 缓存。

