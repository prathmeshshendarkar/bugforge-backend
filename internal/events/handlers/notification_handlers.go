package handlers

import (
	"fmt"

	"bugforge-backend/internal/events"
	iface "bugforge-backend/internal/service/interfaces"
)

func RegisterNotificationHandlers(ns iface.NotificationService) {

	// When an issue is created and assigned
	events.Register(events.EventIssueCreated, func(e events.Event) {
		assigneeID, ok := e.Data["assignee_id"].(string)
		if !ok || assigneeID == "" {
			return
		}

		issueID := fmt.Sprint(e.Data["issue_id"])

		title := "New Issue Assigned"
		message := fmt.Sprintf("You have been assigned Issue #%s", issueID)

		_ = ns.SendInApp(assigneeID, title, message, fmt.Sprintf(`{"issue_id": "%s"}`, issueID))
		_ = ns.SendEmail(assigneeID, title, message) // you may remove this
	})

	// When a comment is added to an issue
	events.Register(events.EventCommentAdded, func(e events.Event) {
		ownerID, ok := e.Data["owner_id"].(string)
		if !ok || ownerID == "" {
			return
		}

		issueID := fmt.Sprint(e.Data["issue_id"])
		commentID := fmt.Sprint(e.Data["comment_id"])

		title := "New Comment"
		message := fmt.Sprintf("A new comment was added on Issue #%s", issueID)

		_ = ns.SendInApp(ownerID, title, message, fmt.Sprintf(
			`{"issue_id": "%s", "comment_id": "%s"}`, issueID, commentID,
		))
	})
}
