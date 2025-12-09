package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
	"time"
)

type IssueFilter struct {
    Status     *string
    Priority   *string
    AssignedTo *string
    Search     *string
    SortBy     string
    Direction  string
    Limit      int
    Offset     int
}

type IssueRepository interface {
    Create(ctx context.Context, issue *models.Issue) error
	ListAll(ctx context.Context, customerID string) ([]models.IssueWithUser, error)
    GetByID(ctx context.Context, issueID string) (*models.Issue, error)
    ListByProject(ctx context.Context, projectID string, f IssueFilter) ([]models.IssueWithUser, error)
    Update(ctx context.Context, issue *models.Issue) error
    Delete(ctx context.Context, issueID string) error

    CreateComment(ctx context.Context, c *models.IssueComment) error
    GetCommentByID(ctx context.Context, id string) (*models.IssueComment, error)
    UpdateComment(ctx context.Context, c *models.IssueComment) error
    DeleteComment(ctx context.Context, id string) error
    ListCommentsByIssue(ctx context.Context, issueID string) ([]models.IssueComment, error)

    CreateAttachment(ctx context.Context, a *models.IssueAttachment) error
    ListAttachmentsByIssue(ctx context.Context, issueID string) ([]models.IssueAttachment, error)
    DeleteAttachment(ctx context.Context, id string) error

    CreateChecklist(ctx context.Context, cl *models.Checklist) error
    CreateChecklistItem(ctx context.Context, it *models.ChecklistItem) error
    UpdateChecklistItem(ctx context.Context, it *models.ChecklistItem) error
    ListChecklistsByIssue(ctx context.Context, issueID string) ([]models.Checklist, error)
    DeleteChecklist(ctx context.Context, checklistID string) error
    DeleteChecklistItem(ctx context.Context, itemID string) error
    ReorderChecklistItems(ctx context.Context, checklistID string, items []models.ChecklistItem) error

    CreateSubtask(ctx context.Context, s *models.Subtask) error
    GetSubtaskByID(ctx context.Context, id string) (*models.Subtask, error)
    UpdateSubtask(ctx context.Context, s *models.Subtask) error
    ListSubtasksByParent(ctx context.Context, parentIssueID string) ([]models.Subtask, error)
    DeleteSubtask(ctx context.Context, id string) error

    CreateRelation(ctx context.Context, r *models.IssueRelation) error
    ListRelations(ctx context.Context, issueID string) ([]models.IssueRelation, error)
    DeleteRelation(ctx context.Context, id string) error

    UpdateDueDate(ctx context.Context, issueID string, dueDate *time.Time) error
}
