package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

type CreateLabelParams struct {
	WorkspaceID uuid.UUID
	TeamID      *uuid.UUID // Optional, nil for workspace-level label
	Name        string
	Color       string
}

type UpdateLabelParams struct {
	Name  *string
	Color *string
}

type LabelService interface {
	CreateLabel(ctx context.Context, cmd *CreateLabelParams) (*model.Label, error)
	ListLabels(ctx context.Context, workspaceID uuid.UUID, teamID *uuid.UUID) ([]*model.Label, error)
	GetLabel(ctx context.Context, id uuid.UUID) (*model.Label, error)
	UpdateLabel(ctx context.Context, id uuid.UUID, cmd *UpdateLabelParams) (*model.Label, error)
	DeleteLabel(ctx context.Context, id uuid.UUID) error
}

type labelService struct {
	labelStore store.LabelStore
}

func NewLabelService(labelStore store.LabelStore) LabelService {
	return &labelService{labelStore: labelStore}
}

func (s *labelService) CreateLabel(ctx context.Context, cmd *CreateLabelParams) (*model.Label, error) {
	if cmd.WorkspaceID == uuid.Nil {
		return nil, errors.New("workspace_id is required")
	}
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}

	// Default color if empty?
	if cmd.Color == "" {
		cmd.Color = "#bec2c8" // Default gray
	}

	label := &model.Label{
		WorkspaceID: cmd.WorkspaceID,
		TeamID:      cmd.TeamID,
		Name:        cmd.Name,
		Color:       cmd.Color,
	}

	if err := s.labelStore.Create(ctx, label); err != nil {
		return nil, err
	}

	return label, nil
}

func (s *labelService) ListLabels(ctx context.Context, workspaceID uuid.UUID, teamID *uuid.UUID) ([]*model.Label, error) {
	if teamID != nil && *teamID != uuid.Nil {
		return s.labelStore.ListForTeam(ctx, workspaceID, *teamID)
	}
	return s.labelStore.ListForWorkspace(ctx, workspaceID)
}

func (s *labelService) GetLabel(ctx context.Context, id uuid.UUID) (*model.Label, error) {
	label, err := s.labelStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if label == nil {
		return nil, errors.New("label not found")
	}
	return label, nil
}

func (s *labelService) UpdateLabel(ctx context.Context, id uuid.UUID, cmd *UpdateLabelParams) (*model.Label, error) {
	label, err := s.labelStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if label == nil {
		return nil, errors.New("label not found")
	}

	if cmd.Name != nil {
		label.Name = *cmd.Name
	}
	if cmd.Color != nil {
		label.Color = *cmd.Color
	}

	if err := s.labelStore.Update(ctx, label); err != nil {
		return nil, err
	}

	return label, nil
}

func (s *labelService) DeleteLabel(ctx context.Context, id uuid.UUID) error {
	// Check reference? e.g. used by issues?
	// Currently Issue model doesn't link to Label via FK in a simple way (MM relationship usually).
	// If MM (issue_labels table), we should check or cascade.
	// Spec doesn't detail Issue-Label relationship yet?
	// Assuming simple soft delete or hard delete. Store implementation is hard delete.
	// For now, allow delete.

	// Check if label exists first
	label, err := s.labelStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if label == nil {
		return errors.New("label not found")
	}

	return s.labelStore.Delete(ctx, id)
}
