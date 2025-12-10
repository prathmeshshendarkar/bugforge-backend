package websocket

import (
	"encoding/json"
	"sync"
)

type CommentHub struct {
	mu    sync.RWMutex
	rooms map[string]*CommentRoom
}

func NewCommentHub() *CommentHub {
	return &CommentHub{
		rooms: make(map[string]*CommentRoom),
	}
}

func (h *CommentHub) GetRoom(issueID string) *CommentRoom {
	h.mu.RLock()
	room, ok := h.rooms[issueID]
	h.mu.RUnlock()

	if ok {
		return room
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check
	if room, ok = h.rooms[issueID]; ok {
		return room
	}

	room = NewCommentRoom(issueID)
	h.rooms[issueID] = room
	go room.Run()

	return room
}

func (h *CommentHub) BroadcastToUser(userID string, evt CommentEvent) {
	msg, _ := json.Marshal(evt)

	h.mu.RLock()
	for _, room := range h.rooms {
		room.mu.RLock()
		for c := range room.clients {
			if c.userID == userID {
				c.send <- msg
			}
		}
		room.mu.RUnlock()
	}
	h.mu.RUnlock()
}
