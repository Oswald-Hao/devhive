package orchestrator

import "sync"

// EventBus is a pub/sub event bus using Go channels.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan *Event
	bus         chan *Event
	quit        chan struct{}
}

// Event represents a system event.
type Event struct {
	Type    string
	TaskID  string
	Payload map[string]interface{}
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan *Event),
		bus:         make(chan *Event, 256),
		quit:        make(chan struct{}),
	}
}

// Subscribe registers a callback for a specific event type.
// Returns a channel that receives matching events.
func (eb *EventBus) Subscribe(eventType string) <-chan *Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan *Event, 64)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

// Publish sends an event to all subscribers.
func (eb *EventBus) Publish(event *Event) {
	eb.bus <- event
}

// Start begins the event dispatch loop.
func (eb *EventBus) Start() {
	go func() {
		for {
			select {
			case event := <-eb.bus:
				eb.mu.RLock()
				subs := eb.subscribers[event.Type]
				// Also send to "*" subscribers (catch-all)
				subs = append(subs, eb.subscribers["*"]...)
				eb.mu.RUnlock()
				for _, ch := range subs {
					select {
					case ch <- event:
					default:
						// Drop if subscriber is too slow
					}
				}
			case <-eb.quit:
				return
			}
		}
	}()
}

// Stop shuts down the event bus.
func (eb *EventBus) Stop() {
	close(eb.quit)
}
