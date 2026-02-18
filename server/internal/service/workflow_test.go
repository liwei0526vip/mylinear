package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowService_CreateState(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	stateStore := store.NewWorkflowStateStore(tx)
	teamStore := store.NewTeamStore(tx)
	// IssueStore needed but used via WorkflowStateStore's CountIssues which operates on model.Issue
	// The service constructor takes IssueStore interface but I removed it in refactor?
	// Let's check service definition in workflow.go
	svc := NewWorkflowService(stateStore, teamStore)

	ctx := context.Background()
	workspaceID := uuid.New()
	teamID := uuid.New()

	// Setup
	ws := &model.Workspace{Model: model.Model{ID: workspaceID}, Name: "WS", Slug: "ws-" + uuid.New().String()[:8]}
	tx.Create(ws)
	team := &model.Team{Model: model.Model{ID: teamID}, WorkspaceID: workspaceID, Name: "Team", Key: "T" + uuid.New().String()[:8]}
	tx.Create(team)

	tests := []struct {
		name    string
		cmd     *CreateStateParams
		wantErr bool
		errMsg  string
		check   func(*testing.T, *model.WorkflowState)
	}{
		{
			name: "Success with Defaults",
			cmd: &CreateStateParams{
				TeamID: teamID,
				Name:   "New State",
				Type:   model.StateTypeBacklog,
			},
			wantErr: false,
			check: func(t *testing.T, s *model.WorkflowState) {
				assert.Equal(t, "New State", s.Name)
				assert.Equal(t, "#808080", s.Color) // Default
				assert.Equal(t, 1000.0, s.Position) // Default start
			},
		},
		{
			name: "Success with Custom",
			cmd: &CreateStateParams{
				TeamID:   teamID,
				Name:     "Custom State",
				Type:     model.StateTypeStarted,
				Color:    "#ff0000",
				Position: 500.0,
			},
			wantErr: false,
			check: func(t *testing.T, s *model.WorkflowState) {
				assert.Equal(t, "#ff0000", s.Color)
				assert.Equal(t, 500.0, s.Position)
			},
		},
		{
			name: "Invalid Team",
			cmd: &CreateStateParams{
				TeamID: uuid.New(),
				Name:   "State",
				Type:   model.StateTypeBacklog,
			},
			wantErr: true,
			errMsg:  "violates foreign key constraint", // Was "team not found", now DB constraint
		},
		{
			name: "Invalid Type",
			cmd: &CreateStateParams{
				TeamID: teamID,
				Name:   "State",
				Type:   "invalid",
			},
			wantErr: true,
			errMsg:  "invalid state type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.CreateState(ctx, tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.check != nil {
					tt.check(t, got)
				}
			}
		})
	}
}

func TestWorkflowService_DeleteState(t *testing.T) {
	if testDB == nil {
		t.Skip("数据库连接不可用")
	}

	tx := testDB.Begin()
	defer tx.Rollback()

	stateStore := store.NewWorkflowStateStore(tx)
	teamStore := store.NewTeamStore(tx)
	svc := NewWorkflowService(stateStore, teamStore)

	ctx := context.Background()
	workspaceID := uuid.New()
	teamID := uuid.New()

	// Setup
	ws := &model.Workspace{Model: model.Model{ID: workspaceID}, Name: "WS", Slug: "ws-del-" + uuid.New().String()[:8]}
	tx.Create(ws)
	team := &model.Team{Model: model.Model{ID: teamID}, WorkspaceID: workspaceID, Name: "Team", Key: "TD" + uuid.New().String()[:8]}
	tx.Create(team)

	// Create 2 states of same type
	s1, _ := svc.CreateState(ctx, &CreateStateParams{TeamID: teamID, Name: "S1", Type: model.StateTypeBacklog})
	s2, _ := svc.CreateState(ctx, &CreateStateParams{TeamID: teamID, Name: "S2", Type: model.StateTypeBacklog})

	// Create 1 state of unique type
	s3, _ := svc.CreateState(ctx, &CreateStateParams{TeamID: teamID, Name: "S3", Type: model.StateTypeStarted})

	// Create issue for s1
	issue := &model.Issue{
		TeamID:      teamID,
		Title:       "Issue",
		Number:      1, // Ensure number if not auto-inc? Model has number. But logic handles it via Store usually.
		StatusID:    s1.ID,
		CreatedByID: uuid.New(), // Dummy
	}
	// We need User for CreatedBy
	user := &model.User{WorkspaceID: workspaceID, Email: "u@e.com", Username: "u", Name: "U"}
	tx.Create(user)
	issue.CreatedByID = user.ID
	tx.Create(issue)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Delete State with Issue",
			id:      s1.ID,
			wantErr: true,
			errMsg:  "cannot delete state with 1 issues",
		},
		{
			name:    "Delete Last State of Type",
			id:      s3.ID,
			wantErr: true,
			errMsg:  "cannot delete the last state of type",
		},
		{
			name:    "Success Delete",
			id:      s2.ID,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteState(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
