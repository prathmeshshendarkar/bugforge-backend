package websocket

type CommentEvent struct {
	Type    string      `json:"type"` // comment_created, comment_updated, deleted, mention
	IssueID string      `json:"issue_id"`
	ActorID string      `json:"actor_id"`
	Payload interface{} `json:"payload"`
}
