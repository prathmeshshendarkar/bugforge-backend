package models

import "time"

type User struct {
    ID               string    `json:"id" db:"id"`
    CustomerID       string    `json:"customer_id" db:"customer_id"`
    Name      *string    `json:"name"` // nullable
    Username         string    `json:"username" db:"username"`
    Email     string     `json:"email"`
    PasswordHash *string `json:"passwordHash"`
    Role             string    `json:"role" db:"role"` // e.g., super_admin, admin, developer, client
    AssignedProjects []string  `json:"assigned_projects" db:"-"` // normalized in join table
    DefaultProjectID *string   `json:"default_project_id" db:"default_project_id"`
    
    IsPending      bool      `json:"is_pending" db:"is_pending"`
    
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

