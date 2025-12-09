package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
	"net/url"
	"time"
)

type IssueService interface {
	// ─────────── Core Issue ───────────
	CreateIssue(ctx context.Context, customerID, projectID, title, description, priority string, assignedTo *string, actorUserID string) (*models.Issue, error)
	ListAllIssues(ctx context.Context, customerID string) ([]models.IssueWithUser, error)
	GetIssue(ctx context.Context, customerID, issueID string) (*models.Issue, error)
	ListIssuesByProject(ctx context.Context, projectID, customerID string, q url.Values) ([]models.IssueWithUser, error)
	UpdateIssue(ctx context.Context, customerID, issueID string, title, description, status, priority string, assignedTo *string, actorUserID string) (*models.Issue, error)
	DeleteIssue(ctx context.Context, customerID, issueID, actorUserID string) error

	// ─────────── Due Date ───────────
	UpdateDueDate(ctx context.Context, customerID, issueID string, dueDate *time.Time, userID string) error

	// ─────────── Relations ───────────
	AddRelation(ctx context.Context, customerID, issueID, relatedIssueID, relationType, userID string) error
	ListRelations(ctx context.Context, customerID, issueID string) ([]models.IssueRelation, error)
	DeleteRelation(ctx context.Context, customerID, relationID, userID string) error

	// ─────────── Comments ───────────
	CreateComment(ctx context.Context, customerID, issueID, userID, body string) (*models.IssueComment, error)
	ListComments(ctx context.Context, customerID, issueID string) ([]models.IssueComment, error)
	UpdateComment(ctx context.Context, customerID, commentID, userID, body string) (*models.IssueComment, error)
	DeleteComment(ctx context.Context, customerID, commentID, userID string) error

	// ─────────── Attachments ───────────
	AddAttachment(ctx context.Context, customerID, issueID, userID string, att *models.IssueAttachment) error
	ListAttachments(ctx context.Context, customerID, issueID string) ([]models.IssueAttachment, error)
	DeleteAttachment(ctx context.Context, customerID, attachmentID, userID string) error

	// ─────────── Checklists ───────────
	CreateChecklist(ctx context.Context, customerID, issueID, title, userID string) (*models.Checklist, error)
	CreateChecklistItem(ctx context.Context, customerID, checklistID, content string, userID string) (*models.ChecklistItem, error)
	UpdateChecklistItem(ctx context.Context, customerID, itemID string, content string, done bool, userID string) (*models.ChecklistItem, error)
	ListChecklists(ctx context.Context, customerID, issueID string) ([]models.Checklist, error)
	DeleteChecklist(ctx context.Context, customerID, checklistID string) error
	DeleteChecklistItem(ctx context.Context, customerID, itemID string) error
	ReorderChecklistItems(ctx context.Context, customerID, checklistID string, order []models.ChecklistItem) error

	// ─────────── Subtasks ───────────
	CreateSubtask(ctx context.Context, customerID, issueID, title string, description *string, assignedTo *string, dueDate *time.Time, userID string) (*models.Subtask, error)
	UpdateSubtask(ctx context.Context, customerID, subtaskID string, title string, description *string, status string, assignedTo *string, dueDate *time.Time, userID string) (*models.Subtask, error)
	ListSubtasks(ctx context.Context, customerID, issueID string) ([]models.Subtask, error)
	DeleteSubtask(ctx context.Context, customerID, subtaskID string) error

	// ─────────── Activity ───────────
	ListActivity(ctx context.Context, customerID, issueID string) ([]models.IssueActivity, error)
}
