package events

type EventName string

// Add your event names here
const (
	EventIssueCreated EventName = "issue.created"
	EventCommentAdded EventName = "comment.added"
)

// Event payload
type Event struct {
	Name EventName
	Data map[string]interface{}
}
