package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type ProjectRepository interface {
    Create(ctx context.Context, p *models.Project) error
    GetAll(ctx context.Context, customerID string) ([]models.Project, error)
    GetByID(ctx context.Context, id string, customerID string) (*models.Project, error)
    GetBySlug(ctx context.Context, slug string, customerID string) (*models.Project, error)
    Update(ctx context.Context, p *models.Project) error
    Delete(ctx context.Context, id string, customerID string) error
}
