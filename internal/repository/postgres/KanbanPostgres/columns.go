package postgres

import (
	"context"

	"bugforge-backend/internal/models"
)

func (r *KanbanRepo) CreateColumn(ctx context.Context, col *models.KanbanColumn) error {
    _, err := r.exec.Exec(ctx, `
        INSERT INTO kanban_columns (id, project_id, name, "order")
        VALUES ($1, $2, $3, $4)
    `, col.ID, col.ProjectID, col.Name, col.Order)
    return err
}

func (r *KanbanRepo) GetNextColumnOrder(ctx context.Context, projectID string) (int, error) {
    var next int
    err := r.exec.QueryRow(ctx,
        `SELECT COALESCE(MAX("order") + 1, 1)
         FROM kanban_columns WHERE project_id = $1`,
        projectID,
    ).Scan(&next)
    return next, err
}

func (r *KanbanRepo) UpdateColumnOrder(ctx context.Context, columnID string, newOrder int) error {
    _, err := r.exec.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = $2
        WHERE id = $1
    `, columnID, newOrder)
    return err
}

func (r *KanbanRepo) UpdateColumnName(ctx context.Context, columnID, name string) error {
    _, err := r.exec.Exec(ctx, `
        UPDATE kanban_columns
        SET name = $2
        WHERE id = $1
    `, columnID, name)
    return err
}

func (r *KanbanRepo) DeleteColumn(ctx context.Context, columnID string) error {
    _, err := r.exec.Exec(ctx, `DELETE FROM kanban_columns WHERE id = $1`, columnID)
    return err
}

func (r *KanbanRepo) DeleteCardsByColumn(ctx context.Context, columnID string) error {
    _, err := r.exec.Exec(ctx, `DELETE FROM issues WHERE column_id = $1`, columnID)
    return err
}
