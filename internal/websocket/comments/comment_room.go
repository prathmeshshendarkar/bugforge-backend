package websocket

import "sync"

type CommentRoom struct {
	issueID    string
	clients    map[*CommentClient]bool
	register   chan *CommentClient
	unregister chan *CommentClient
	broadcast  chan []byte
	mu         sync.RWMutex
}

func NewCommentRoom(issueID string) *CommentRoom {
	return &CommentRoom{
		issueID:    issueID,
		clients:    make(map[*CommentClient]bool),
		register:   make(chan *CommentClient),
		unregister: make(chan *CommentClient),
		broadcast:  make(chan []byte, 256),
	}
}

func (r *CommentRoom) Run() {
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
		}
	}
}

func (r *CommentRoom) Broadcast(msg []byte) {
	r.broadcast <- msg
}
