package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestLabelStore_CRUD(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	store := NewLabelStore(tx)
	ctx := context.Background()

	// 准备环境
	workspaceID := uuid.New()
	teamID1 := uuid.New()
	teamID2 := uuid.New()

	suffix := uuid.New().String()[:8]
	ws := &model.Workspace{Model: model.Model{ID: workspaceID}, Name: "Label WS", Slug: "lbl-ws-" + suffix}
	t1 := &model.Team{Model: model.Model{ID: teamID1}, WorkspaceID: workspaceID, Name: "Team 1", Key: "T1" + suffix}
	t2 := &model.Team{Model: model.Model{ID: teamID2}, WorkspaceID: workspaceID, Name: "Team 2", Key: "T2" + suffix}

	tx.Create(ws)
	tx.Create(t1)
	tx.Create(t2)

	// 1. Create Global Label
	globalL := &model.Label{
		WorkspaceID: workspaceID,
		Name:        "Global Bug",
		Color:       "#FF0000",
		TeamID:      nil, // Global
	}
	err := store.Create(ctx, globalL)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, globalL.ID)

	// 2. Create Team Label
	teamL1 := &model.Label{
		WorkspaceID: workspaceID,
		Name:        "Team Feature",
		Color:       "#00FF00",
		TeamID:      &teamID1,
	}
	err = store.Create(ctx, teamL1)
	assert.NoError(t, err)

	teamL2 := &model.Label{
		WorkspaceID: workspaceID,
		Name:        "Team Feature", // Same name but different team (or global?)
		Color:       "#0000FF",
		TeamID:      &teamID2,
	}
	err = store.Create(ctx, teamL2) // Should succeed (scope isolation)
	assert.NoError(t, err)

	// 3. ListForWorkspace (Should get ONLY global)
	// Debug: Check if label exists by ID
	debugL, err := store.GetByID(ctx, globalL.ID)
	assert.NoError(t, err)
	if debugL != nil {
		t.Logf("Debug Label: ID=%v, TeamID=%v", debugL.ID, debugL.TeamID)
	} else {
		t.Log("Debug Label: Not Found by ID")
	}

	listWs, err := store.ListForWorkspace(ctx, workspaceID)
	assert.NoError(t, err)
	if assert.Len(t, listWs, 1) {
		assert.Equal(t, "Global Bug", listWs[0].Name)
	}

	// 4. ListForTeam 1 (Global + Team1)
	listT1, err := store.ListForTeam(ctx, workspaceID, teamID1)
	assert.NoError(t, err)
	if assert.Len(t, listT1, 2, "ListForTeam should have 2 items") {
		// Check content
		names := make(map[string]bool)
		for _, l := range listT1 {
			names[l.Name] = true
		}
		assert.True(t, names["Global Bug"], "Global label missing")
		assert.True(t, names["Team Feature"], "Team label missing")
	} else {
		for i, l := range listT1 {
			t.Logf("T1 Expected 2, Got Item %d: %v (TeamID: %v)", i, l.Name, l.TeamID)
		}
	}

	// 5. Update
	globalL.Name = "Global Defect"
	err = store.Update(ctx, globalL)
	assert.NoError(t, err)
	t.Logf("DEBUG: GlobalL after update: ID=%v Name=%v TeamID=%v", globalL.ID, globalL.Name, globalL.TeamID)

	got, err := store.GetByID(ctx, globalL.ID)
	if assert.NoError(t, err) {
		assert.NotNil(t, got)
		if got != nil {
			assert.Equal(t, "Global Defect", got.Name)
			t.Logf("DEBUG: GlobalL fetched: ID=%v Name=%v TeamID=%v", got.ID, got.Name, got.TeamID)
		}
	}

	// 6. Delete
	err = store.Delete(ctx, teamL1.ID)
	assert.NoError(t, err)

	// List all labels in workspace to see what's left
	allWs, _ := store.ListForWorkspace(ctx, workspaceID)
	t.Logf("DEBUG: ListForWorkspace after delete: %d items", len(allWs))
	for i, l := range allWs {
		t.Logf("DEBUG: WS Item %d: %v (TeamID: %v)", i, l.Name, l.TeamID)
	}

	listT1, err = store.ListForTeam(ctx, workspaceID, teamID1)
	assert.NoError(t, err)
	if assert.Len(t, listT1, 1, "ListForTeam should have 1 item (Global)") {
		assert.Equal(t, "Global Defect", listT1[0].Name)
	} else {
		for i, l := range listT1 {
			t.Logf("DEBUG: T1 Expected 1, Got Item %d: %v (TeamID: %v)", i, l.Name, l.TeamID)
		}
	}
}

func TestLabelStore_Constraints(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()
	store := NewLabelStore(tx)
	ctx := context.Background()

	// ... (Setup code omitted) ...
	// (Need to keep setup code from previous edit or context)
	// Re-inserting setup code here for completeness if tool replaces block
	workspaceID := uuid.New()
	teamID := uuid.New()
	suffix := uuid.New().String()[:8]
	ws := &model.Workspace{Model: model.Model{ID: workspaceID}, Name: "Label Const WS", Slug: "lbl-c-ws-" + suffix}
	t1 := &model.Team{Model: model.Model{ID: teamID}, WorkspaceID: workspaceID, Name: "Team C", Key: "TC" + suffix}
	tx.Create(ws)
	tx.Create(t1)

	// Ensure Index Exists (Manually apply if TestMigration deleted it)
	tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_workspace_name_test ON labels(workspace_id, name) WHERE team_id IS NULL")
	tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_team_name_test ON labels(team_id, name) WHERE team_id IS NOT NULL")

	// 1. Duplicate Global Name
	l1 := &model.Label{WorkspaceID: workspaceID, Name: "Unique Global", TeamID: nil}
	store.Create(ctx, l1)

	l2 := &model.Label{WorkspaceID: workspaceID, Name: "Unique Global", TeamID: nil} // Duplicate

	tx.Exec("SAVEPOINT sp1")
	err := store.Create(ctx, l2)
	assert.Error(t, err)
	tx.Exec("ROLLBACK TO SAVEPOINT sp1")

	// Check TX health
	if err := tx.Exec("SELECT 1").Error; err != nil {
		t.Fatalf("DEBUG: Transaction aborted after sp1 rollback: %v", err)
	}

	// 2. Duplicate Team Name
	lt1 := &model.Label{WorkspaceID: workspaceID, Name: "Unique Team", TeamID: &teamID}
	store.Create(ctx, lt1)

	lt2 := &model.Label{WorkspaceID: workspaceID, Name: "Unique Team", TeamID: &teamID} // Duplicate

	tx.Exec("SAVEPOINT sp2")
	err = store.Create(ctx, lt2)
	assert.Error(t, err)
	tx.Exec("ROLLBACK TO SAVEPOINT sp2")

	// Check TX health
	if err := tx.Exec("SELECT 1").Error; err != nil {
		t.Fatalf("DEBUG: Transaction aborted after sp2 rollback: %v", err)
	}

	// 3. Same Name in Different Scope (Should Succeed)
	// Global vs Team
	lt3 := &model.Label{WorkspaceID: workspaceID, Name: "Unique Global", TeamID: &teamID} // Same name as Global l1
	err = store.Create(ctx, lt3)
	assert.NoError(t, err)
}
