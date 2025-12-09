package models

import "time"

type KanbanColumn struct {
    ID        string    `json:"id" db:"id"`
    ProjectID string    `json:"project_id" db:"project_id"`
    Name      string    `json:"name" db:"name"`
    Order     int       `json:"order" db:"order"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
