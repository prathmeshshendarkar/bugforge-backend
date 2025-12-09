package postgres

import (
	"context"

	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ====================================================================
// Root Repo (non-transactional)
// ====================================================================

type KanbanRepoPG struct {
    db *pgxpool.Pool
}

func NewKanbanRepoPG(db *pgxpool.Pool) repo.KanbanRepository {
    return &KanbanRepoPG{db: db}
}

func (r *KanbanRepoPG) GetColumnsWithCards(ctx context.Context, projectID string) ([]models.KanbanColumnWithCards, error) {
    columns := []models.KanbanColumnWithCards{}

    // 1. Load all columns
    colRows, err := r.db.Query(ctx, `
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
        cardRows, err := r.db.Query(ctx, `
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


// ====================================================================
// CARD OPERATIONS
// ====================================================================

func (r *KanbanRepoPG) GetCardByID(ctx context.Context, id string) (*models.Issue, error) {
    var card models.Issue

    query := `
        SELECT id, project_id, column_id, title, description, "order",
               created_by, created_at, updated_at
        FROM issues
        WHERE id = $1
    `
    err := r.db.QueryRow(ctx, query, id).Scan(
        &card.ID,
        &card.ProjectID,
        &card.ColumnID,
        &card.Title,
        &card.Description,
        &card.Order,
        &card.CreatedBy,
        &card.CreatedAt,
        &card.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }

    return &card, nil
}

func (r *KanbanRepoPG) CreateCard(ctx context.Context, card *models.Issue) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO issues (id, project_id, column_id, title, description, "order", created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `,
        card.ID, card.ProjectID, card.ColumnID,
        card.Title, card.Description, card.Order, card.CreatedBy,
    )
    return err
}

func (r *KanbanRepoPG) UpdateCardPosition(ctx context.Context, card *models.Issue) error {
    _, err := r.db.Exec(ctx, `
        UPDATE issues
        SET column_id = $2,
            "order" = $3,
            updated_at = NOW()
        WHERE id = $1
    `,
        card.ID, card.ColumnID, card.Order,
    )
    return err
}

// ====================================================================
// ORDER HELPERS
// ====================================================================

func (r *KanbanRepoPG) GetNextOrder(ctx context.Context, columnID string) (int, error) {
    var next int
    err := r.db.QueryRow(ctx,
        `SELECT COALESCE(MAX("order") + 1, 1)
         FROM issues WHERE column_id = $1`,
        columnID,
    ).Scan(&next)
    return next, err
}

func (r *KanbanRepoPG) ShiftOrders(ctx context.Context, columnID string, fromOrder int) error {
    _, err := r.db.Exec(ctx, `
        UPDATE issues
        SET "order" = "order" + 1
        WHERE column_id = $1
          AND "order" >= $2
    `, columnID, fromOrder)

    return err
}

// ====================================================================
// COLUMN OPERATIONS
// ====================================================================

func (r *KanbanRepoPG) CreateColumn(ctx context.Context, col *models.KanbanColumn) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO kanban_columns (id, project_id, name, "order")
        VALUES ($1, $2, $3, $4)
    `, col.ID, col.ProjectID, col.Name, col.Order)

    return err
}

func (r *KanbanRepoPG) GetNextColumnOrder(ctx context.Context, projectID string) (int, error) {
    var next int
    err := r.db.QueryRow(ctx,
        `SELECT COALESCE(MAX("order") + 1, 1)
         FROM kanban_columns WHERE project_id = $1`,
        projectID,
    ).Scan(&next)
    return next, err
}

// GetColumnsWithCards loads all columns for a project and their cards using the transaction.
func (r *KanbanTxRepoPG) GetColumnsWithCards(ctx context.Context, projectID string) ([]models.KanbanColumnWithCards, error) {
	columns := []models.KanbanColumnWithCards{}

	// 1. Load all columns for the project ordered by column order
	colRows, err := r.tx.Query(ctx, `
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

		// 2. Load cards for this column
		cardRows, err := r.tx.Query(ctx, `
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

// ====================================================================
// TRANSACTION SUPPORT
// ====================================================================

func (r *KanbanRepoPG) Tx(ctx context.Context, fn func(repo.KanbanRepository) error) error {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return err
    }

    txRepo := &KanbanTxRepoPG{tx: tx}

    if err := fn(txRepo); err != nil {
        tx.Rollback(ctx)
        return err
    }

    return tx.Commit(ctx)
}

// ====================================================================
// Transactional Repo
// ====================================================================

type KanbanTxRepoPG struct {
    tx pgx.Tx
}

// ---------------------- CARD ----------------------------------------

func (r *KanbanTxRepoPG) GetCardByID(ctx context.Context, id string) (*models.Issue, error) {
    var card models.Issue

    query := `
        SELECT id, project_id, column_id, title, description, "order",
               created_by, created_at, updated_at
        FROM issues
        WHERE id = $1
    `
    err := r.tx.QueryRow(ctx, query, id).Scan(
        &card.ID,
        &card.ProjectID,
        &card.ColumnID,
        &card.Title,
        &card.Description,
        &card.Order,
        &card.CreatedBy,
        &card.CreatedAt,
        &card.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }

    return &card, nil
}

func (r *KanbanTxRepoPG) CreateCard(ctx context.Context, card *models.Issue) error {
    _, err := r.tx.Exec(ctx, `
        INSERT INTO issues (id, project_id, column_id, title, description, "order", created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `,
        card.ID, card.ProjectID, card.ColumnID,
        card.Title, card.Description, card.Order, card.CreatedBy,
    )
    return err
}

func (r *KanbanTxRepoPG) UpdateCardPosition(ctx context.Context, card *models.Issue) error {
    _, err := r.tx.Exec(ctx, `
        UPDATE issues
        SET column_id = $2,
            "order" = $3,
            updated_at = NOW()
        WHERE id = $1
    `,
        card.ID, card.ColumnID, card.Order,
    )
    return err
}

// ---------------------- ORDERING ------------------------------------

func (r *KanbanTxRepoPG) GetNextOrder(ctx context.Context, columnID string) (int, error) {
    var next int
    err := r.tx.QueryRow(ctx,
        `SELECT COALESCE(MAX("order") + 1, 1)
         FROM issues WHERE column_id = $1`,
        columnID,
    ).Scan(&next)
    return next, err
}

func (r *KanbanTxRepoPG) ShiftOrders(ctx context.Context, columnID string, fromOrder int) error {
    _, err := r.tx.Exec(ctx, `
        UPDATE issues
        SET "order" = "order" + 1
        WHERE column_id = $1 AND "order" >= $2
    `, columnID, fromOrder)

    return err
}

// ---------------------- COLUMNS -------------------------------------

func (r *KanbanTxRepoPG) CreateColumn(ctx context.Context, col *models.KanbanColumn) error {
    _, err := r.tx.Exec(ctx, `
        INSERT INTO kanban_columns (id, project_id, name, "order")
        VALUES ($1, $2, $3, $4)
    `,
        col.ID, col.ProjectID, col.Name, col.Order,
    )
    return err
}

func (r *KanbanTxRepoPG) GetNextColumnOrder(ctx context.Context, projectID string) (int, error) {
    var next int
    err := r.tx.QueryRow(ctx,
        `SELECT COALESCE(MAX("order") + 1, 1)
         FROM kanban_columns WHERE project_id = $1`,
        projectID,
    ).Scan(&next)
    return next, err
}

func (r *KanbanRepoPG) ShiftOrdersDown(ctx context.Context, columnID string, fromOrder int) error {
    _, err := r.db.Exec(ctx, `
        UPDATE issues
        SET "order" = "order" - 1
        WHERE column_id = $1 AND "order" > $2
    `, columnID, fromOrder)
    return err
}

func (r *KanbanTxRepoPG) ShiftOrdersDown(ctx context.Context, columnID string, fromOrder int) error {
    _, err := r.tx.Exec(ctx, `
        UPDATE issues
        SET "order" = "order" - 1
        WHERE column_id = $1 AND "order" > $2
    `, columnID, fromOrder)
    return err
}

func (r *KanbanTxRepoPG) ShiftRangeUp(ctx context.Context, columnID string, start, end int) error {
    _, err := r.tx.Exec(ctx, `
        UPDATE issues SET "order" = "order" + 1
        WHERE column_id = $1 AND "order" >= $2 AND "order" < $3
    `, columnID, start, end)
    return err
}

func (r *KanbanTxRepoPG) ShiftRangeDown(ctx context.Context, columnID string, start, end int) error {
    _, err := r.tx.Exec(ctx, `
        UPDATE issues SET "order" = "order" - 1
        WHERE column_id = $1 AND "order" <= $2 AND "order" > $3
    `, columnID, start, end)
    return err
}

func (r *KanbanRepoPG) ShiftRangeUp(ctx context.Context, columnID string, start, end int) error {
    _, err := r.db.Exec(ctx, `
        UPDATE issues SET "order" = "order" + 1
        WHERE column_id = $1 AND "order" >= $2 AND "order" < $3
    `, columnID, start, end)
    return err
}

func (r *KanbanRepoPG) ShiftRangeDown(ctx context.Context, columnID string, start, end int) error {
    _, err := r.db.Exec(ctx, `
        UPDATE issues SET "order" = "order" - 1
        WHERE column_id = $1 AND "order" <= $2 AND "order" > $3
    `, columnID, start, end)
    return err
}

func (r *KanbanTxRepoPG) ReorderColumnsShiftUp(
    ctx context.Context,
    projectID string,
    newOrder int,
    oldOrder int,
    columnID string,
) error {

    // Shift other columns down
    _, err := r.tx.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = "order" + 1
        WHERE project_id = $1
          AND "order" >= $2
          AND "order" <  $3
          AND id != $4
    `, projectID, newOrder, oldOrder, columnID)
    if err != nil { return err }

    // Set new order for target column
    return r.UpdateColumnOrder(ctx, columnID, newOrder)
}

func (r *KanbanTxRepoPG) ReorderColumnsShiftDown(
    ctx context.Context,
    projectID string,
    oldOrder int,
    newOrder int,
    columnID string,
) error {

    // Shift other columns up
    _, err := r.tx.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = "order" - 1
        WHERE project_id = $1
          AND "order" <= $2
          AND "order" >  $3
          AND id != $4
    `, projectID, newOrder, oldOrder, columnID)
    if err != nil { return err }

    // Set new order for target column
    return r.UpdateColumnOrder(ctx, columnID, newOrder)
}


func (r *KanbanTxRepoPG) UpdateColumnOrder(ctx context.Context, columnID string, newOrder int) error {
    _, err := r.tx.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = $2
        WHERE id = $1
    `, columnID, newOrder)
    return err
}

// ---------------------------------------------------------------
// Column reorder operations (Non-TX)
// ---------------------------------------------------------------

func (r *KanbanRepoPG) ReorderColumnsShiftUp(
    ctx context.Context,
    projectID string,
    newOrder int,
    oldOrder int,
    columnID string,
) error {

    // Shift others down
    _, err := r.db.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = "order" + 1
        WHERE project_id = $1
          AND "order" >= $2
          AND "order" <  $3
          AND id != $4
    `, projectID, newOrder, oldOrder, columnID)
    if err != nil { return err }

    // Update this column's order
    return r.UpdateColumnOrder(ctx, columnID, newOrder)
}

func (r *KanbanRepoPG) ReorderColumnsShiftDown(
    ctx context.Context,
    projectID string,
    oldOrder int,
    newOrder int,
    columnID string,
) error {

    // Shift others up
    _, err := r.db.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = "order" - 1
        WHERE project_id = $1
          AND "order" <= $2
          AND "order" >  $3
          AND id != $4
    `, projectID, newOrder, oldOrder, columnID)
    if err != nil { return err }

    // Update this column
    return r.UpdateColumnOrder(ctx, columnID, newOrder)
}

func (r *KanbanRepoPG) UpdateColumnOrder(ctx context.Context, columnID string, newOrder int) error {
    _, err := r.db.Exec(ctx, `
        UPDATE kanban_columns
        SET "order" = $2
        WHERE id = $1
    `, columnID, newOrder)
    return err
}

func (r *KanbanRepoPG) UpdateColumnName(ctx context.Context, columnID, name string) error {
    _, err := r.db.Exec(ctx, `
        UPDATE kanban_columns
        SET name = $2
        WHERE id = $1
    `, columnID, name)
    return err
}

func (r *KanbanTxRepoPG) UpdateColumnName(ctx context.Context, columnID, name string) error {
    _, err := r.tx.Exec(ctx, `
        UPDATE kanban_columns
        SET name = $2
        WHERE id = $1
    `, columnID, name)
    return err
}

func (r *KanbanRepoPG) DeleteCardsByColumn(ctx context.Context, columnID string) error {
    _, err := r.db.Exec(ctx, `DELETE FROM issues WHERE column_id = $1`, columnID)
    return err
}

func (r *KanbanRepoPG) DeleteColumn(ctx context.Context, columnID string) error {
    _, err := r.db.Exec(ctx, `DELETE FROM kanban_columns WHERE id = $1`, columnID)
    return err
}

func (r *KanbanTxRepoPG) DeleteCardsByColumn(ctx context.Context, columnID string) error {
    _, err := r.tx.Exec(ctx, `DELETE FROM issues WHERE column_id = $1`, columnID)
    return err
}

func (r *KanbanTxRepoPG) DeleteColumn(ctx context.Context, columnID string) error {
    _, err := r.tx.Exec(ctx, `DELETE FROM kanban_columns WHERE id = $1`, columnID)
    return err
}

func (r *KanbanRepoPG) DeleteCard(ctx context.Context, cardID string) error {
    _, err := r.db.Exec(ctx, `
        DELETE FROM issues WHERE id = $1
    `, cardID)
    return err
}

func (r *KanbanTxRepoPG) DeleteCard(ctx context.Context, cardID string) error {
    _, err := r.tx.Exec(ctx, `
        DELETE FROM issues WHERE id = $1
    `, cardID)
    return err
}

// ---------------------- TX SUPPORT ----------------------------------

func (r *KanbanTxRepoPG) Tx(ctx context.Context, fn func(repo.KanbanRepository) error) error {
    // Nested TX NOT supported with pgx.
    return fn(r)
}

