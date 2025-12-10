package postgres

import (
	"context"

	"bugforge-backend/internal/models"
)

func (r *KanbanRepo) GetCardByID(ctx context.Context, id string) (*models.Issue, error) {
    var card models.Issue

    query := `
        SELECT id, project_id, column_id, title, description, "order",
               created_by, created_at, updated_at
        FROM issues
        WHERE id = $1
    `
    err := r.exec.QueryRow(ctx, query, id).Scan(
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

func (r *KanbanRepo) CreateCard(ctx context.Context, card *models.Issue) error {
    _, err := r.exec.Exec(ctx, `
        INSERT INTO issues (id, project_id, column_id, title, description, "order", created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `,
        card.ID, card.ProjectID, card.ColumnID,
        card.Title, card.Description, card.Order, card.CreatedBy,
    )
    return err
}

func (r *KanbanRepo) UpdateCardPosition(ctx context.Context, card *models.Issue) error {
    _, err := r.exec.Exec(ctx, `
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

func (r *KanbanRepo) DeleteCard(ctx context.Context, cardID string) error {
    _, err := r.exec.Exec(ctx, `DELETE FROM issues WHERE id = $1`, cardID)
    return err
}
