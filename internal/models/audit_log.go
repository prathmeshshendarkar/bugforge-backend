package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
    ID         string          `json:"id" db:"id"`
    CustomerID string          `json:"customer_id" db:"customer_id"`
    UserID     *string         `json:"user_id,omitempty" db:"user_id"`
    Action     string          `json:"action" db:"action"`
    Metadata   json.RawMessage `json:"metadata" db:"metadata"`
    CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}
