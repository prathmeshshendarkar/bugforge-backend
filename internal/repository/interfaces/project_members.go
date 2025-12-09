package interfaces

import (
	"bugforge-backend/internal/models"
	"context"
)

type ProjectMemberRepository interface {
    AddMember(ctx context.Context, projectID, userID string) error
    RemoveMember(ctx context.Context, projectID, userID string) error
    ListMembers(ctx context.Context, projectID, customerID string) ([]models.User, error)
    IsMember(ctx context.Context, projectID, userID string) (bool, error)

    GetAssignedProjectIDsForUser(ctx context.Context, userID string) ([]string, error)
    SyncMembersForUser(ctx context.Context, userID, customerID string, projectIDs []string) error
}
