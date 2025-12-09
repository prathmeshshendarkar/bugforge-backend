package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (*models.User, string, error)
	AcceptInvite(ctx context.Context, token, name, password string) (*models.User, error)
}
