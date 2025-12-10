package handlers

import (
	"context"
	"fmt"

	"bugforge-backend/internal/events"
	iface "bugforge-backend/internal/service/interfaces"
)

func RegisterActivityHandlers(as iface.ActivityService) {

	events.Register(events.EventIssueCreated, func(e events.Event) {
		issueID := fmt.Sprint(e.Data["issue_id"])
		userID := fmt.Sprint(e.Data["created_by"])

		ctx := context.Background()
		userPtr := &userID

		as.Log(ctx, issueID, userPtr, "issue_created", map[string]interface{}{
			"description": fmt.Sprintf("Created issue %s", issueID),
		})
	})

	events.Register(events.EventCommentAdded, func(e events.Event) {
		issueID := fmt.Sprint(e.Data["issue_id"])
		userID := fmt.Sprint(e.Data["commented_by"])

		ctx := context.Background()
		userPtr := &userID

		as.Log(ctx, issueID, userPtr, "comment_added", map[string]interface{}{
			"description": "Added a comment",
		})
	})
}
