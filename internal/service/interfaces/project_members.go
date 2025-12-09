package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type ProjectMemberService interface {
    AddMember(ctx context.Context, projectID, customerID, userID string) error
    RemoveMember(ctx context.Context, projectID, customerID, userID string) error
    ListMembers(ctx context.Context, projectID, customerID string) ([]models.User, error)
    Invite(ctx context.Context, projectID, customerID, email, role string) error
}
