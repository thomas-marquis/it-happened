package event

import "context"

type Type string

func (t Type) String() string {
	return string(t)
}

type Payload interface {
	Type() Type
}

type Event struct {
	ID      string
	Payload Payload
	Context context.Context
	Ref     string
}
