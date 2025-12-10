package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetAllByCustomer(ctx context.Context, customerID string) ([]models.User, error)
	Update(ctx context.Context, u *models.User) error
	Delete(ctx context.Context, id, customerID string) error
	DeleteInvitesByUserID(ctx context.Context, userID string) error

	// project assignment helpers
	AssignProjects(ctx context.Context, userID string, projectIDs []string) error
	DeleteProjectAssignments(ctx context.Context, userID string) error
	GetAssignedProjectIDs(ctx context.Context, userID string) ([]string, error)

	CreatePending(ctx context.Context, u *models.User) error
	GetByInviteToken(ctx context.Context, token string) (*models.User, error)
	SaveInviteToken(ctx context.Context, userID, token string) error
	MarkInviteAccepted(ctx context.Context, userID string) error
}
