package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type ActivityRepository interface {
	Create(ctx context.Context, a *models.IssueActivity) error
	ListByIssue(ctx context.Context, issueID string) ([]models.IssueActivity, error)
	ListByProject(ctx context.Context, projectID string) ([]models.IssueActivity, error)
}
