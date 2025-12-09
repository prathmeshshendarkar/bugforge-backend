package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type UserService interface {
	CreateUser(ctx context.Context, customerID, name, email, password, role string, assignedProjectIDs []string, defaultProjectID *string) (*models.User, error)
	GetByID(ctx context.Context, id, customerID string) (*models.User, error)
	GetByEmail(ctx context.Context, email, customerID string) (*models.User, error)
	GetAllByCustomer(ctx context.Context, customerID string) ([]models.User, error)
	UpdateUser(ctx context.Context, id, customerID, name, email, password, role string, assignedProjectIDs []string, defaultProjectID *string) (*models.User, error)
	DeleteUser(ctx context.Context, id, customerID string) error
}
