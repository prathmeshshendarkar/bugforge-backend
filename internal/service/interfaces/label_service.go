package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type LabelService interface {
	CreateLabel(ctx context.Context, customerID, projectID, name, color, userID string) (*models.Label, error)
	UpdateLabel(ctx context.Context, customerID, projectID, labelID, name, color, userID string) (*models.Label, error)
	DeleteLabel(ctx context.Context, customerID, projectID, labelID, userID string) error
	ListLabelsByProject(ctx context.Context, customerID, projectID string) ([]models.Label, error)
	GetLabel(ctx context.Context, customerID, labelID string) (*models.Label, error)
}
