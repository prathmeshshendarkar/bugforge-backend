package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type CommentRepository interface {
	Create(ctx context.Context, c *models.IssueComment) error
	ListByIssue(ctx context.Context, issueID string) ([]models.IssueComment, error)
}
