package event

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type Type string

func (t Type) String() string {
	return string(t)
}

type Payload interface {
	EventType() Type
}

type Chainable interface {
	ChainRef() string
	ChainPosition() uint
}

type Event interface {
	ID() string
	Type() Type
	Payload() Payload
	Context() context.Context
}

type ChainableEvent interface {
	Event
	Chainable

	NewFollowup(newPayload Payload, opts ...Option) ChainableEvent
}

type impl struct {
	ctx context.Context

	id        string
	payload   Payload
	ref       string
	position  uint
	eventType Type
}

func (e *impl) Type() Type {
	return e.eventType
}

func (e *impl) ChainRef() string {
	return e.ref
}

func (e *impl) ChainPosition() uint {
	return e.position
}

func (e *impl) Payload() Payload {
	return e.payload
}

func (e *impl) ID() string {
	return e.id
}

func (e *impl) Context() context.Context {
	return e.ctx
}

func (e *impl) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.payload)
}

// NewFollowup creates a new event that is a followup of the current (parent) event.
//
// A followup event may be created from a previous parent event. Both are composing an event **chain**.
// A chain can be longer than two events.
// Each event in a chain shares the same ChainRef -- which is basically the ID of the first event in the chain.
// The position is incremented for each new event in the chain.
func (e *impl) NewFollowup(newPayload Payload, opts ...Option) ChainableEvent {
	prevRef := e.ref
	if prevRef == "" {
		prevRef = uuid.New().String()
	}
	ne := newEventImpl(newPayload, opts...)
	ne.ref = prevRef
	ne.position = e.position + 1
	return ne
}

func New(payload Payload, opts ...Option) ChainableEvent {
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
