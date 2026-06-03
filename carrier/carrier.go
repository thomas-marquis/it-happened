package carrier

import (
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

const (
	// TypePrefix defines the prefix used to identify carrier events.
	TypePrefix = "__carrier__"

	defaultCarrierTimeout     = 60 * time.Second
	defaultCarrierConcurrency = 10
)

// Carrier describes what an event carrier's payload be like.
// An event carrier is an event with a special kind of payload that allows dispatching multiple events to the bus.
type Carrier interface {
	event.Payload

	// Dispatch dispatches all events in the carrier to the given channel.
	// Depending on bus implementation, this may be blocking or non-blocking.
	Dispatch(bus event.Bus)
}

type carrierConfig struct {
	maxConcurrency      int
	timeout             time.Duration
	completionCondition CompletionCondition
}

// Option allows configuring a carrier on creation.
type Option func(config *carrierConfig)

func WithTimeout(d time.Duration) Option {
	return func(c *carrierConfig) {
		c.timeout = d
	}
}

func WithMaxConcurrency(n int) Option {
	return func(c *carrierConfig) {
		c.maxConcurrency = n
	}
}

func WithCompletionCondition(cond CompletionCondition) Option {
	return func(c *carrierConfig) {
		c.completionCondition = cond
	}
}

// CompletionCondition is a type of function that defines when an event emitted by a carrier is considered as completed.
// By default, all carriers will consider only followup events (from the ones they sent). So, both event expected as CompletionCondition parameters share the same Ref.
type CompletionCondition func(sent, received event.Event) bool

// CompletedOnFollowupReceived is a completion condition that returns true when the received event is a followup of the one sent (they both share the same Ref).
// Due to the current carriers implementation, this condition is always true (by construction: we already know that the received event is a followup of the one sent).
// Default completion condition function.
func CompletedOnFollowupReceived(sent, received event.Event) bool {
	return true
}
