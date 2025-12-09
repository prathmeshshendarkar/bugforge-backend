package models

import "time"

type Issue struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	ColumnID    string     `json:"column_id"`
  Order       int        `json:"order"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	CreatedBy   string     `json:"created_by"`
	AssignedTo  *string    `json:"assigned_to,omitempty"`
  DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type IssueWithUser struct {
	ID              string     `json:"id"`
	ProjectID       string     `json:"project_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	Priority        string     `json:"priority"`
  DueDate     *time.Time `json:"due_date"`

	CreatedBy       string     `json:"created_by"`
	CreatedByEmail  *string    `json:"created_by_email"`
	CreatedByName   *string    `json:"created_by_name"`

	AssignedTo      *string    `json:"assigned_to"`
	AssignedToEmail *string    `json:"assigned_to_email"`
	AssignedToName  *string    `json:"assigned_to_name"`

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type IssueAttachment struct {
  ID        string    `json:"id"`
  IssueID   string    `json:"issue_id"`
  UserID    string    `json:"user_id"`
  URL       string    `json:"url"`
  Key       string    `json:"key"`
  Filename  string    `json:"filename"`
  ContentType *string `json:"content_type"`
  Size      int64     `json:"size"`
  CreatedAt time.Time `json:"created_at"`
}

type Checklist struct {
    ID        string          `json:"id"`
    IssueID   string          `json:"issue_id"`
    Title     string          `json:"title"`
    CreatedAt time.Time       `json:"created_at"`
    Items     []ChecklistItem `json:"items"` // NEW FIELD
}

type ChecklistItem struct {
  ID         string  `json:"id"`
  ChecklistID string `json:"checklist_id"`
  Content    string  `json:"content"`
  Done       bool    `json:"done"`
  OrderIndex int     `json:"order_index"`
}

type Subtask struct {
  ID           string     `json:"id"`
  ParentIssueID string    `json:"parent_issue_id"`
  Title        string     `json:"title"`
  Description  *string    `json:"description"`
  Status       string     `json:"status"`
  AssignedTo   *string    `json:"assigned_to"`
  DueDate      *time.Time `json:"due_date"`
  OrderIndex   int        `json:"order_index"`
  CreatedAt    time.Time  `json:"created_at"`
}

type IssueRelation struct {
  ID              string    `json:"id"`
  IssueID         string    `json:"issue_id"`
  RelatedIssueID  string    `json:"related_issue_id"`
  RelationType    string    `json:"relation_type"`
  CreatedAt       time.Time `json:"created_at"`
}
