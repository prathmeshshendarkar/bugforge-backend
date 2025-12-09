package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type CustomerRepository interface {
    GetByID(ctx context.Context, id string) (*models.Customer, error)
}
