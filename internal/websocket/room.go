package websocket

import (
	"sync"
	"time"
)

type Room struct {
    projectID  string
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    done       chan struct{}
    mu         sync.RWMutex
}

func NewRoom(projectID string) *Room {
    return &Room{
        projectID:  projectID,
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte, 256),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        done:       make(chan struct{}),
    }
}

func (r *Room) Run() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case c := <-r.register:
            r.mu.Lock()
            r.clients[c] = true
            r.mu.Unlock()

        case c := <-r.unregister:
            r.mu.Lock()
            if _, ok := r.clients[c]; ok {
                delete(r.clients, c)
                close(c.send)
            }
            r.mu.Unlock()

        case msg := <-r.broadcast:
            r.mu.RLock()
            for c := range r.clients {
                select {
                case c.send <- msg:
                default:
                    close(c.send)
                    delete(r.clients, c)
                }
            }
            r.mu.RUnlock()

        case <-ticker.C:
            // keepalive / cleanup

        case <-r.done:
            return
        }
    }
}

func (r *Room) Broadcast(msg []byte) {
    r.broadcast <- msg
}

