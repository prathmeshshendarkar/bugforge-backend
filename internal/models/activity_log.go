package models

import "time"

type IssueActivity struct {
	ID        string      `json:"id"`
	IssueID   string      `json:"issue_id"`
	UserID    *string     `json:"user_id,omitempty"`
	Action    string      `json:"action"`    // e.g., created, status_changed, assigned, commented
	Metadata  interface{} `json:"metadata"`  // any JSON-friendly struct/map
	CreatedAt time.Time   `json:"created_at"`
}
