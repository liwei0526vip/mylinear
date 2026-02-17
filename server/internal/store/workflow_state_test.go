package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestWorkflowStateStore_CRUD(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewWorkflowStateStore(tx)
	ctx := context.Background()

	// 准备 Team
	teamID := uuid.New()
	workspaceID := uuid.New()

	workspace := &model.Workspace{
		Model: model.Model{ID: workspaceID},
		Name:  "Test Workspace WS",
		Slug:  "test-ws-" + uuid.New().String(),
	}
	// 需要创建 Workspace 吗？Workspace 也有外键？
	// 000001 created workspaces. 000002 created teams references workspace.
	// So we need a workspace.
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建 Workspace 失败: %v", err)
	}

	team := &model.Team{
		Model:       model.Model{ID: teamID},
		WorkspaceID: workspaceID,
		Key:         "T" + uuid.New().String()[:8],
		Name:        "Test Team",
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建 Team 失败: %v", err)
	}

	// 1. Create
	s1 := &model.WorkflowState{
		TeamID:   teamID,
		Name:     "Backlog",
		Type:     model.StateTypeBacklog,
		Position: 100,
		Color:    "#000000",
	}
	err := store.Create(ctx, s1)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, s1.ID)

	s2 := &model.WorkflowState{
		TeamID:   teamID,
		Name:     "In Progress",
		Type:     model.StateTypeStarted,
		Position: 200,
		Color:    "#0000FF",
	}
	err = store.Create(ctx, s2)
	assert.NoError(t, err)

	// 2. List (Verify Order)
	s3 := &model.WorkflowState{
		TeamID:   teamID,
		Name:     "Todo",
		Type:     model.StateTypeUnstarted,
		Position: 150, // Between 100 and 200
		Color:    "#FFFFFF",
	}
	err = store.Create(ctx, s3)
	assert.NoError(t, err)

	list, err := store.ListByTeamID(ctx, teamID)
	assert.NoError(t, err)
	assert.Len(t, list, 3)
	assert.Equal(t, "Backlog", list[0].Name)
	assert.Equal(t, "Todo", list[1].Name) // 150 < 200
	assert.Equal(t, "In Progress", list[2].Name)

	// 3. GetByID
	got, err := store.GetByID(ctx, s1.ID)
	assert.NoError(t, err)
	assert.Equal(t, s1.Name, got.Name)

	// Get Not Found
	_, err = store.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// 4. Update
	s1.Name = "New Backlog"
	err = store.Update(ctx, s1)
	assert.NoError(t, err)

	got, _ = store.GetByID(ctx, s1.ID)
	assert.Equal(t, "New Backlog", got.Name)

	// 5. GetMaxPosition
	maxPos, err := store.GetMaxPosition(ctx, teamID)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), maxPos)

	// 6. Delete
	err = store.Delete(ctx, s1.ID)
	assert.NoError(t, err)

	list, err = store.ListByTeamID(ctx, teamID)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestWorkflowStateStore_Constraints(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewWorkflowStateStore(tx)
	ctx := context.Background()

	// 准备 Team
	teamID := uuid.New()
	workspaceID := uuid.New()

	workspace := &model.Workspace{
		Model: model.Model{ID: workspaceID},
		Name:  "Constraint WS",
		Slug:  "const-ws-" + uuid.New().String(),
	}
	team := &model.Team{
		Model:       model.Model{ID: teamID},
		WorkspaceID: workspaceID,
		Key:         "C" + uuid.New().String()[:8],
		Name:        "Constraint Team",
	}
	if err := tx.Create(workspace).Error; err != nil {
		t.Fatalf("创建 Workspace 失败: %v", err)
	}
	if err := tx.Create(team).Error; err != nil {
		t.Fatalf("创建 Team 失败: %v", err)
	}

	// 手动添加唯一约束（因为 TestMigration 可能已经移除了它，且 TestMain 未包含）
	// 注意：这是为了测试 Constraint 行为，模拟真实环境
	tx.Exec("ALTER TABLE workflow_states ADD CONSTRAINT idx_workflow_states_team_name_test UNIQUE (team_id, name)")

	// 1. Unique Name per Team
	s1 := &model.WorkflowState{
		TeamID: teamID,
		Name:   "UniqueName",
		Type:   model.StateTypeBacklog,
	}
	store.Create(ctx, s1)

	s2 := &model.WorkflowState{
		TeamID: teamID,
		Name:   "UniqueName", // Same Name
		Type:   model.StateTypeBacklog,
	}

	// 使用 SavePoint 隔离错误
	tx.SavePoint("sp_unique")
	err := store.Create(ctx, s2)
	assert.Error(t, err)
	tx.RollbackTo("sp_unique")

	// Manually add check constraint for test (AutoMigrate doesn't add it)
	tx.Exec("ALTER TABLE workflow_states ADD CONSTRAINT chk_workflow_state_type_test CHECK (type IN ('backlog', 'unstarted', 'started', 'completed', 'canceled'))")

	// 2. Invalid Type
	// 注意：Model 中 Type 是 StateType 类型，已经是强类型。
	// 但如果通过 reflect 或者底层 DB 约束检查...
	// 我们这里直接通过 struct 赋值，如果 insert 成功则说明 DB 约束生效？
	// 不，enum type check 是在 DB 层。
	// Go 的 enum 只是 string。
	s3 := &model.WorkflowState{
		TeamID: teamID,
		Name:   "InvalidType",
		Type:   model.StateType("invalid_type_value"),
	}
	tx.Exec("SAVEPOINT sp_type")
	err = store.Create(ctx, s3)
	assert.Error(t, err)
	// 期望 DB 返回 check violation
	tx.Exec("ROLLBACK TO SAVEPOINT sp_type")
}
