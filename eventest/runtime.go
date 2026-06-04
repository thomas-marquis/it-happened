package eventest

import (
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

const (
	DefaultPayloadType event.Type = "__eventest__.default"
)

type DefaultPayload int

func (DefaultPayload) Type() event.Type {
	return DefaultPayloadType
}

type marbleRuntime struct {
	clock    VirtualClock
	events   map[string]event.Event
	matchers map[string]event.Matcher
}

func (r *marbleRuntime) Run(ops []marble.Op) error {
	r.clock.Start()
	defer r.clock.Stop()

	//for i, op := range ops {
	//
	//}
	return nil
}
