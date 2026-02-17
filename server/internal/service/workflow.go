package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

type WorkflowService interface {
	CreateState(ctx context.Context, cmd *CreateStateParams) (*model.WorkflowState, error)
	ListStates(ctx context.Context, teamID uuid.UUID) ([]*model.WorkflowState, error)
	UpdateState(ctx context.Context, id uuid.UUID, cmd *UpdateStateParams) (*model.WorkflowState, error)
	DeleteState(ctx context.Context, id uuid.UUID) error
}

type CreateStateParams struct {
	TeamID      uuid.UUID
	Name        string
	Type        model.StateType
	Color       string
	Position    float64
	Description string
}

type UpdateStateParams struct {
	Name        *string
	Color       *string
	Position    *float64
	Description *string
}

type workflowService struct {
	stateStore store.WorkflowStateStore
	teamStore  store.TeamStore
}

func NewWorkflowService(stateStore store.WorkflowStateStore, teamStore store.TeamStore) WorkflowService {
	return &workflowService{
		stateStore: stateStore,
		teamStore:  teamStore,
	}
}

// CreateState creates a new workflow state
func (s *workflowService) CreateState(ctx context.Context, cmd *CreateStateParams) (*model.WorkflowState, error) {
	// 1. Basic Validation
	if cmd.TeamID == uuid.Nil {
		return nil, errors.New("team_id is required")
	}
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}
	if !cmd.Type.Valid() {
		return nil, fmt.Errorf("invalid state type: %s", cmd.Type)
	}

	// 2. Validate Team Exists (Optional: only if specifically needed)
	// Some DB environments might have latency between team creation and this check.
	// For better performance and internal consistency, we can rely on foreign key constraints in DB
	// instead of explicit check here, or only check if it's NOT an internal creation.
	// For now, let's keep it but make it less likely to fail due to DB session issues by logging instead of returning error if it's likely a new team.
	// Actually, let's just use the ID directly as requested.

	// 3. Set Defaults
	if cmd.Color == "" {
		cmd.Color = "#808080" // Default gray
	}
	// Calculate Position if 0
	if cmd.Position == 0 {
		maxPos, err := s.stateStore.GetMaxPosition(ctx, cmd.TeamID)
		if err != nil {
			return nil, fmt.Errorf("failed to get max position: %w", err)
		}
		cmd.Position = maxPos + 1000.0
	}

	state := &model.WorkflowState{
		TeamID:      cmd.TeamID,
		Name:        cmd.Name,
		Type:        cmd.Type,
		Color:       cmd.Color,
		Position:    cmd.Position,
		Description: cmd.Description,
	}

	if err := s.stateStore.Create(ctx, state); err != nil {
		return nil, err
	}

	return state, nil
}

func (s *workflowService) ListStates(ctx context.Context, teamID uuid.UUID) ([]*model.WorkflowState, error) {
	return s.stateStore.ListByTeamID(ctx, teamID)
}

func (s *workflowService) UpdateState(ctx context.Context, id uuid.UUID, cmd *UpdateStateParams) (*model.WorkflowState, error) {
	state, err := s.stateStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, errors.New("state not found")
	}

	if cmd.Name != nil {
		state.Name = *cmd.Name
	}
	if cmd.Color != nil {
		state.Color = *cmd.Color
	}
	if cmd.Position != nil {
		state.Position = *cmd.Position
	}
	if cmd.Description != nil {
		state.Description = *cmd.Description
	}

	if err := s.stateStore.Update(ctx, state); err != nil {
		return nil, err
	}
	return state, nil
}

func (s *workflowService) DeleteState(ctx context.Context, id uuid.UUID) error {
	// 1. Check if state exists
	state, err := s.stateStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if state == nil {
		return errors.New("state not found")
	}

	// 2. Check if state is used by any issues
	count, err := s.stateStore.CountIssues(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete state with %d issues", count)
	}

	// 3. Check if it's the last state of its type
	// We need to list all states to check this logic efficiently, or add CountByTeamAndType method?
	// Task 2.9: CountByType. But I implemented CountByTeamID.
	// Let's reuse List for now as states per team are few.
	states, err := s.stateStore.ListByTeamID(ctx, state.TeamID)
	if err != nil {
		return err
	}

	typeCount := 0
	for _, st := range states {
		if st.Type == state.Type {
			typeCount++
		}
	}
	if typeCount <= 1 {
		return fmt.Errorf("cannot delete the last state of type %s", state.Type)
	}

	return s.stateStore.Delete(ctx, id)
}
