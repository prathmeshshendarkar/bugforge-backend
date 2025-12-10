package models

import "time"

type Notification struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Type      string    `json:"type"` // email, in-app, slack
    Title     string    `json:"title"`
    Message   string    `json:"message"`
    Metadata  string    `json:"metadata,omitempty"`
    IsRead    bool      `json:"is_read"`
    CreatedAt time.Time `json:"created_at"`
}
