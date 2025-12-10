package postgres

import (
	"context"

	"bugforge-backend/internal/models"
)

func (r *KanbanRepo) GetNextOrder(ctx context.Context, columnID string) (int, error) {
    var next int
    err := r.exec.QueryRow(ctx,
        `SELECT COALESCE(MAX("order") + 1, 1)
         FROM issues WHERE column_id = $1`,
        columnID,
    ).Scan(&next)
    return next, err
}

func (r *KanbanRepo) ShiftOrders(ctx context.Context, columnID string, fromOrder int) error {
    _, err := r.exec.Exec(ctx, `
        UPDATE issues
        SET "order" = "order" + 1
        WHERE column_id = $1
          AND "order" >= $2
    `, columnID, fromOrder)
    return err
}

func (r *KanbanRepo) ShiftOrdersDown(ctx context.Context, columnID string, fromOrder int) error {
    _, err := r.exec.Exec(ctx, `
        UPDATE issues
        SET "order" = "order" - 1
        WHERE column_id = $1 AND "order" > $2
    `, columnID, fromOrder)
    return err
}

func (r *KanbanRepo) ShiftRangeUp(ctx context.Context, columnID string, start, end int) error {
    _, err := r.exec.Exec(ctx, `
        UPDATE issues SET "order" = "order" + 1
        WHERE column_id = $1 AND "order" >= $2 AND "order" < $3
    `, columnID, start, end)
    return err
}

func (r *KanbanRepo) ShiftRangeDown(ctx context.Context, columnID string, start, end int) error {
    _, err := r.exec.Exec(ctx, `
        UPDATE issues SET "order" = "order" - 1
        WHERE column_id = $1 AND "order" <= $2 AND "order" > $3
    `, columnID, start, end)
    return err
}

// GetColumnsWithCards loads all columns for a project and their cards.
func (r *KanbanRepo) GetColumnsWithCards(ctx context.Context, projectID string) ([]models.KanbanColumnWithCards, error) {
    columns := []models.KanbanColumnWithCards{}

    // 1. Load all columns
    colRows, err := r.exec.Query(ctx, `
        SELECT id, project_id, name, "order"
        FROM kanban_columns
        WHERE project_id = $1
        ORDER BY "order" ASC
    `, projectID)
    if err != nil {
        return nil, err
    }
    defer colRows.Close()

    for colRows.Next() {
        var col models.KanbanColumnWithCards
        if err := colRows.Scan(&col.ID, &col.ProjectID, &col.Name, &col.Order); err != nil {
            return nil, err
        }

        // 2. Load cards for each column
        cardRows, err := r.exec.Query(ctx, `
            SELECT id, project_id, column_id, title, description, "order",
                   created_by, created_at, updated_at
            FROM issues
            WHERE column_id = $1
            ORDER BY "order" ASC
        `, col.ID)
        if err != nil {
            return nil, err
        }

        cards := []models.Issue{}
        for cardRows.Next() {
            var issue models.Issue
            if err := cardRows.Scan(
                &issue.ID,
                &issue.ProjectID,
                &issue.ColumnID,
                &issue.Title,
                &issue.Description,
                &issue.Order,
                &issue.CreatedBy,
                &issue.CreatedAt,
                &issue.UpdatedAt,
            ); err != nil {
                cardRows.Close()
                return nil, err
            }
            cards = append(cards, issue)
        }
        cardRows.Close()

        col.Cards = cards
        columns = append(columns, col)
    }

    return columns, nil
}

func (r *KanbanRepo) ReorderColumnsShiftUp(
    ctx context.Context,
    projectID string,
    newOrder int,
    oldOrder int,
    columnID string,
) error {

    // Shift other columns down
    _, err := r.exec.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = "order" + 1
        WHERE project_id = $1
          AND "order" >= $2
          AND "order" <  $3
          AND id != $4
    `, projectID, newOrder, oldOrder, columnID)
    if err != nil {
        return err
    }

    // Set new order for target column
    return r.UpdateColumnOrder(ctx, columnID, newOrder)
}

func (r *KanbanRepo) ReorderColumnsShiftDown(
    ctx context.Context,
    projectID string,
    oldOrder int,
    newOrder int,
    columnID string,
) error {

    // Shift other columns up
    _, err := r.exec.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = "order" - 1
        WHERE project_id = $1
          AND "order" <= $2
          AND "order" >  $3
          AND id != $4
    `, projectID, newOrder, oldOrder, columnID)
    if err != nil {
        return err
    }

    // Set new order for target column
    return r.UpdateColumnOrder(ctx, columnID, newOrder)
}
