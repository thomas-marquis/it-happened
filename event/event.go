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

type Event interface {
	Type() Type
	Payload() Payload
	ID() string
	ChainRef() string
	ChainPosition() uint
	Context() context.Context
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

func (e *impl) NewFollowup(newPayload Payload, opts ...Option) Event {
	prevRef := e.ref
	if prevRef == "" {
		prevRef = uuid.New().String()
	}
	ne := newEventImpl(newPayload, opts...)
	ne.ref = prevRef
	return ne
}

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
