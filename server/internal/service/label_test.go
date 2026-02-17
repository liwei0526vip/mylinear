package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelService_LabelCRUD(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	labelStore := store.NewLabelStore(tx)
	svc := NewLabelService(labelStore)

	ctx := context.Background()
	workspaceID := uuid.New()
	teamID := uuid.New()

	// Setup: Create real workspace and team to satisfy FK constraints
	ws := &model.Workspace{
		Model: model.Model{ID: workspaceID},
		Name:  "Test Workspace",
		Slug:  "test-ws-" + uuid.New().String()[:8],
	}
	tx.Create(ws)

	team := &model.Team{
		Model:       model.Model{ID: teamID},
		WorkspaceID: workspaceID,
		Name:        "Test Team",
		Key:         "TEST",
	}
	tx.Create(team)

	// 1. Create Workspace Label
	t.Run("Create Workspace Label", func(t *testing.T) {
		cmd := &CreateLabelParams{
			WorkspaceID: workspaceID,
			Name:        "Global Bug",
			Color:       "#ff0000",
		}
		label, err := svc.CreateLabel(ctx, cmd)
		require.NoError(t, err)
		assert.Equal(t, cmd.Name, label.Name)
		assert.Nil(t, label.TeamID)
	})

	// 2. Create Team Label
	var teamLabelID uuid.UUID
	t.Run("Create Team Label", func(t *testing.T) {
		cmd := &CreateLabelParams{
			WorkspaceID: workspaceID,
			TeamID:      &teamID,
			Name:        "Team Feature",
			Color:       "#00ff00",
		}
		label, err := svc.CreateLabel(ctx, cmd)
		require.NoError(t, err)
		assert.Equal(t, &teamID, label.TeamID)
		teamLabelID = label.ID
	})

	// 3. List Labels
	t.Run("List Labels", func(t *testing.T) {
		// List for workspace only
		wsLabels, err := svc.ListLabels(ctx, workspaceID, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, wsLabels)
		for _, v := range wsLabels {
			assert.Nil(t, v.TeamID)
		}

		// List for team (should include workspace labels)
		teamLabels, err := svc.ListLabels(ctx, workspaceID, &teamID)
		require.NoError(t, err)
		// Should have at least the two we created
		assert.GreaterOrEqual(t, len(teamLabels), 2)
	})

	// 4. Update Label
	t.Run("Update Label", func(t *testing.T) {
		newName := "Updated Team Feature"
		newColor := "#0000ff"
		updated, err := svc.UpdateLabel(ctx, teamLabelID, &UpdateLabelParams{
			Name:  &newName,
			Color: &newColor,
		})
		require.NoError(t, err)
		assert.Equal(t, newName, updated.Name)
		assert.Equal(t, newColor, updated.Color)
	})

	// 5. Delete Label
	t.Run("Delete Label", func(t *testing.T) {
		err := svc.DeleteLabel(ctx, teamLabelID)
		require.NoError(t, err)

		_, err = svc.GetLabel(ctx, teamLabelID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
