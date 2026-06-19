package event

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// Type represents the category or kind of an event.
// It is used to identify and classify different types of events in the system.
type Type string

// String returns the string representation of the Type.
func (t Type) String() string {
	return string(t)
}

// Payload is the interface that all event payloads must implement.
// A payload contains the data associated with an event and provides its type.
type Payload interface {
	// EventType returns the type of the event this payload represents.
	EventType() Type
}

// Event is the interface representing a domain event.
// Events are the fundamental building blocks of event-driven applications.
// An event is chainable, meaning it can be composed of other events to form a chain.
type Event interface {
	// ID returns the unique identifier of the event.
	ID() string
	// Type returns the type of the event.
	Type() Type
	// Payload returns the data payload of the event.
	Payload() Payload
	// Context returns the context associated with the event.
	Context() context.Context
	// ChainRef returns the unique reference identifier for the event chain.
	// All events in the same chain share the same ChainRef.
	ChainRef() string
	// ChainPosition returns the position of this event within its chain.
	// The first event in a chain has position 0, the next has position 1, etc.
	ChainPosition() uint
	// NewFollowup creates a new event that is a followup of this event.
	// The new event will share the same ChainRef and have an incremented position.
	NewFollowup(newPayload Payload, opts ...Option) Event
}

type impl struct {
	ctx context.Context

	id        string
	payload   Payload
	ref       string
	position  uint
	eventType Type
}

// Type returns the event type of the implementation.
func (e *impl) Type() Type {
	return e.eventType
}

// ChainRef returns the chain reference identifier.
// If this event is part of a chain, it shares this reference with other events in the same chain.
func (e *impl) ChainRef() string {
	return e.ref
}

// ChainPosition returns the position of this event within its chain.
// The first event in a chain has position 0.
func (e *impl) ChainPosition() uint {
	return e.position
}

// Payload returns the payload data of the event.
func (e *impl) Payload() Payload {
	return e.payload
}

// ID returns the unique identifier of the event.
func (e *impl) ID() string {
	return e.id
}

// Context returns the context associated with the event.
func (e *impl) Context() context.Context {
	return e.ctx
}

// MarshalJSON implements json.Marshaler for the event implementation.
// It marshals the event's payload to JSON.
func (e *impl) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.payload)
}

// NewFollowup creates a new event that is a followup of the current (parent) event.
//
// A followup event may be created from a previous parent event. Both are composing an event **chain**.
// A chain can be longer than two events.
// Each event in a chain shares the same ChainRef -- which is basically the ID of the first event in the chain.
// The position is incremented for each new event in the chain.
func (e *impl) NewFollowup(newPayload Payload, opts ...Option) Event {
	prevRef := e.ref
	if prevRef == "" {
		prevRef = uuid.New().String()
	}
	ne := newEventImpl(newPayload, opts...)
	ne.ref = prevRef
	ne.position = e.position + 1
	return ne
}

// New creates a new event with the given payload and options.
//
// The event will have a unique ID, and if no ChainRef is provided through options,
// the event's ID will be used as the ChainRef (starting a new chain).
//
// Parameters:
//
//	payload - The payload containing the event data
//	opts - Optional configuration options for the event
func New(payload Payload, opts ...Option) Event {
	return newEventImpl(payload, opts...)
}

func newEventImpl(payload Payload, opts ...Option) *impl {
	id := uuid.New().String()
	e := &impl{
		id:        id,
		payload:   payload,
		ctx:       context.Background(),
		eventType: payload.EventType(),
		ref:       id,
	}

	for _, opt := range opts {
		opt(e)
	}

	if e.ctx == nil {
		e.ctx = context.Background()
	}

	return e
}
