package models

import "time"

type Project struct {
    ID         string    `json:"id" db:"id"`
    CustomerID string    `json:"customer_id" db:"customer_id"`
    Name       string    `json:"name" db:"name"`
    Slug       string    `json:"slug" db:"slug"` // used for subdomain/route
    CreatedAt  time.Time `json:"created_at" db:"created_at"`
    UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
