package service

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	service "bugforge-backend/internal/service/interfaces"
	"context"
	"errors"

	"github.com/google/uuid"
)

type LabelServiceImpl struct {
	labelRepo   repo.LabelRepository
	projectRepo repo.ProjectRepository
}

func NewLabelService(labelRepo repo.LabelRepository, projectRepo repo.ProjectRepository) service.LabelService {
	return &LabelServiceImpl{
		labelRepo:   labelRepo,
		projectRepo: projectRepo,
	}
}

func (s *LabelServiceImpl) CreateLabel(ctx context.Context, customerID, projectID, name, color, userID string) (*models.Label, error) {

	if err := s.ensureProjectBelongsToCustomer(ctx, projectID, customerID); err != nil {
		return nil, err
	}

	if name == "" {
		return nil, errors.New("name is required")
	}

	if color == "" {
		color = "#999999"
	}

	l := &models.Label{
		ID:         uuid.NewString(),
		CustomerID: customerID,
		ProjectID:  projectID,
		Name:       name,
		Color:      color,
	}

	if err := s.labelRepo.CreateLabel(ctx, l); err != nil {
		return nil, err
	}

	return l, nil
}

func (s *LabelServiceImpl) UpdateLabel(ctx context.Context, customerID, projectID, labelID, name, color, userID string) (*models.Label, error) {

	l, err := s.labelRepo.GetLabelByID(ctx, labelID)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, errors.New("label not found")
	}

	if l.ProjectID != projectID || l.CustomerID != customerID {
		return nil, errors.New("label not found or invalid tenant")
	}

	if name != "" {
		l.Name = name
	}
	if color != "" {
		l.Color = color
	}

	if err := s.labelRepo.UpdateLabel(ctx, l); err != nil {
		return nil, err
	}

	return l, nil
}

func (s *LabelServiceImpl) DeleteLabel(ctx context.Context, customerID, projectID, labelID, userID string) error {

	l, err := s.labelRepo.GetLabelByID(ctx, labelID)
	if err != nil {
		return err
	}
	if l == nil {
		return errors.New("label not found")
	}

	if l.ProjectID != projectID || l.CustomerID != customerID {
		return errors.New("label not found or invalid tenant")
	}

	return s.labelRepo.DeleteLabel(ctx, labelID)
}

func (s *LabelServiceImpl) ListLabelsByProject(ctx context.Context, customerID, projectID string) ([]models.Label, error) {

	if err := s.ensureProjectBelongsToCustomer(ctx, projectID, customerID); err != nil {
		return nil, err
	}

	return s.labelRepo.ListLabelsByProject(ctx, projectID)
}

func (s *LabelServiceImpl) GetLabel(ctx context.Context, customerID, labelID string) (*models.Label, error) {

	l, err := s.labelRepo.GetLabelByID(ctx, labelID)
	if err != nil {
		return nil, err
	}
	if l == nil || l.CustomerID != customerID {
		return nil, errors.New("label not found")
	}

	return l, nil
}

func (s *LabelServiceImpl) ensureProjectBelongsToCustomer(ctx context.Context, projectID, customerID string) error {
	pr, err := s.projectRepo.GetByID(ctx, projectID, customerID)
	if err != nil || pr == nil {
		return errors.New("project not found")
	}
	return nil
}
