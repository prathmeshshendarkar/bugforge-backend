package interfaces

import (
	"bugforge-backend/internal/models"
)

type KanbanService interface {
    GetBoard(projectID string, userID string) (*models.KanbanBoard, error)
    MoveCard(
    cardID string,
    toColumnID string,
    newOrder int,
    userID string,
) (*models.Issue, string, error) 
    CreateCard(projectID, columnID, title, description, userID string) (*models.Issue, error)
    CreateColumn(projectID, name string, userID string) (*models.KanbanColumn, error)
}
