package event

// Notifier is the interface for receiving notifications about published events.
// Implementations can use this to monitor event publishing without subscribing to all events.
type Notifier interface {
	// Notify is called when an event is published to the bus.
	//
	// Parameters:
	//   event - The event that was published
	Notify(Event)
}

// NopNotifier is a no-operation implementation of Notifier.
// It discards all notifications, which is the default behavior when no notifier is provided.
type NopNotifier struct{}

// Notify implements the Notifier interface for NopNotifier.
// This method does nothing.
func (n NopNotifier) Notify(Event) {}
