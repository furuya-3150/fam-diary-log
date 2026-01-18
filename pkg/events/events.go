package events

// Event is the base interface for all events
type Event interface {
	EventType() string
}
