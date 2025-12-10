package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type CommentRepository interface {
	Create(ctx context.Context, c *models.IssueComment) error
	ListCommentsByIssue(ctx context.Context, issueID string) ([]models.IssueComment, error)
	Update(ctx context.Context, c *models.IssueComment) error
}
