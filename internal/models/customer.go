package models

import "time"

type Customer struct {
	Id string
	Name string
	CreatedAt time.Time
	UpdatedAt time.Time
}