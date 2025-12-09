package websocket

import "encoding/json"

type MoveCardPayload struct {
    CardID     string `json:"card_id"`
    FromColumn string `json:"from_column"`
    ToColumn   string `json:"to_column"`
    NewOrder   int    `json:"new_order"`
}

func HandleMoveCardIntent(evt IncomingEvent, userID string) (OutEvent, error) {
    // parse
    raw, _ := json.Marshal(evt.Payload)
    var mp MoveCardPayload
    json.Unmarshal(raw, &mp)

    // TODO: call service.Placement logic
    // kanbanService.MoveCard(mp.CardID, ...)

    // Broadcast event
    return OutEvent{
        Type:      "card_moved",
        ProjectID: evt.ProjectID,
        ActorID:   userID,
        Payload:   mp,
    }, nil
}
