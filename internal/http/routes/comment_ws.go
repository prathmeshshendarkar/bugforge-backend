package routes

import (
	commentws "bugforge-backend/internal/websocket/comments"

	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"
)

// /ws/issues/:issueID
func RegisterIssueCommentWS(router fiber.Router, hub *commentws.CommentHub) {
    router.Get("/issues/:issueID", fiberws.New(
        commentws.WSHandler(hub),
    ))
}