package websocket

type IncomingEvent struct {
	Type      string      `json:"type"`
	ProjectID string      `json:"project_id"`
	Payload   interface{} `json:"payload"`
}

type OutEvent struct {
	Type      string      `json:"type"`
	ProjectID string      `json:"project_id"`
	ActorID   string      `json:"actor_id"`
	Payload   interface{} `json:"payload"`
}
