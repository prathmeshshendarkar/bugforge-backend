package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type ProjectService interface {
    CreateProject(ctx context.Context, customerID, name, slug string) (*models.Project, error)
    GetProjects(ctx context.Context, customerID string) ([]models.Project, error)
    GetProjectByID(ctx context.Context, id, customerID string) (*models.Project, error)
    UpdateProject(ctx context.Context, id, customerID, name, slug string) (*models.Project, error)
    DeleteProject(ctx context.Context, id, customerID string) error
}
