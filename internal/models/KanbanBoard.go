package models

type KanbanBoard struct {
	Columns []KanbanColumnWithCards `json:"columns"`
}

type KanbanColumnWithCards struct {
	ID        string  `json:"id"`
	ProjectID string  `json:"project_id"`
	Name      string  `json:"name"`
	Order     int     `json:"order"`
	Cards     []Issue `json:"cards"` // issues belonging to this column
}
