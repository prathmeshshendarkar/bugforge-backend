package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type KanbanRepository interface {
    // CARD OPERATIONS
    GetCardByID(ctx context.Context, id string) (*models.Issue, error)
    CreateCard(ctx context.Context, card *models.Issue) error
    UpdateCardPosition(ctx context.Context, card *models.Issue) error

    // COLUMN OPERATIONS
    CreateColumn(ctx context.Context, col *models.KanbanColumn) error
    GetNextColumnOrder(ctx context.Context, projectID string) (int, error)

    // ORDERING HELPERS
    GetNextOrder(ctx context.Context, columnID string) (int, error)
    ShiftOrders(ctx context.Context, columnID string, fromOrder int) error

    ShiftOrdersDown(ctx context.Context, columnID string, fromOrder int) error
    ShiftRangeUp(ctx context.Context, columnID string, start, end int) error
    ShiftRangeDown(ctx context.Context, columnID string, start, end int) error
    GetColumnsWithCards(ctx context.Context, projectID string) ([]models.KanbanColumnWithCards, error)

    ReorderColumnsShiftUp(ctx context.Context, projectID string, newOrder, oldOrder int, columnID string) error
    ReorderColumnsShiftDown(ctx context.Context, projectID string, oldOrder, newOrder int, columnID string) error
    UpdateColumnOrder(ctx context.Context, columnID string, newOrder int) error
    UpdateColumnName(ctx context.Context, columnID, name string) error
    DeleteColumn(ctx context.Context, columnID string) error
    DeleteCardsByColumn(ctx context.Context, columnID string) error
    DeleteCard(ctx context.Context, cardID string) error



    // TRANSACTION SUPPORT
    Tx(ctx context.Context, fn func(KanbanRepository) error) error
}
