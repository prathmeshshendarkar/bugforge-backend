package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type LabelRepository interface {
	CreateLabel(ctx context.Context, l *models.Label) error
	UpdateLabel(ctx context.Context, l *models.Label) error
	DeleteLabel(ctx context.Context, labelID string) error
	GetLabelByID(ctx context.Context, labelID string) (*models.Label, error)
	ListLabelsByProject(ctx context.Context, projectID string) ([]models.Label, error)
}
