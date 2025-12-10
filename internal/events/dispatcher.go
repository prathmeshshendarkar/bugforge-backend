package events

import "sync"

var (
	mu       sync.RWMutex
	registry = make(map[EventName][]func(Event))
)

// Register adds a handler for an event.
func Register(name EventName, handler func(Event)) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = append(registry[name], handler)
}

// Dispatch triggers all handlers for an event.
func Dispatch(e Event) {
	mu.RLock()
	handlers := registry[e.Name]
	mu.RUnlock()

	for _, h := range handlers {
		go h(e) // async so nothing blocks
	}
}
