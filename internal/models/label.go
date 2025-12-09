package models

import "time"

type Label struct {
    ID         string    `json:"id"`
    CustomerID string    `json:"customer_id"`
    ProjectID  string    `json:"project_id"`
    Name       string    `json:"name"`
    Color      string    `json:"color"`
    CreatedAt  time.Time `json:"created_at"`
}
