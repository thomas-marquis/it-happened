package event

import (
	"encoding/json"
	"log"
	"sync"
)

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

// HarvesterNotifier is a notifier that harvests all published events and store them into a list.
// Do not use in production: this can lead to memory leaks.
type HarvesterNotifier struct {
	NopNotifier
	mu     sync.Mutex
	Events []Event
}

func (n *HarvesterNotifier) NotifyPublished(evt Event) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Events = append(n.Events, evt)
}

// LoggerNotifier is a notifier that logs all published events with the provided logger.
// The logger can't be nil.
type LoggerNotifier struct {
	Logger log.Logger
}

func (n *LoggerNotifier) NotifyPublished(evt Event) {
	content, err := json.MarshalIndent(evt, "", "  ")
	if err != nil {
		n.Logger.Println("event emitted (unmarshalled)")
		return
	}
	n.Logger.Println(string(content))
}

func (n *LoggerNotifier) NotifySubscribed(*Subscriber) {
	n.Logger.Print("subscribed")
}

func (n *LoggerNotifier) NotifyUnsubscribed(*Subscriber) {
	n.Logger.Print("unsubscribed")
}

// CombinedNotifier allows you to merge multiple notifiers at once.
// Each registered notifier will be notified of all events.
type CombinedNotifier struct {
	mu        sync.RWMutex
	Notifiers []Notifier
}

// NewCombinedNotifier creates a new CombinedNotifier with the provided notifiers.
func NewCombinedNotifier(notifiers ...Notifier) *CombinedNotifier {
	return &CombinedNotifier{
		Notifiers: notifiers,
	}
}

func (n *CombinedNotifier) Add(notifier Notifier) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Notifiers = append(n.Notifiers, notifier)
}

func (n *CombinedNotifier) NotifyPublished(evt Event) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, notifier := range n.Notifiers {
		notifier.NotifyPublished(evt)
	}
}

func (n *CombinedNotifier) NotifySubscribed(sub *Subscriber) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, notifier := range n.Notifiers {
		notifier.NotifySubscribed(sub)
	}
}

func (n *CombinedNotifier) NotifyUnsubscribed(sub *Subscriber) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, notifier := range n.Notifiers {
		notifier.NotifyUnsubscribed(sub)
	}
}
