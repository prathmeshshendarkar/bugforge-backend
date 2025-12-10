package service

import (
	"bugforge-backend/internal/models"
	repo "bugforge-backend/internal/repository/interfaces"
	"context"
	"time"

	"github.com/google/uuid"
)

type ActivityServiceImpl struct {
    repo repo.ActivityRepository
}

func NewActivityService(repo repo.ActivityRepository) *ActivityServiceImpl {
    return &ActivityServiceImpl{repo: repo}
}

func (a *ActivityServiceImpl) Log(
    ctx context.Context,
    issueID string,
    userID *string,
    action string,
    meta map[string]interface{},
) error {

    entry := &models.IssueActivity{
        ID:        uuid.NewString(),
        IssueID:   issueID,
        UserID:    userID,
        Action:    action,
        Metadata:  meta,
        CreatedAt: time.Now(),
    }

    return a.repo.Create(ctx, entry)
}
