package carrier

import (
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

const (
	// TypePrefix defines the prefix used to identify carrier events.
	// All carrier event types start with this prefix.
	TypePrefix = "__carrier__"

	// defaultCarrierTimeout is the default timeout for carrier operations.
	defaultCarrierTimeout = 60 * time.Second
	// defaultCarrierConcurrency is the default maximum number of concurrent operations for carriers.
	defaultCarrierConcurrency = 10
)

// Carrier describes what an event carrier's payload must be like.
// An event carrier is an event with a special kind of payload that allows dispatching multiple events to the bus.
type Carrier interface {
	event.Payload

	// Dispatch dispatches all events in the carrier to the given bus.
	// Depending on bus implementation, this may be blocking or non-blocking.
	//
	// Parameters:
	//   bus - The event bus to dispatch events to
	Dispatch(bus event.Bus)
}

// carrierConfig holds configuration options for a carrier.
type carrierConfig struct {
	maxConcurrency      int
	timeout             time.Duration
	completionCondition CompletionCondition
}

// Option allows configuring a carrier on creation.
// Options use the functional options pattern for flexible configuration.
type Option func(config *carrierConfig)

// WithTimeout sets the timeout for carrier operations.
//
// Parameters:
//
//	d - The timeout duration
//
// Returns:
//
//	An Option that configures the carrier's timeout
func WithTimeout(d time.Duration) Option {
	return func(c *carrierConfig) {
		c.timeout = d
	}
}

// WithMaxConcurrency sets the maximum number of concurrent operations for the carrier.
//
// Parameters:
//
//	n - The maximum number of concurrent operations
//
// Returns:
//
//	An Option that configures the carrier's concurrency
func WithMaxConcurrency(n int) Option {
	return func(c *carrierConfig) {
		c.maxConcurrency = n
	}
}

// WithCompletionCondition sets the completion condition for the carrier.
//
// Parameters:
//
//	cond - The completion condition function
//
// Returns:
//
//	An Option that configures the carrier's completion condition
func WithCompletionCondition(cond CompletionCondition) Option {
	return func(c *carrierConfig) {
		c.completionCondition = cond
	}
}

// CompletionCondition is a function type that defines when an event emitted by a carrier is considered as completed.
//
// By default, all carriers will consider only followup events (from the ones they sent).
// Both events passed as parameters to a CompletionCondition share the same Ref.
//
// Parameters:
//
//	sent - The event that was sent by the carrier
//	received - The event that was received
//
// Returns:
//
//	true if the received event completes the sent event, false otherwise
type CompletionCondition func(sent, received event.Event) bool

// CompletedOnFollowupReceived is a completion condition that returns true when the received event
// is a followup of the one sent (they both share the same Ref).
//
// Due to the current carriers implementation, this condition is always true
// (by construction: we already know that the received event is a followup of the one sent).
// This is the default completion condition function.
//
// Parameters:
//
//	sent - The event that was sent
//	received - The event that was received
//
// Returns:
//
//	Always true (the received event is a followup of the sent event)
func CompletedOnFollowupReceived(sent, received event.Event) bool {
	return true
}

// DoneFactory is a function type that creates a Done event from the event carrier and a list of received events.
type DoneFactory func(evtCarrier event.Event, received []event.Event) event.Event

// NoopDoneFactory is a DoneFactory that can be used when you don't want to emmit a specific done event.
func NoopDoneFactory(event.Event, []event.Event) event.Event {
	return nil
}
