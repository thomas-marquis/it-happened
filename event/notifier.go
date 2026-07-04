package event

// Notifier is the interface for receiving notifications about published events.
// Implementations can use this to monitor event publishing without subscribing to all events.
type Notifier interface {
	// NotifyPublished is called when an event is published to the bus.
	// This method is called once per event, whether one or more subscribers listen for it or not.
	//
	// Parameters:
	//   event - The event that was published
	NotifyPublished(Event)

	// NotifySubscribed is called when a new subscriber is added to the bus
	NotifySubscribed(subscriber *Subscriber)

	// NotifyUnsubscribed is called when a subscriber is removed
	NotifyUnsubscribed(subscriber *Subscriber)
}

// NopNotifier is a no-operation implementation of Notifier.
// It discards all notifications, which is the default behavior when no notifier is provided.
type NopNotifier struct{}

var _ Notifier = (*NopNotifier)(nil)

func (n NopNotifier) NotifyPublished(Event) {}

func (n NopNotifier) NotifySubscribed(*Subscriber) {}

func (n NopNotifier) NotifyUnsubscribed(*Subscriber) {}
