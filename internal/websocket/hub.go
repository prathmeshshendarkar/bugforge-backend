package websocket

import "sync"

type Hub struct {
    mu    sync.RWMutex
    rooms map[string]*Room // projectID → Room
}

func NewHub() *Hub {
    return &Hub{rooms: make(map[string]*Room)}
}

func (h *Hub) GetRoom(projectID string) *Room {
    h.mu.RLock()
    r, ok := h.rooms[projectID]
    h.mu.RUnlock()

    if ok {
        return r
    }

    h.mu.Lock()
    defer h.mu.Unlock()

    // double-check
    if r, ok = h.rooms[projectID]; ok {
        return r
    }

    r = NewRoom(projectID)
    h.rooms[projectID] = r
    go r.Run()
    return r
}
