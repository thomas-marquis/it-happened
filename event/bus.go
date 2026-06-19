package event

// Bus is the interface for publishing and subscribing to events.
// Implementations of this interface manage the delivery of events to subscribers.
type Bus interface {

	// Publish publishes an event to the bus.
	// The event will be delivered to all matching subscribers.
	Publish(evt Event)

	// Subscribe creates a new subscriber for this bus.
	// The subscriber can register callbacks for specific event matchers.
	Subscribe() *Subscriber

	// Unsubscribe removes a subscriber from the bus.
	Unsubscribe(sub *Subscriber)
}
