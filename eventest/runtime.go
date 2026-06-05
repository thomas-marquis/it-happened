package eventest

import (
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

const (
	DefaultPayloadType event.Type = "__eventest__.default"
)

type DefaultPayload int

func (DefaultPayload) Type() event.Type {
	return DefaultPayloadType
}

type marbleRuntime struct {
	clock            VirtualClock
	events           map[string]event.Event
	matchers         map[string]event.Matcher
	baseTickDuration time.Duration
}
