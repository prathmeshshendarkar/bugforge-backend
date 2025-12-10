package interfaces

import (
	"context"
)

type ActivityService interface {
    Log(ctx context.Context, issueID string, userID *string, action string, meta map[string]interface{}) error
}
