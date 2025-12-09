package models

import "time"

type IssueComment struct {
	ID        string    `json:"id"`
	IssueID   string    `json:"issue_id"`
	UserID    string    `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}
