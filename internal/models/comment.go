package models

import "time"

type IssueComment struct {
    ID          string     `json:"id"`
    IssueID     string     `json:"issue_id"`
    UserID      string     `json:"user_id"`
    Body        string     `json:"body"`
    BodyHTML    *string    `json:"body_html"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at,omitempty"`

    AuthorName  *string    `json:"author_name,omitempty"`
    AuthorEmail *string    `json:"author_email,omitempty"`
}
