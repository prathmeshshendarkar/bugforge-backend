package models

import "time"

type IssueActivity struct {
	ID        string                 `json:"id"`
	IssueID   string                 `json:"issue_id"`
	UserID    *string                `json:"user_id,omitempty"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	IssueTitle *string `json:"issue_title,omitempty"`
}
